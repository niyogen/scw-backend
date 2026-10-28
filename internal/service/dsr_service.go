package service

import (
	"errors"
	"fmt"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErasureGraceDays is the window during which a user can cancel their
// own erasure request before hard-delete processing begins. Matches
// USE_CASES.md UC-09 and the Privacy Policy's §8 deletion SLA.
const ErasureGraceDays = 7

// DSRService satisfies PDPA §20-22 data-subject-rights for the two
// biggest items: right of access (export) and right to erasure.
// Rectification is handled by the existing Edit Profile flow.
// Consent withdrawal (which effectively enforces restriction) is
// Task #12. Automated-decision contest is deferred until we actually
// have automated decisions in production.
type DSRService interface {
	// Export returns every row the user appears in across the schema,
	// serialized as a DSRExportResponse. The caller streams this to
	// the user (JSON download or an emailed GCS signed URL).
	Export(userID uuid.UUID) (*dto.DSRExportResponse, error)

	// RequestErasure flips the user to is_active=false and sets
	// ScheduledDeletionAt = now + ErasureGraceDays. During the window
	// the user can CancelErasure. After the window a background job
	// (not yet shipped — v1 is manual admin) hard-deletes related rows
	// and anonymises reviews. Consent records are preserved.
	RequestErasure(userID uuid.UUID) (*dto.DSRErasureResponse, error)

	// CancelErasure clears the erasure flags if we're still within the
	// grace window. Returns an error if the window has already closed
	// or no erasure was pending.
	CancelErasure(userID uuid.UUID) (*dto.DSRErasureCancelResponse, error)
}

type dsrService struct {
	db             *gorm.DB
	userRepo       repository.UserRepository
	bookingRepo    repository.BookingRepository
	locationRepo   repository.LocationRepository
	workerRepo     repository.WorkerRepository
	consentRepo    repository.ConsentRepository
	consentService ConsentService
	emailSvc       EmailService
}

func NewDSRService(
	db *gorm.DB,
	userRepo repository.UserRepository,
	bookingRepo repository.BookingRepository,
	locationRepo repository.LocationRepository,
	workerRepo repository.WorkerRepository,
	consentRepo repository.ConsentRepository,
	consentService ConsentService,
	emailSvc EmailService,
) DSRService {
	return &dsrService{
		db:             db,
		userRepo:       userRepo,
		bookingRepo:    bookingRepo,
		locationRepo:   locationRepo,
		workerRepo:     workerRepo,
		consentRepo:    consentRepo,
		consentService: consentService,
		emailSvc:       emailSvc,
	}
}

func (s *dsrService) Export(userID uuid.UUID) (*dto.DSRExportResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Account
	accountDTO := mapUserToDTO(user)

	// Bookings
	bookings, _ := s.bookingRepo.FindByUserID(userID)
	bookingDTOs := make([]dto.DSRExportBooking, 0, len(bookings))
	for _, b := range bookings {
		bookingDTOs = append(bookingDTOs, dto.DSRExportBooking{
			BookingID:        b.ID,
			BookingNumber:    b.BookingNumber,
			ServiceID:        b.ServiceID,
			WorkerID:         b.WorkerID,
			LocationTitle:    b.LocationTitle,
			LocationSubtitle: b.LocationSubtitle,
			Status:           b.Status,
			JobTitle:         b.JobTitle,
			JobDescription:   b.JobDescription,
			JobPhotos:        b.JobPhotos,
			ScheduledAt:      b.ScheduledAt,
			CreatedAt:        b.CreatedAt,
			UpdatedAt:        b.UpdatedAt,
		})
	}

	// Saved addresses
	addresses, _ := s.locationRepo.FindByUserID(userID)
	addressDTOs := make([]dto.DSRExportAddress, 0, len(addresses))
	for _, a := range addresses {
		addressDTOs = append(addressDTOs, dto.DSRExportAddress{
			AddressID: a.ID,
			Title:     a.Title,
			Subtitle:  a.Subtitle,
			Latitude:  a.Latitude,
			Longitude: a.Longitude,
			IsDefault: a.IsDefault,
			CreatedAt: a.CreatedAt,
		})
	}

	// Reviews given (if the user is also a worker, these are the ones
	// users wrote about them — those are already "theirs" by GDPR
	// definition since the review subject can access the review).
	reviewDTOs := make([]dto.DSRExportReview, 0)
	if s.workerRepo != nil {
		if worker, err := s.workerRepo.FindByUserID(userID); err == nil && worker != nil {
			for _, r := range worker.Reviews {
				reviewDTOs = append(reviewDTOs, dto.DSRExportReview{
					ReviewID:  r.ID,
					WorkerID:  r.WorkerID,
					BookingID: r.BookingID,
					Rating:    r.Rating,
					Comment:   r.Comment,
					CreatedAt: r.CreatedAt,
				})
			}
		}
	}

	// Consent history — reuse the existing ConsentService.History
	consentHistory, _ := s.consentService.History(userID)
	consentDTOs := make([]dto.ConsentRecordResponse, 0)
	if consentHistory != nil {
		consentDTOs = consentHistory.Records
	}

	// Media — walk booking job photos + the account avatar
	mediaRefs := make([]dto.DSRExportMediaReference, 0)
	if user.AvatarURL != "" {
		mediaRefs = append(mediaRefs, dto.DSRExportMediaReference{
			Context: "avatar",
			URL:     user.AvatarURL,
		})
	}
	for _, b := range bookings {
		for _, url := range b.JobPhotos {
			mediaRefs = append(mediaRefs, dto.DSRExportMediaReference{
				Context: fmt.Sprintf("job_photo:%s", b.BookingNumber),
				URL:     url,
			})
		}
	}

	return &dto.DSRExportResponse{
		ExportedAt: time.Now().UTC(),
		Subject: dto.DSRExportSubject{
			UserID:        userID,
			SubjectIDHash: repository.HashSubjectID(userID),
			ExportVersion: "1.0.0",
		},
		Account:   &accountDTO,
		Bookings:  bookingDTOs,
		Addresses: addressDTOs,
		Reviews:   reviewDTOs,
		Consents:  consentDTOs,
		Media:     mediaRefs,
	}, nil
}

func (s *dsrService) RequestErasure(userID uuid.UUID) (*dto.DSRErasureResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Idempotent: if erasure is already pending for this user, return
	// the existing schedule instead of resetting the clock.
	if user.ErasureRequestedAt != nil && user.ScheduledDeletionAt != nil &&
		time.Now().UTC().Before(*user.ScheduledDeletionAt) {
		return &dto.DSRErasureResponse{
			Status:              "scheduled",
			RequestedAt:         *user.ErasureRequestedAt,
			ScheduledDeletionAt: *user.ScheduledDeletionAt,
			GraceDays:           ErasureGraceDays,
			CancelableUntil:     *user.ScheduledDeletionAt,
			Message:             fmt.Sprintf("Account deletion is already scheduled for %s. You can cancel until that date in Settings → Privacy.", user.ScheduledDeletionAt.Format("2006-01-02 15:04 UTC")),
		}, nil
	}

	now := time.Now().UTC()
	scheduled := now.AddDate(0, 0, ErasureGraceDays)

	user.ErasureRequestedAt = &now
	user.ScheduledDeletionAt = &scheduled
	user.IsActive = false
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to schedule erasure: %w", err)
	}

	// Email the user acknowledging the request (best-effort, non-blocking).
	if s.emailSvc != nil && user.Email != "" {
		go func(u models.User, schedule time.Time) {
			subject := "Smart Sewa — account deletion scheduled"
			htmlBody := fmt.Sprintf(`
				<p style="color:#475569;font-size:14px;line-height:1.6;">
					Hi <strong>%s</strong>, this confirms your request to delete your Smart Sewa account.
				</p>
				<div style="background:#fef2f2;border:1px solid #fecaca;border-radius:12px;padding:16px;margin:16px 0;">
					<div style="font-size:13px;color:#991b1b;">Scheduled deletion</div>
					<div style="font-size:18px;font-weight:800;color:#991b1b;margin-top:4px;">%s</div>
				</div>
				<p style="color:#475569;font-size:13.5px;line-height:1.5;">
					You have a %d-day grace window. To cancel, sign in and go to <strong>Settings → Privacy</strong>, then tap "Cancel account deletion". After the grace window, your personal data will be deleted and reviews you wrote will be anonymised. Consent records are preserved indefinitely as required by law.
				</p>
			`, u.Name, schedule.Format("Mon, 2 Jan 2006"), ErasureGraceDays)
			_ = s.emailSvc.SendEmail(u.Email, subject, htmlBody)
		}(*user, scheduled)
	}

	return &dto.DSRErasureResponse{
		Status:              "scheduled",
		RequestedAt:         now,
		ScheduledDeletionAt: scheduled,
		GraceDays:           ErasureGraceDays,
		CancelableUntil:     scheduled,
		Message:             fmt.Sprintf("Your account has been scheduled for deletion in %d days. Cancel any time before then in Settings → Privacy.", ErasureGraceDays),
	}, nil
}

func (s *dsrService) CancelErasure(userID uuid.UUID) (*dto.DSRErasureCancelResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.ErasureRequestedAt == nil || user.ScheduledDeletionAt == nil {
		return nil, errors.New("no erasure request is pending")
	}
	if time.Now().UTC().After(*user.ScheduledDeletionAt) {
		return nil, errors.New("the grace window has already closed")
	}

	user.ErasureRequestedAt = nil
	user.ScheduledDeletionAt = nil
	user.IsActive = true
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to cancel erasure: %w", err)
	}

	return &dto.DSRErasureCancelResponse{
		Status:      "cancelled",
		CancelledAt: time.Now().UTC(),
		Message:     "Your deletion request has been cancelled. Your account is active again.",
	}, nil
}
