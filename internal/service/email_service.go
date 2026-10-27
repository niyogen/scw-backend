package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"delivery-backend/internal/config"
	"delivery-backend/internal/models"
)

type EmailService interface {
	SendEmail(to string, subject string, htmlBody string) error
	SendBookingCreatedCustomer(booking *models.Booking)
	SendJobAlertWorker(booking *models.Booking, worker *models.Worker)
	SendBookingApprovedCustomer(booking *models.Booking, worker *models.Worker)
	SendBookingApprovedWorker(booking *models.Booking, worker *models.Worker)
	SendBookingRejectedCustomer(booking *models.Booking, worker *models.Worker)
	SendBookingStatusUpdate(booking *models.Booking, newStatus string)
	SendPasswordResetConfirmation(user *models.User)
}

type emailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{cfg: cfg}
}

// SendEmail sends a responsive HTML email via SMTP with automatic SSL/STARTTLS fallback
func (s *emailService) SendEmail(to string, subject string, htmlBody string) error {
	to = strings.TrimSpace(to)
	if to == "" {
		return nil
	}

	smtpUser := strings.TrimSpace(s.cfg.SMTPUser)
	smtpPass := strings.TrimSpace(s.cfg.SMTPPassword)
	smtpHost := strings.TrimSpace(s.cfg.SMTPHost)
	smtpPort := strings.TrimSpace(s.cfg.SMTPPort)

	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}
	if smtpUser == "" {
		smtpUser = "itranga@gmail.com"
	}

	fromName := s.cfg.SMTPFromName
	if fromName == "" {
		fromName = "Sewa Home Services"
	}
	fromEmail := s.cfg.SMTPFromEmail
	if fromEmail == "" {
		fromEmail = smtpUser
	}

	// If no password configured yet, log helpful instruction and return gracefully
	if smtpPass == "" {
		log.Printf("[EMAIL NOTICE] SMTP password not set in .env. Skipping actual dispatch of email to '%s' with subject '%s'. Set SMTP_PASSWORD in .env to enable.", to, subject)
		return nil
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, fromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + htmlBody)
	msgBytes := message.Bytes()

	// 1. Try Direct SSL on Port 465 (preferred for cloud / containers)
	var err error
	if smtpPort == "465" {
		err = sendMailSSL(smtpHost, "465", smtpUser, smtpPass, fromEmail, to, msgBytes)
	} else {
		// Try STARTTLS on 587 first, fallback to SSL 465 on failure
		err = sendMailSTARTTLS(smtpHost, smtpPort, smtpUser, smtpPass, fromEmail, to, msgBytes)
		if err != nil {
			log.Printf("[EMAIL RETRY] STARTTLS on port %s returned '%v'. Retrying with direct SSL on port 465...", smtpPort, err)
			err = sendMailSSL(smtpHost, "465", smtpUser, smtpPass, fromEmail, to, msgBytes)
		}
	}

	if err != nil {
		log.Printf("[EMAIL ERROR] Failed to send email to %s: %v", to, err)
		return err
	}

	log.Printf("[EMAIL SUCCESS] Sent '%s' to %s", subject, to)
	return nil
}

func sendMailSSL(host, port, user, pass, from, to string, msg []byte) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls.Dial error: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp.NewClient error: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth error: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from error: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to error: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data writer error: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write msg error: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close writer error: %w", err)
	}

	return client.Quit()
}

func sendMailSTARTTLS(host, port, user, pass, from, to string, msg []byte) error {
	addr := fmt.Sprintf("%s:%s", host, port)
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp.Dial error: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{
		ServerName: host,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starttls error: %w", err)
	}

	auth := smtp.PlainAuth("", user, pass, host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth error: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from error: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to error: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data writer error: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("write msg error: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close writer error: %w", err)
	}

	return client.Quit()
}

// 1. Send when customer creates a new service booking
func (s *emailService) SendBookingCreatedCustomer(booking *models.Booking) {
	if booking == nil || booking.User == nil || booking.User.Email == "" {
		return
	}

	go func() {
		customerEmail := booking.User.Email
		customerName := booking.User.Name
		if customerName == "" {
			customerName = "Valued Customer"
		}
		serviceTitle := "Home Service"
		if booking.Service != nil && booking.Service.Title != "" {
			serviceTitle = booking.Service.Title
		} else if booking.JobTitle != "" {
			serviceTitle = booking.JobTitle
		}

		subject := fmt.Sprintf("✅ Booking Confirmation - Order #%s (%s)", booking.BookingNumber, serviceTitle)
		statusBadge := `<span style="background-color: #FEF3C7; color: #B45309; padding: 4px 10px; border-radius: 12px; font-weight: 700; font-size: 12px;">⏳ Pending Specialist Acceptance</span>`
		if booking.Status == "confirmed" || booking.Status == "accepted" {
			statusBadge = `<span style="background-color: #DCFCE7; color: #15803D; padding: 4px 10px; border-radius: 12px; font-weight: 700; font-size: 12px;">✓ Confirmed</span>`
		}

		content := fmt.Sprintf(`
			<p style="font-size: 16px; color: #1e293b; margin-top: 0;">Dear <strong>%s</strong>,</p>
			<p style="color: #475569; font-size: 14px; line-height: 1.6;">
				Thank you for choosing <strong>Sewa</strong>. Your service booking request has been successfully placed on our platform and dispatched to our specialists network.
			</p>

			<div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 14px; padding: 18px; margin: 20px 0;">
				<table style="width: 100%%; border-collapse: collapse; font-size: 14px; color: #334155;">
					<tr>
						<td style="padding: 6px 0; color: #64748b; width: 40%%;">Order Number:</td>
						<td style="padding: 6px 0; font-weight: 700; color: #0f172a;">#%s</td>
					</tr>
					<tr>
						<td style="padding: 6px 0; color: #64748b;">Service:</td>
						<td style="padding: 6px 0; font-weight: 700; color: #0f172a;">%s</td>
					</tr>
					<tr>
						<td style="padding: 6px 0; color: #64748b;">Current Status:</td>
						<td style="padding: 6px 0;">%s</td>
					</tr>
					<tr>
						<td style="padding: 6px 0; color: #64748b;">Location:</td>
						<td style="padding: 6px 0; font-weight: 600; color: #0f172a;">%s</td>
					</tr>
					<tr>
						<td style="padding: 6px 0; color: #64748b;">Address Details:</td>
						<td style="padding: 6px 0; color: #475569;">%s</td>
					</tr>
					<tr>
						<td style="padding: 8px 0 0 0; color: #64748b; font-size: 15px;">Total Amount:</td>
						<td style="padding: 8px 0 0 0; font-size: 18px; font-weight: 800; color: #0f392b;">Rs. %.0f</td>
					</tr>
				</table>
			</div>

			<p style="color: #475569; font-size: 13.5px; line-height: 1.5;">
				Our service specialists will review your request and confirm acceptance. You can track this request in real-time in the <strong>My Jobs</strong> section on your mobile app.
			</p>
		`, customerName, booking.BookingNumber, serviceTitle, statusBadge, booking.LocationTitle, booking.LocationSubtitle, booking.FinalAmount)

		html := s.wrapResponsiveLayout("Service Booking Confirmation", content)
		_ = s.SendEmail(customerEmail, subject, html)
	}()
}

// 2. Send job request alert to assigned worker
func (s *emailService) SendJobAlertWorker(booking *models.Booking, worker *models.Worker) {
	if booking == nil || worker == nil || worker.Email == "" {
		return
	}

	go func() {
		workerEmail := worker.Email
		workerName := worker.Name
		serviceTitle := "Home Service"
		if booking.Service != nil && booking.Service.Title != "" {
			serviceTitle = booking.Service.Title
		} else if booking.JobTitle != "" {
			serviceTitle = booking.JobTitle
		}

		customerName := "Customer"
		if booking.User != nil && booking.User.Name != "" {
			customerName = booking.User.Name
		}

		subject := fmt.Sprintf("🔔 New Job Alert - Order #%s (%s)", booking.BookingNumber, serviceTitle)

		content := fmt.Sprintf(`
			<p style="font-size: 16px; color: #1e293b; margin-top: 0;">Hello <strong>%s</strong>,</p>
			<p style="color: #475569; font-size: 14px; line-height: 1.6;">
				You have a new incoming customer job request on the Sewa platform. Please open your <strong>Partner Dashboard</strong> to review and accept or decline.
			</p>

			<div style="background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 14px; padding: 18px; margin: 20px 0;">
				<table style="width: 100%%; border-collapse: collapse; font-size: 14px; color: #334155;">
					<tr>
						<td style="padding: 6px 0; color: #166534; width: 40%%;">Order ID:</td>
						<td style="padding: 6px 0; font-weight: 700; color: #0f172a;">#%s</td>
					</tr>
					<tr>
						<td style="padding: 6px 0; color: #166534;">Service:</td>
						<td style="padding: 6px 0; font-weight: 700; color: #0f172a;">%s</td>
					</tr>
					<tr>
						<td style="padding: 6px 0; color: #166534;">Customer:</td>
						<td style="padding: 6px 0; font-weight: 600; color: #0f172a;">%s</td>
					</tr>
					<tr>
						<td style="padding: 6px 0; color: #166534;">Location Area:</td>
						<td style="padding: 6px 0; font-weight: 600; color: #0f172a;">%s (%s)</td>
					</tr>
				</table>
			</div>

			<p style="color: #475569; font-size: 13.5px; line-height: 1.5;">
				⚡ Please respond promptly in your mobile app Partner Dispatch Radar to confirm this booking.
			</p>
		`, workerName, booking.BookingNumber, serviceTitle, customerName, booking.LocationTitle, booking.LocationSubtitle)

		html := s.wrapResponsiveLayout("New Service Dispatch Request", content)
		_ = s.SendEmail(workerEmail, subject, html)
	}()
}

// 3. Send to customer when worker approves / accepts the booking
func (s *emailService) SendBookingApprovedCustomer(booking *models.Booking, worker *models.Worker) {
	if booking == nil || booking.User == nil || booking.User.Email == "" {
		return
	}

	go func() {
		customerEmail := booking.User.Email
		customerName := booking.User.Name
		serviceTitle := "Home Service"
		if booking.Service != nil && booking.Service.Title != "" {
			serviceTitle = booking.Service.Title
		}

		workerName := "Specialist"
		workerPhone := ""
		if worker != nil {
			workerName = worker.Name
			workerPhone = worker.Phone
		}

		subject := fmt.Sprintf("🎉 Booking Confirmed! Specialist Assigned for Order #%s", booking.BookingNumber)

		content := fmt.Sprintf(`
			<p style="font-size: 16px; color: #1e293b; margin-top: 0;">Dear <strong>%s</strong>,</p>
			<p style="color: #475569; font-size: 14px; line-height: 1.6;">
				Great news! Service specialist <strong>%s</strong> has accepted your booking request for <strong>%s</strong>.
			</p>

			<div style="background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 14px; padding: 18px; margin: 20px 0;">
				<h4 style="margin: 0 0 10px 0; color: #1d4ed8; font-size: 15px;">Assigned Specialist Details</h4>
				<table style="width: 100%%; border-collapse: collapse; font-size: 14px; color: #334155;">
					<tr>
						<td style="padding: 5px 0; color: #64748b; width: 40%%;">Technician Name:</td>
						<td style="padding: 5px 0; font-weight: 700; color: #0f172a;">%s</td>
					</tr>
					<tr>
						<td style="padding: 5px 0; color: #64748b;">Direct Phone:</td>
						<td style="padding: 5px 0; font-weight: 700; color: #0284c7;">%s</td>
					</tr>
					<tr>
						<td style="padding: 5px 0; color: #64748b;">Order Number:</td>
						<td style="padding: 5px 0; font-weight: 600; color: #0f172a;">#%s</td>
					</tr>
					<tr>
						<td style="padding: 5px 0; color: #64748b;">Service Location:</td>
						<td style="padding: 5px 0; color: #475569;">%s</td>
					</tr>
					<tr>
						<td style="padding: 8px 0 0 0; color: #64748b;">Total Amount:</td>
						<td style="padding: 8px 0 0 0; font-size: 17px; font-weight: 800; color: #0f392b;">Rs. %.0f</td>
					</tr>
				</table>
			</div>

			<p style="color: #475569; font-size: 13.5px; line-height: 1.5;">
				You can contact your specialist directly via Phone or WhatsApp from the <strong>My Jobs</strong> screen in your Sewa mobile app.
			</p>
		`, customerName, workerName, serviceTitle, workerName, workerPhone, booking.BookingNumber, booking.LocationTitle, booking.FinalAmount)

		html := s.wrapResponsiveLayout("Specialist Confirmed", content)
		_ = s.SendEmail(customerEmail, subject, html)
	}()
}

// 4. Send booking accepted confirmation to the worker
func (s *emailService) SendBookingApprovedWorker(booking *models.Booking, worker *models.Worker) {
	if booking == nil || worker == nil || worker.Email == "" {
		return
	}

	go func() {
		workerEmail := worker.Email
		workerName := worker.Name
		serviceTitle := "Home Service"
		if booking.Service != nil && booking.Service.Title != "" {
			serviceTitle = booking.Service.Title
		}

		customerName := "Customer"
		customerPhone := ""
		if booking.User != nil {
			customerName = booking.User.Name
			customerPhone = booking.User.Phone
		}

		subject := fmt.Sprintf("✓ Job Accepted Confirmation - Order #%s (%s)", booking.BookingNumber, serviceTitle)

		content := fmt.Sprintf(`
			<p style="font-size: 16px; color: #1e293b; margin-top: 0;">Hello <strong>%s</strong>,</p>
			<p style="color: #475569; font-size: 14px; line-height: 1.6;">
				You have confirmed acceptance of Order <strong>#%s</strong>. Here are the job details:
			</p>

			<div style="background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 14px; padding: 18px; margin: 20px 0;">
				<table style="width: 100%%; border-collapse: collapse; font-size: 14px; color: #334155;">
					<tr>
						<td style="padding: 5px 0; color: #64748b; width: 40%%;">Customer Name:</td>
						<td style="padding: 5px 0; font-weight: 700; color: #0f172a;">%s</td>
					</tr>
					<tr>
						<td style="padding: 5px 0; color: #64748b;">Customer Phone:</td>
						<td style="padding: 5px 0; font-weight: 700; color: #0284c7;">%s</td>
					</tr>
					<tr>
						<td style="padding: 5px 0; color: #64748b;">Address:</td>
						<td style="padding: 5px 0; font-weight: 600; color: #0f172a;">%s, %s</td>
					</tr>
					<tr>
						<td style="padding: 8px 0 0 0; color: #64748b;">Your Payout:</td>
						<td style="padding: 8px 0 0 0; font-size: 17px; font-weight: 800; color: #15803d;">Rs. %.0f</td>
					</tr>
				</table>
			</div>

			<p style="color: #475569; font-size: 13.5px; line-height: 1.5;">
				When you arrive at the location, tap <strong>Start Work</strong> in your partner app. Once finished, tap <strong>Mark as Completed</strong>.
			</p>
		`, workerName, booking.BookingNumber, customerName, customerPhone, booking.LocationTitle, booking.LocationSubtitle, booking.FinalAmount)

		html := s.wrapResponsiveLayout("Job Accepted Confirmation", content)
		_ = s.SendEmail(workerEmail, subject, html)
	}()
}

// 5. Send to customer when a worker declines/rejects a booking
func (s *emailService) SendBookingRejectedCustomer(booking *models.Booking, worker *models.Worker) {
	if booking == nil || booking.User == nil || booking.User.Email == "" {
		return
	}

	go func() {
		customerEmail := booking.User.Email
		customerName := booking.User.Name
		serviceTitle := "Home Service"
		if booking.Service != nil && booking.Service.Title != "" {
			serviceTitle = booking.Service.Title
		}

		subject := fmt.Sprintf("ℹ️ Service Request Update - Order #%s (%s)", booking.BookingNumber, serviceTitle)

		content := fmt.Sprintf(`
			<p style="font-size: 16px; color: #1e293b; margin-top: 0;">Dear <strong>%s</strong>,</p>
			<p style="color: #475569; font-size: 14px; line-height: 1.6;">
				The specialist initially selected was unavailable for order <strong>#%s</strong>.
			</p>
			<div style="background: #fef2f2; border: 1px solid #fecaca; border-radius: 14px; padding: 16px; margin: 18px 0; color: #991b1b; font-size: 13.5px;">
				Your request has been automatically placed on our high-priority dispatch radar to connect with other qualified specialists in your area.
			</div>
			<p style="color: #475569; font-size: 13.5px; line-height: 1.5;">
				You can monitor the status or cancel your request at any time from the <strong>My Jobs</strong> screen in your app.
			</p>
		`, customerName, booking.BookingNumber)

		html := s.wrapResponsiveLayout("Service Update", content)
		_ = s.SendEmail(customerEmail, subject, html)
	}()
}

// 6. Send general status update (in_progress, completed, cancelled)
func (s *emailService) SendBookingStatusUpdate(booking *models.Booking, newStatus string) {
	if booking == nil || booking.User == nil || booking.User.Email == "" {
		return
	}

	go func() {
		customerEmail := booking.User.Email
		customerName := booking.User.Name
		serviceTitle := "Home Service"
		if booking.Service != nil && booking.Service.Title != "" {
			serviceTitle = booking.Service.Title
		}

		var subject, title, body string
		switch newStatus {
		case "in_progress":
			subject = fmt.Sprintf("🚀 Specialist Started Work - Order #%s", booking.BookingNumber)
			title = "Work in Progress"
			body = fmt.Sprintf(`
				<p style="color: #475569; font-size: 14px; line-height: 1.6;">
					Your service specialist has arrived and marked Order <strong>#%s</strong> (%s) as <strong>In Progress</strong>.
				</p>
			`, booking.BookingNumber, serviceTitle)
		case "completed":
			subject = fmt.Sprintf("🎉 Service Completed & Receipt - Order #%s", booking.BookingNumber)
			title = "Service Completed"
			body = fmt.Sprintf(`
				<p style="color: #475569; font-size: 14px; line-height: 1.6;">
					Your service for <strong>%s</strong> (Order #%s) has been marked as <strong>Completed</strong>.
				</p>
				<div style="background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 14px; padding: 16px; margin: 16px 0;">
					<div style="font-size: 13px; color: #166534;">Final Amount Settled:</div>
					<div style="font-size: 20px; font-weight: 800; color: #15803d; margin-top: 4px;">Rs. %.0f</div>
					<div style="font-size: 12px; color: #15803d; margin-top: 6px;">✓ 7-Day Workmanship Warranty Active</div>
				</div>
				<p style="color: #475569; font-size: 13.5px; line-height: 1.5;">
					Please take a moment in the Sewa mobile app to rate your experience and review your specialist.
				</p>
			`, serviceTitle, booking.BookingNumber, booking.FinalAmount)
		case "cancelled":
			subject = fmt.Sprintf("✕ Order Cancelled - Order #%s", booking.BookingNumber)
			title = "Order Cancelled"
			body = fmt.Sprintf(`
				<p style="color: #475569; font-size: 14px; line-height: 1.6;">
					Order <strong>#%s</strong> for %s has been cancelled.
				</p>
			`, booking.BookingNumber, serviceTitle)
		default:
			return
		}

		content := fmt.Sprintf(`
			<p style="font-size: 16px; color: #1e293b; margin-top: 0;">Dear <strong>%s</strong>,</p>
			%s
		`, customerName, body)

		html := s.wrapResponsiveLayout(title, content)
		_ = s.SendEmail(customerEmail, subject, html)
	}()
}

// 7. Send password reset confirmation email
func (s *emailService) SendPasswordResetConfirmation(user *models.User) {
	if user == nil || user.Email == "" {
		return
	}

	go func() {
		userEmail := user.Email
		userName := user.Name
		if userName == "" {
			userName = "User"
		}

		subject := "🔒 Security Alert: Your Sewa Password Was Reset"

		content := fmt.Sprintf(`
			<p style="font-size: 16px; color: #1e293b; margin-top: 0;">Hello <strong>%s</strong>,</p>
			<p style="color: #475569; font-size: 14px; line-height: 1.6;">
				This email confirms that the password for your Sewa account (<strong>%s</strong>) was recently changed successfully.
			</p>
			<div style="background: #eff6ff; border: 1px solid #bfdbfe; border-radius: 12px; padding: 14px; margin: 18px 0; font-size: 13px; color: #1d4ed8;">
				If you initiated this change, you can safely disregard this email.
			</div>
			<p style="color: #b91c1c; font-size: 13px; line-height: 1.5;">
				If you did <strong>not</strong> make this change, please contact our support desk immediately at <strong>support@sewa.lk</strong> or call +94 11 234 5678.
			</p>
		`, userName, userEmail)

		html := s.wrapResponsiveLayout("Password Reset Confirmation", content)
		_ = s.SendEmail(userEmail, subject, html)
	}()
}

// wrapResponsiveLayout provides an email template matching Sewa's brand aesthetic
func (s *emailService) wrapResponsiveLayout(headerTitle, contentHtml string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <style>
    body {
      margin: 0;
      padding: 0;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      background-color: #f1f5f9;
      color: #334155;
      -webkit-text-size-adjust: 100%%;
    }
    .wrapper {
      width: 100%%;
      background-color: #f1f5f9;
      padding: 30px 10px;
    }
    .container {
      max-width: 580px;
      margin: 0 auto;
      background-color: #ffffff;
      border-radius: 20px;
      overflow: hidden;
      border: 1px solid #e2e8f0;
      box-shadow: 0 4px 16px rgba(0,0,0,0.06);
    }
    .header {
      background: linear-gradient(135deg, #0f392b 0%%, #0f172a 100%%);
      padding: 28px 24px;
      text-align: center;
      color: #ffffff;
    }
    .header h1 {
      margin: 0;
      font-size: 24px;
      font-weight: 800;
      letter-spacing: -0.5px;
      color: #ffffff;
    }
    .header .subtitle {
      font-size: 12px;
      color: #c5a059;
      font-weight: 700;
      letter-spacing: 1px;
      text-transform: uppercase;
      margin-top: 4px;
    }
    .body {
      padding: 28px 24px;
    }
    .footer {
      background-color: #f8fafc;
      padding: 20px 24px;
      text-align: center;
      font-size: 12px;
      color: #94a3b8;
      border-top: 1px solid #e2e8f0;
      line-height: 1.5;
    }
    .footer a {
      color: #0f392b;
      text-decoration: underline;
    }
  </style>
</head>
<body>
  <div class="wrapper">
    <div class="container">
      <div class="header">
        <div class="subtitle">Sewa Services • Sri Lanka</div>
        <h1>%s</h1>
      </div>
      <div class="body">
        %s
      </div>
      <div class="footer">
        <p style="margin: 0 0 6px 0;"><strong>Sewa Technologies (Pvt) Ltd</strong></p>
        <p style="margin: 0 0 6px 0;">Colombo, Sri Lanka • Hotline: +94 11 234 5678</p>
        <p style="margin: 0;">Need help? Email <a href="mailto:support@sewa.lk">support@sewa.lk</a> or message us on WhatsApp.</p>
      </div>
    </div>
  </div>
</body>
</html>
`, headerTitle, headerTitle, contentHtml)
}
