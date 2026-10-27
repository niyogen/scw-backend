package main

import (
	"fmt"
	"log"

	"delivery-backend/internal/config"
	"delivery-backend/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Testing email delivery with settings:\n")
	fmt.Printf("SMTP Host: %s\n", cfg.SMTPHost)
	fmt.Printf("SMTP Port: %s\n", cfg.SMTPPort)
	fmt.Printf("SMTP User: %s\n", cfg.SMTPUser)
	fmt.Printf("SMTP From: %s <%s>\n", cfg.SMTPFromName, cfg.SMTPFromEmail)
	fmt.Printf("Target: itranga@gmail.com\n\n")

	emailSvc := service.NewEmailService(cfg)

	subject := "Test Email from Carfenter / Sewa Services Backend"
	htmlBody := `
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: Arial, sans-serif; background-color: #f4f6f8; margin: 0; padding: 20px;">
  <div style="max-width: 600px; margin: auto; background: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.05); padding: 30px;">
    <h2 style="color: #2b50aa; margin-top: 0;">Email Verification Test</h2>
    <p style="color: #333333; font-size: 15px; line-height: 1.5;">Hello,</p>
    <p style="color: #333333; font-size: 15px; line-height: 1.5;">This is a test email sent from the <strong>Carfenter / Sewa Services Backend</strong> to confirm that Gmail SMTP integration is working successfully.</p>
    <div style="background-color: #e8f4fd; border-left: 4px solid #2b50aa; padding: 12px 16px; margin: 20px 0;">
      <p style="margin: 0; color: #1e3a8a; font-size: 14px;"><strong>Status:</strong> Email dispatch verified!</p>
      <p style="margin: 4px 0 0 0; color: #1e3a8a; font-size: 14px;"><strong>Sender:</strong> ` + cfg.SMTPFromEmail + `</p>
      <p style="margin: 4px 0 0 0; color: #1e3a8a; font-size: 14px;"><strong>Recipient:</strong> itranga@gmail.com</p>
    </div>
    <p style="color: #777777; font-size: 13px;">If you received this message, the backend email notification system is fully operational.</p>
  </div>
</body>
</html>`

	err = emailSvc.SendEmail("itranga@gmail.com", subject, htmlBody)
	if err != nil {
		log.Fatalf("FAILED to send email: %v", err)
	}

	fmt.Println("\nSUCCESS: Email was sent successfully to itranga@gmail.com!")
}
