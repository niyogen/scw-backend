package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"delivery-backend/internal/config"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/utils"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// WorkerOnboardingService wraps the applicant self-service flow + the
// admin approve/reject/ban/unban actions. Kept separate from the
// discovery-oriented WorkerService so neither surface grows unwieldy.
type WorkerOnboardingService interface {
	// Self-service
	RegisterSelfService(req *dto.RegisterWorkerRequest) (*dto.RegisterWorkerResponse, error)
	SignKYCUploadURL(workerID uuid.UUID, req *dto.KYCUploadURLRequest) (*dto.KYCUploadURLResponse, error)
	UpdateKYC(workerID uuid.UUID, req *dto.UpdateKYCRequest) (*dto.WorkerResponse, error)
	GetMyStatus(userID uuid.UUID) (*dto.WorkerStatusResponse, error)
	UpdateMyApplication(userID uuid.UUID, req *dto.UpdateMyWorkerRequest) (*dto.WorkerResponse, error)

	// Admin
	ListForAdmin(status string, page, pageSize int) (*AdminWorkerListResult, error)
	GetForAdmin(workerID uuid.UUID) (*dto.WorkerResponse, error)
	Approve(workerID, adminID uuid.UUID) (*dto.WorkerResponse, error)
	Reject(workerID uuid.UUID, reason string) (*dto.WorkerResponse, error)
	Ban(workerID, adminID uuid.UUID, reason string) (*dto.WorkerResponse, error)
	Unban(workerID uuid.UUID) (*dto.WorkerResponse, error)
}

// AdminWorkerListResult is the service-side envelope for the paginated
// admin list. The handler maps it to utils.PaginatedResponse.
type AdminWorkerListResult struct {
	Items      []dto.AdminWorkerListItem
	TotalCount int64
	Page       int
	PageSize   int
}

type workerOnboardingService struct {
	workerRepo repository.WorkerRepository
	userRepo   repository.UserRepository
	cfg        *config.Config
	gcsClient  *storage.Client // may be nil in tests / local dev — we fall back to a signed-URL stub
}

// NewWorkerOnboardingService wires the dependencies. gcsClient may be
// nil; when it is, signing degrades to a stub URL so local dev + CI
// tests don't need GCS credentials.
func NewWorkerOnboardingService(
	workerRepo repository.WorkerRepository,
	userRepo repository.UserRepository,
	cfg *config.Config,
	gcsClient *storage.Client,
) WorkerOnboardingService {
	return &workerOnboardingService{
		workerRepo: workerRepo,
		userRepo:   userRepo,
		cfg:        cfg,
		gcsClient:  gcsClient,
	}
}

// --- validation ---

// nicOldFormat and nicNewFormat per Sri Lanka Department for Registration
// of Persons: old = 9 digits + V/X; new = 12 digits.
var (
	nicOldFormat = regexp.MustCompile(`^\d{9}[VvXx]$`)
	nicNewFormat = regexp.MustCompile(`^\d{12}$`)
)

// ValidateNIC returns nil if the input matches either Sri Lankan NIC
// format. Exported so handlers / tests can reuse it. Empty string is
// treated as "not provided" — callers decide whether that's an error.
func ValidateNIC(nic string) error {
	nic = strings.TrimSpace(nic)
	if nic == "" {
		return errors.New("nic_number is required")
	}
	if nicOldFormat.MatchString(nic) || nicNewFormat.MatchString(nic) {
		return nil
	}
	return errors.New("nic_number must match the Sri Lankan NIC format (either 9 digits + V/X, or 12 digits)")
}

// --- self-service ---

func (s *workerOnboardingService) RegisterSelfService(req *dto.RegisterWorkerRequest) (*dto.RegisterWorkerResponse, error) {
	// 1. Validate age gate. We treat age=0 as "not provided" and reject
	//    (the applicant must declare they're 18+).
	if req.Age < 18 {
		return nil, errors.New("age must be at least 18")
	}

	// 2. NIC format check. Only enforced when provided — the legacy
	//    Vercel form sometimes POSTs without NIC and we don't want to
	//    break it; the admin UI will nag the applicant to add one
	//    before approval.
	if strings.TrimSpace(req.NICNumber) != "" {
		if err := ValidateNIC(req.NICNumber); err != nil {
			return nil, err
		}
	}

	if strings.TrimSpace(req.Email) == "" {
		return nil, errors.New("email is required")
	}
	if strings.TrimSpace(req.Phone) == "" {
		return nil, errors.New("phone is required")
	}

	// 3. Find or create the User row for this applicant. Workers log back
	//    in via /auth/login once an admin approves them; the account is
	//    kept inactive until then so they can't book as a customer either.
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	user, _ := s.userRepo.FindByEmail(email)
	if user == nil {
		// Also check phone uniqueness — users table has unique index on it.
		if existing, _ := s.userRepo.FindByPhone(phone); existing != nil {
			return nil, errors.New("phone number is already registered to another account")
		}
		// Random password; applicant uses /auth/forgot-password to set
		// their own via email. Keeps the self-service surface tiny.
		randPass := uuid.New().String()
		hashedPassword, err := utils.HashPassword(randPass)
		if err != nil {
			return nil, errors.New("failed to secure account")
		}
		avatarURL := req.AvatarURL
		if avatarURL == "" {
			avatarURL = "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&auto=format&fit=crop&q=80"
		}
		newUser := models.User{
			Name:         req.Name,
			Email:        email,
			Phone:        phone,
			PasswordHash: hashedPassword,
			AvatarURL:    avatarURL,
			Role:         "worker",
			IsActive:     false, // flipped to true on admin approval
		}
		if err := s.userRepo.Create(&newUser); err != nil {
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
		user = &newUser
	} else {
		// Existing user — promote role to worker but don't force
		// is_active=false if they were already an active customer.
		if user.Role == "customer" {
			user.Role = "worker"
			_ = s.userRepo.Update(user)
		}
	}

	// 4. If a worker row already exists for this user, treat the submit
	//    as an edit (idempotent) when it's still pending — don't 500 on
	//    double-click. If it's already approved/rejected/banned, reject
	//    the resubmit.
	if existing, _ := s.workerRepo.FindByUserID(user.ID); existing != nil {
		switch existing.Status {
		case "pending", "":
			// Resubmit → merge the fields and keep status=pending.
			applyPendingEdits(existing, req)
			if err := s.workerRepo.Update(existing); err != nil {
				return nil, err
			}
			token, _ := utils.GenerateJWT(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
			return &dto.RegisterWorkerResponse{
				WorkerID: existing.ID,
				UserID:   user.ID,
				Status:   existing.Status,
				Token:    token,
			}, nil
		default:
			return nil, fmt.Errorf("worker application already %s for this account", existing.Status)
		}
	}

	// 5. Create the pending worker row. Defaults mirror the Vercel form's
	//    client-side derivations so the data shape stays consistent.
	workerCode := fmt.Sprintf("wrk_%s", user.ID.String()[:8])
	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = user.AvatarURL
	}
	rateValue := req.RateValue
	if rateValue == 0 {
		rateValue = 1800.0
	}
	hourlyRate := req.HourlyRate
	if hourlyRate == "" {
		hourlyRate = fmt.Sprintf("Rs. %.0f / hr", rateValue)
	}
	specializations := req.Specializations
	if specializations == nil {
		if len(req.ServiceTypes) > 0 {
			specializations = []string{req.ServiceTypes[0], req.City + " Area"}
		} else {
			specializations = []string{"General Services"}
		}
	}
	badges := req.Badges
	if badges == nil {
		badges = []string{"Partner Applicant", "NIC Submitted", req.City}
	}
	serviceAreas := req.ServiceAreas
	if serviceAreas == nil && req.City != "" {
		serviceAreas = []string{req.City}
	}

	worker := models.Worker{
		UserID:           &user.ID,
		WorkerCode:       workerCode,
		Name:             req.Name,
		AvatarURL:        avatarURL,
		Age:              req.Age,
		ExperienceYears:  req.ExperienceYears,
		Rating:           5.0,
		ReviewsCount:     0,
		ServiceTypes:     req.ServiceTypes,
		Specializations:  specializations,
		Phone:            phone,
		Email:            email,
		Bio:              req.Bio,
		CompletedJobs:    0,
		HourlyRate:       hourlyRate,
		RateValue:        rateValue,
		Latitude:         6.9271,
		Longitude:        79.8612,
		OperatingAddress: req.City,
		IsOpenForWork:    false, // off until approved
		RadiusKm:         10.0,
		IsVerified:       false,
		IsActive:         false, // not visible on /workers until approved
		IsAvailableToday: false,
		Badges:           badges,
		Status:           "pending",
		NICNumber:        strings.TrimSpace(req.NICNumber),
		City:             req.City,
		ServiceAreas:     serviceAreas,
	}
	if err := s.workerRepo.Create(&worker); err != nil {
		return nil, fmt.Errorf("failed to submit application: %w", err)
	}

	token, _ := utils.GenerateJWT(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	return &dto.RegisterWorkerResponse{
		WorkerID: worker.ID,
		UserID:   user.ID,
		Status:   worker.Status,
		Token:    token,
	}, nil
}

// applyPendingEdits merges a RegisterWorkerRequest onto an existing
// pending worker row. Only non-zero fields overwrite; empty values are
// treated as "unchanged" so partial resubmits don't blow away data.
func applyPendingEdits(w *models.Worker, req *dto.RegisterWorkerRequest) {
	if req.Name != "" {
		w.Name = req.Name
	}
	if req.AvatarURL != "" {
		w.AvatarURL = req.AvatarURL
	}
	if req.Age > 0 {
		w.Age = req.Age
	}
	if req.ExperienceYears > 0 {
		w.ExperienceYears = req.ExperienceYears
	}
	if len(req.ServiceTypes) > 0 {
		w.ServiceTypes = req.ServiceTypes
	}
	if len(req.Specializations) > 0 {
		w.Specializations = req.Specializations
	}
	if req.Phone != "" {
		w.Phone = req.Phone
	}
	if req.Bio != "" {
		w.Bio = req.Bio
	}
	if req.HourlyRate != "" {
		w.HourlyRate = req.HourlyRate
	}
	if req.RateValue > 0 {
		w.RateValue = req.RateValue
	}
	if req.NICNumber != "" {
		w.NICNumber = strings.TrimSpace(req.NICNumber)
	}
	if req.City != "" {
		w.City = req.City
		w.OperatingAddress = req.City
	}
	if len(req.ServiceAreas) > 0 {
		w.ServiceAreas = req.ServiceAreas
	}
}

// --- KYC ---

// SignKYCUploadURL returns a V4 signed URL that the client can PUT the
// KYC file to directly. The URL is scoped to one object and expires in
// 15 minutes. If the GCS client isn't configured (local dev), returns
// a stub URL that still roundtrips through the backend's /upload/photo
// pathway — that way the mobile/web app can be developed without
// touching production storage.
func (s *workerOnboardingService) SignKYCUploadURL(workerID uuid.UUID, req *dto.KYCUploadURLRequest) (*dto.KYCUploadURLResponse, error) {
	validTypes := map[string]bool{"nic_front": true, "nic_back": true, "selfie": true, "trade_cert": true}
	if !validTypes[req.DocumentType] {
		return nil, errors.New("document_type must be one of: nic_front, nic_back, selfie, trade_cert")
	}

	// Confirm worker exists.
	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil || worker == nil {
		return nil, errors.New("worker not found")
	}

	ext := filepath.Ext(req.Filename)
	if ext == "" {
		switch req.ContentType {
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		case "application/pdf":
			ext = ".pdf"
		default:
			ext = ".jpg"
		}
	}
	contentType := req.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Dedicated bucket per WORKER_REGISTRATION_FLOW.md. In prod the
	// bucket is created out-of-band in Terraform; the service doesn't
	// create it. Falls back to the configured media bucket if the KYC
	// one is unset, so a half-configured environment still boots.
	kycBucket := "smartsewa-kyc"
	if override := strings.TrimSpace(s.cfg.GCSBucketName); override != "" && strings.Contains(override, "kyc") {
		kycBucket = override
	}

	objectPath := fmt.Sprintf("%s/%s%s", workerID.String(), req.DocumentType, strings.ToLower(ext))
	expiresAt := time.Now().Add(15 * time.Minute)

	// If GCS client is nil (dev/CI), hand back a stub signed URL that
	// points at our own /upload/photo endpoint. The client can PUT to
	// it (via multipart POST adapter in the frontend) and the backend
	// writes to its local fallback dir. Not for production use.
	if s.gcsClient == nil {
		stub := fmt.Sprintf("%s/api/v1/upload/photo?stub=1&object_path=%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), url.QueryEscape(objectPath))
		return &dto.KYCUploadURLResponse{
			UploadURL:  stub,
			ObjectURI:  fmt.Sprintf("gs://%s/%s", kycBucket, objectPath),
			MediaURL:   fmt.Sprintf("%s/api/v1/media/kyc/%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), objectPath),
			ExpiresAt:  expiresAt,
			ObjectPath: objectPath,
		}, nil
	}

	// Real GCS signed URL path. Credentials come from Application
	// Default Credentials (the Cloud Run service account must have
	// roles/iam.serviceAccountTokenCreator for self-signing).
	opts := &storage.SignedURLOptions{
		Scheme:      storage.SigningSchemeV4,
		Method:      "PUT",
		Expires:     expiresAt,
		ContentType: contentType,
	}
	bucket := s.gcsClient.Bucket(kycBucket)
	signedURL, err := bucket.SignedURL(objectPath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to sign kyc upload url: %w", err)
	}

	return &dto.KYCUploadURLResponse{
		UploadURL:  signedURL,
		ObjectURI:  fmt.Sprintf("gs://%s/%s", kycBucket, objectPath),
		MediaURL:   fmt.Sprintf("%s/api/v1/media/kyc/%s", strings.TrimRight(s.cfg.AppBaseURL, "/"), objectPath),
		ExpiresAt:  expiresAt,
		ObjectPath: objectPath,
	}, nil
}

// UpdateKYC records the uploaded KYC document URIs on the worker row.
// Partial updates are supported: a zero-value field leaves the existing
// value alone, so the applicant can upload one document at a time.
func (s *workerOnboardingService) UpdateKYC(workerID uuid.UUID, req *dto.UpdateKYCRequest) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil || worker == nil {
		return nil, errors.New("worker not found")
	}
	kyc := worker.KYCDocuments
	if req.NICFront != "" {
		kyc.NICFront = req.NICFront
	}
	if req.NICBack != "" {
		kyc.NICBack = req.NICBack
	}
	if req.Selfie != "" {
		kyc.Selfie = req.Selfie
	}
	if req.TradeCert != "" {
		kyc.TradeCert = req.TradeCert
	}
	worker.KYCDocuments = kyc
	if err := s.workerRepo.Update(worker); err != nil {
		return nil, err
	}
	res := mapWorkerToDTO(worker)
	return &res, nil
}

// GetMyStatus returns the applicant-facing status page payload.
func (s *workerOnboardingService) GetMyStatus(userID uuid.UUID) (*dto.WorkerStatusResponse, error) {
	worker, err := s.workerRepo.FindByUserID(userID)
	if err != nil || worker == nil {
		return nil, errors.New("worker application not found for this account")
	}
	return &dto.WorkerStatusResponse{
		Status:         worker.Status,
		KYCStatus:      kycStatus(worker.KYCDocuments),
		RejectedReason: worker.RejectedReason,
		BannedReason:   worker.BannedReason,
	}, nil
}

func kycStatus(k models.KYCDocs) string {
	count := 0
	if k.NICFront != "" {
		count++
	}
	if k.NICBack != "" {
		count++
	}
	if k.Selfie != "" {
		count++
	}
	// trade_cert is optional and doesn't count toward completeness
	switch {
	case count == 0:
		return "none"
	case count < 3:
		return "partial"
	default:
		return "complete"
	}
}

// UpdateMyApplication lets a worker self-edit while pending. Only pending
// applications can be edited — once approved, the worker uses PUT
// /workers/:id (admin-only via middleware).
func (s *workerOnboardingService) UpdateMyApplication(userID uuid.UUID, req *dto.UpdateMyWorkerRequest) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByUserID(userID)
	if err != nil || worker == nil {
		return nil, errors.New("worker application not found for this account")
	}
	if worker.Status != "pending" {
		return nil, fmt.Errorf("cannot edit application in status '%s'", worker.Status)
	}
	if req.Name != nil {
		worker.Name = *req.Name
	}
	if req.AvatarURL != nil {
		worker.AvatarURL = *req.AvatarURL
	}
	if req.Age != nil {
		if *req.Age < 18 {
			return nil, errors.New("age must be at least 18")
		}
		worker.Age = *req.Age
	}
	if req.ExperienceYears != nil {
		worker.ExperienceYears = *req.ExperienceYears
	}
	if req.ServiceTypes != nil {
		worker.ServiceTypes = *req.ServiceTypes
	}
	if req.Specializations != nil {
		worker.Specializations = *req.Specializations
	}
	if req.Phone != nil {
		worker.Phone = *req.Phone
	}
	if req.Bio != nil {
		worker.Bio = *req.Bio
	}
	if req.City != nil {
		worker.City = *req.City
	}
	if req.ServiceAreas != nil {
		worker.ServiceAreas = *req.ServiceAreas
	}
	if req.NICNumber != nil {
		if err := ValidateNIC(*req.NICNumber); err != nil {
			return nil, err
		}
		worker.NICNumber = strings.TrimSpace(*req.NICNumber)
	}
	if err := s.workerRepo.Update(worker); err != nil {
		return nil, err
	}
	res := mapWorkerToDTO(worker)
	return &res, nil
}

// --- admin ---

func (s *workerOnboardingService) ListForAdmin(status string, page, pageSize int) (*AdminWorkerListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	workers, total, err := s.workerRepo.ListByStatus(status, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]dto.AdminWorkerListItem, 0, len(workers))
	for _, w := range workers {
		items = append(items, dto.AdminWorkerListItem{
			ID:           w.ID,
			UserID:       w.UserID,
			Name:         w.Name,
			Email:        w.Email,
			Phone:        w.Phone,
			City:         w.City,
			Status:       w.Status,
			NICNumber:    w.NICNumber,
			CreatedAt:    w.CreatedAt,
			ServiceTypes: w.ServiceTypes,
		})
	}
	return &AdminWorkerListResult{
		Items:      items,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

func (s *workerOnboardingService) GetForAdmin(workerID uuid.UUID) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil || worker == nil {
		return nil, errors.New("worker not found")
	}
	res := mapWorkerToDTO(worker)
	return &res, nil
}

func (s *workerOnboardingService) Approve(workerID, adminID uuid.UUID) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil || worker == nil {
		return nil, errors.New("worker not found")
	}
	if worker.Status == "banned" {
		return nil, errors.New("cannot approve a banned worker; use /unban")
	}
	now := time.Now().UTC()
	worker.Status = "approved"
	worker.ApprovedAt = &now
	worker.ApprovedBy = &adminID
	worker.IsActive = true
	worker.IsVerified = true
	worker.IsOpenForWork = true
	worker.IsAvailableToday = true
	// Clear previous rejection reason if any — PDPA audit trail
	// stays in app logs; the row reflects current state.
	worker.RejectedAt = nil
	worker.RejectedReason = ""
	if err := s.workerRepo.Update(worker); err != nil {
		return nil, err
	}
	// Flip the paired user row to active so the applicant can sign in.
	if worker.UserID != nil {
		_ = s.userRepo.ToggleActive(*worker.UserID, true)
	}
	res := mapWorkerToDTO(worker)
	return &res, nil
}

func (s *workerOnboardingService) Reject(workerID uuid.UUID, reason string) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil || worker == nil {
		return nil, errors.New("worker not found")
	}
	now := time.Now().UTC()
	worker.Status = "rejected"
	worker.RejectedAt = &now
	worker.RejectedReason = reason
	worker.IsActive = false
	worker.IsOpenForWork = false
	if err := s.workerRepo.Update(worker); err != nil {
		return nil, err
	}
	if worker.UserID != nil {
		_ = s.userRepo.ToggleActive(*worker.UserID, false)
	}
	res := mapWorkerToDTO(worker)
	return &res, nil
}

func (s *workerOnboardingService) Ban(workerID, adminID uuid.UUID, reason string) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil || worker == nil {
		return nil, errors.New("worker not found")
	}
	now := time.Now().UTC()
	worker.Status = "banned"
	worker.BannedAt = &now
	worker.BannedReason = reason
	worker.BannedBy = &adminID
	worker.IsActive = false
	worker.IsOpenForWork = false
	if err := s.workerRepo.Update(worker); err != nil {
		return nil, err
	}
	if worker.UserID != nil {
		_ = s.userRepo.ToggleActive(*worker.UserID, false)
	}
	res := mapWorkerToDTO(worker)
	return &res, nil
}

func (s *workerOnboardingService) Unban(workerID uuid.UUID) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil || worker == nil {
		return nil, errors.New("worker not found")
	}
	if worker.Status != "banned" {
		return nil, fmt.Errorf("cannot unban: worker status is '%s', expected 'banned'", worker.Status)
	}
	worker.Status = "approved"
	worker.BannedAt = nil
	worker.BannedReason = ""
	worker.BannedBy = nil
	worker.IsActive = true
	worker.IsOpenForWork = true
	if err := s.workerRepo.Update(worker); err != nil {
		return nil, err
	}
	if worker.UserID != nil {
		_ = s.userRepo.ToggleActive(*worker.UserID, true)
	}
	res := mapWorkerToDTO(worker)
	return &res, nil
}

// NewGCSClientOrNil tries to initialise a Google Cloud Storage client
// using Application Default Credentials. Returns (nil, err) on failure
// so main() can log + continue with the stub signer. Exported for use
// from main.go.
func NewGCSClientOrNil(ctx context.Context) (*storage.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return storage.NewClient(ctx)
}
