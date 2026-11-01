package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"delivery-backend/internal/config"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/middleware"
	"delivery-backend/internal/models"
	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- mocks ---

type fakeUserRepo struct {
	byEmail map[string]*models.User
	byID    map[uuid.UUID]*models.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		byEmail: make(map[string]*models.User),
		byID:    make(map[uuid.UUID]*models.User),
	}
}

func (f *fakeUserRepo) Create(user *models.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	f.byEmail[user.Email] = user
	f.byID[user.ID] = user
	return nil
}

func (f *fakeUserRepo) FindByID(id uuid.UUID) (*models.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, nil
}

func (f *fakeUserRepo) FindByEmailOrPhone(input string) (*models.User, error) {
	if u, ok := f.byEmail[input]; ok {
		return u, nil
	}
	return nil, nil
}

func (f *fakeUserRepo) FindByEmail(email string) (*models.User, error) {
	if u, ok := f.byEmail[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (f *fakeUserRepo) FindByPhone(phone string) (*models.User, error) {
	for _, u := range f.byEmail {
		if u.Phone == phone {
			return u, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepo) FindAll(role string) ([]models.User, error) {
	var list []models.User
	for _, u := range f.byEmail {
		if role == "" || u.Role == role {
			list = append(list, *u)
		}
	}
	return list, nil
}

func (f *fakeUserRepo) Update(user *models.User) error {
	f.byEmail[user.Email] = user
	f.byID[user.ID] = user
	return nil
}

func (f *fakeUserRepo) UpdatePassword(id uuid.UUID, newHash string) error {
	if u, ok := f.byID[id]; ok {
		u.PasswordHash = newHash
	}
	return nil
}

func (f *fakeUserRepo) Delete(id uuid.UUID) error {
	if u, ok := f.byID[id]; ok {
		delete(f.byEmail, u.Email)
		delete(f.byID, id)
	}
	return nil
}

func (f *fakeUserRepo) ToggleActive(id uuid.UUID, isActive bool) error {
	if u, ok := f.byID[id]; ok {
		u.IsActive = isActive
	}
	return nil
}

type fakeWorkerRepo struct {
	workers map[uuid.UUID]*models.Worker
}

func newFakeWorkerRepo() *fakeWorkerRepo {
	return &fakeWorkerRepo{workers: make(map[uuid.UUID]*models.Worker)}
}

func (f *fakeWorkerRepo) FindAll(serviceType string, availableOnly bool, all bool) ([]models.Worker, error) {
	var list []models.Worker
	for _, w := range f.workers {
		if !all && !w.IsActive {
			continue
		}
		if !all && w.Status != "approved" && w.Status != "" {
			continue
		}
		list = append(list, *w)
	}
	return list, nil
}

func (f *fakeWorkerRepo) FindByIDOrCode(idOrCode string) (*models.Worker, error) {
	id, err := uuid.Parse(idOrCode)
	if err == nil {
		if w, ok := f.workers[id]; ok {
			return w, nil
		}
	}
	for _, w := range f.workers {
		if w.WorkerCode == idOrCode {
			return w, nil
		}
	}
	return nil, nil
}

func (f *fakeWorkerRepo) FindByUserID(userID uuid.UUID) (*models.Worker, error) {
	for _, w := range f.workers {
		if w.UserID != nil && *w.UserID == userID {
			return w, nil
		}
	}
	return nil, nil
}

func (f *fakeWorkerRepo) FindByEmail(email string) (*models.Worker, error) {
	for _, w := range f.workers {
		if w.Email == email {
			return w, nil
		}
	}
	return nil, nil
}

func (f *fakeWorkerRepo) Create(w *models.Worker) error {
	w.ID = uuid.New()
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	f.workers[w.ID] = w
	return nil
}

func (f *fakeWorkerRepo) Update(w *models.Worker) error {
	w.UpdatedAt = time.Now()
	f.workers[w.ID] = w
	return nil
}

func (f *fakeWorkerRepo) UpdateLocationStatus(id uuid.UUID, lat, lng float64, address string, isOpen *bool, radiusKm *float64) error {
	return nil
}

func (f *fakeWorkerRepo) ToggleActive(id uuid.UUID, isActive bool) error {
	if w, ok := f.workers[id]; ok {
		w.IsActive = isActive
	}
	return nil
}

func (f *fakeWorkerRepo) Delete(id uuid.UUID) error {
	delete(f.workers, id)
	return nil
}

func (f *fakeWorkerRepo) AddReview(review *models.WorkerReview) error { return nil }
func (f *fakeWorkerRepo) GetReviews(workerID uuid.UUID) ([]models.WorkerReview, error) {
	return nil, nil
}
func (f *fakeWorkerRepo) GetAllReviews(status string) ([]models.WorkerReview, error) {
	return nil, nil
}
func (f *fakeWorkerRepo) FindReviewByID(id uuid.UUID) (*models.WorkerReview, error) { return nil, nil }
func (f *fakeWorkerRepo) UpdateReviewStatus(id uuid.UUID, isApproved bool) error     { return nil }
func (f *fakeWorkerRepo) DeleteReview(id uuid.UUID) error                            { return nil }
func (f *fakeWorkerRepo) UpdateWorkerRating(workerID uuid.UUID) error                { return nil }
func (f *fakeWorkerRepo) IncrementCompletedJobs(workerID uuid.UUID) error            { return nil }

func (f *fakeWorkerRepo) ListByStatus(status string, offset, limit int) ([]models.Worker, int64, error) {
	var list []models.Worker
	for _, w := range f.workers {
		if status == "" || w.Status == status {
			list = append(list, *w)
		}
	}
	total := int64(len(list))
	if offset >= len(list) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end], total, nil
}

// --- test fixtures ---

func newOnboardingRouter(t *testing.T) (*gin.Engine, *fakeUserRepo, *fakeWorkerRepo, *config.Config) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Environment:  "test",
		JWTSecret:    "ob-test-secret-" + t.Name(),
		JWTExpiresIn: 1,
		AppBaseURL:   "http://localhost:8080",
	}
	userRepo := newFakeUserRepo()
	workerRepo := newFakeWorkerRepo()
	svc := service.NewWorkerOnboardingService(workerRepo, userRepo, cfg, nil)
	h := NewWorkerOnboardingHandler(svc)

	r := gin.New()
	v1 := r.Group("/api/v1")
	// Public
	v1.POST("/workers/register-v2", h.RegisterSelfService)
	// Authenticated self-service
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	protected.GET("/workers/me/status", h.GetMyStatus)
	protected.PUT("/workers/me", h.UpdateMyApplication)
	protected.POST("/workers/:id/kyc-upload-url", h.SignKYCUploadURL)
	protected.PUT("/workers/:id/kyc", h.UpdateKYC)
	// Admin
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	admin.Use(middleware.RequireRole("admin"))
	admin.GET("/workers", h.AdminListWorkers)
	admin.GET("/workers/:id", h.AdminGetWorker)
	admin.POST("/workers/:id/approve", h.AdminApproveWorker)
	admin.POST("/workers/:id/reject", h.AdminRejectWorker)
	admin.POST("/workers/:id/ban", h.AdminBanWorker)
	admin.POST("/workers/:id/unban", h.AdminUnbanWorker)
	return r, userRepo, workerRepo, cfg
}

func doJSON(r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

type envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// --- tests ---

func TestRegisterSelfService_HappyPath(t *testing.T) {
	r, userRepo, workerRepo, _ := newOnboardingRouter(t)

	body := `{
		"name": "Kasun Perera",
		"email": "kasun@test.lk",
		"phone": "+94 77 123 4567",
		"age": 30,
		"experience_years": 5,
		"service_types": ["plumbing"],
		"bio": "Plumber with 5 years",
		"nic_number": "199012345678",
		"city": "Colombo 05"
	}`
	rec := doJSON(r, "POST", "/api/v1/workers/register-v2", body, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	var payload dto.RegisterWorkerResponse
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("bad payload: %v", err)
	}
	if payload.Status != "pending" {
		t.Errorf("want status=pending, got %q", payload.Status)
	}
	if payload.Token == "" {
		t.Error("expected JWT in response")
	}
	// Users row should exist with role=worker, is_active=false.
	u, _ := userRepo.FindByEmail("kasun@test.lk")
	if u == nil {
		t.Fatal("user row not created")
	}
	if u.Role != "worker" {
		t.Errorf("user role: want worker, got %q", u.Role)
	}
	if u.IsActive {
		t.Error("new worker user should start inactive")
	}
	// Workers row should exist with status=pending.
	if len(workerRepo.workers) != 1 {
		t.Fatalf("want 1 worker, got %d", len(workerRepo.workers))
	}
	for _, w := range workerRepo.workers {
		if w.Status != "pending" {
			t.Errorf("worker status: want pending, got %q", w.Status)
		}
		if w.NICNumber != "199012345678" {
			t.Errorf("NIC not persisted; got %q", w.NICNumber)
		}
	}
}

func TestRegisterSelfService_AgeGate(t *testing.T) {
	r, _, _, _ := newOnboardingRouter(t)
	rec := doJSON(r, "POST", "/api/v1/workers/register-v2", `{
		"name":"Teen","email":"teen@test.lk","phone":"+94 77 111 1111","age":17,
		"service_types":["plumbing"]
	}`, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for age<18, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterSelfService_InvalidNIC(t *testing.T) {
	r, _, _, _ := newOnboardingRouter(t)
	rec := doJSON(r, "POST", "/api/v1/workers/register-v2", `{
		"name":"Bad NIC","email":"bn@test.lk","phone":"+94 77 222 2222","age":30,
		"service_types":["plumbing"],
		"nic_number":"NOT-A-NIC"
	}`, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for invalid NIC, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRegisterSelfService_NIC_OldFormat(t *testing.T) {
	r, _, _, _ := newOnboardingRouter(t)
	rec := doJSON(r, "POST", "/api/v1/workers/register-v2", `{
		"name":"Old NIC","email":"old@test.lk","phone":"+94 77 333 3333","age":30,
		"service_types":["plumbing"],
		"nic_number":"901234567V"
	}`, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201 for old-format NIC, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetMyStatus_HappyPath(t *testing.T) {
	r, _, workerRepo, cfg := newOnboardingRouter(t)
	// Seed a pending worker directly.
	userID := uuid.New()
	w := &models.Worker{UserID: &userID, Email: "s@test.lk", Status: "pending"}
	_ = workerRepo.Create(w)

	tok, _ := utils.GenerateJWT(userID, "s@test.lk", "worker", cfg.JWTSecret, 1)
	rec := doJSON(r, "GET", "/api/v1/workers/me/status", "", tok)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var env envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	var payload dto.WorkerStatusResponse
	_ = json.Unmarshal(env.Data, &payload)
	if payload.Status != "pending" {
		t.Errorf("want status=pending, got %q", payload.Status)
	}
	if payload.KYCStatus != "none" {
		t.Errorf("want kyc_status=none, got %q", payload.KYCStatus)
	}
}

func TestUpdateMyApplication_RejectsAfterApproval(t *testing.T) {
	r, _, workerRepo, cfg := newOnboardingRouter(t)
	userID := uuid.New()
	w := &models.Worker{UserID: &userID, Email: "a@test.lk", Status: "approved", IsActive: true}
	_ = workerRepo.Create(w)
	tok, _ := utils.GenerateJWT(userID, "a@test.lk", "worker", cfg.JWTSecret, 1)
	rec := doJSON(r, "PUT", "/api/v1/workers/me", `{"bio":"updated"}`, tok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for edit while approved, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestKYCUploadURL_StubReturnsExpiryAndPath(t *testing.T) {
	r, _, workerRepo, cfg := newOnboardingRouter(t)
	userID := uuid.New()
	w := &models.Worker{UserID: &userID, Email: "k@test.lk", Status: "pending"}
	_ = workerRepo.Create(w)
	tok, _ := utils.GenerateJWT(userID, "k@test.lk", "worker", cfg.JWTSecret, 1)

	body := `{"document_type":"nic_front","content_type":"image/jpeg","filename":"nic.jpg"}`
	rec := doJSON(r, "POST", "/api/v1/workers/"+w.ID.String()+"/kyc-upload-url", body, tok)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var env envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	var payload dto.KYCUploadURLResponse
	_ = json.Unmarshal(env.Data, &payload)
	if payload.UploadURL == "" || payload.ObjectURI == "" {
		t.Fatalf("missing fields in signed URL response: %+v", payload)
	}
	if payload.ExpiresAt.Before(time.Now()) {
		t.Errorf("expiry should be in the future, got %v", payload.ExpiresAt)
	}
}

func TestKYCUploadURL_InvalidDocumentType(t *testing.T) {
	r, _, workerRepo, cfg := newOnboardingRouter(t)
	userID := uuid.New()
	w := &models.Worker{UserID: &userID, Email: "ki@test.lk", Status: "pending"}
	_ = workerRepo.Create(w)
	tok, _ := utils.GenerateJWT(userID, "ki@test.lk", "worker", cfg.JWTSecret, 1)
	body := `{"document_type":"passport","content_type":"image/jpeg","filename":"p.jpg"}`
	rec := doJSON(r, "POST", "/api/v1/workers/"+w.ID.String()+"/kyc-upload-url", body, tok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for invalid doc type, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdmin_ApproveFlipsIsActive(t *testing.T) {
	r, userRepo, workerRepo, cfg := newOnboardingRouter(t)
	userID := uuid.New()
	u := &models.User{Name: "N", Email: "n@test.lk", Phone: "+94 77 999 9999", Role: "worker", IsActive: false}
	u.ID = userID
	userRepo.byEmail[u.Email] = u
	userRepo.byID[u.ID] = u
	w := &models.Worker{UserID: &userID, Email: "n@test.lk", Status: "pending"}
	_ = workerRepo.Create(w)

	admin := uuid.New()
	adminTok, _ := utils.GenerateJWT(admin, "ad@test.lk", "admin", cfg.JWTSecret, 1)

	rec := doJSON(r, "POST", "/api/v1/admin/workers/"+w.ID.String()+"/approve", "", adminTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 for approve, got %d body=%s", rec.Code, rec.Body.String())
	}
	// Worker should be approved + active.
	w2 := workerRepo.workers[w.ID]
	if w2.Status != "approved" {
		t.Errorf("worker status: want approved, got %q", w2.Status)
	}
	if !w2.IsActive {
		t.Error("worker should be active after approve")
	}
	// Paired user row should be is_active=true.
	if !userRepo.byID[userID].IsActive {
		t.Error("user should be active after worker approve")
	}
}

func TestAdmin_RejectRequiresReason(t *testing.T) {
	r, _, workerRepo, cfg := newOnboardingRouter(t)
	userID := uuid.New()
	w := &models.Worker{UserID: &userID, Email: "rj@test.lk", Status: "pending"}
	_ = workerRepo.Create(w)
	adminTok, _ := utils.GenerateJWT(uuid.New(), "a@test.lk", "admin", cfg.JWTSecret, 1)
	// Empty body → binding fails on required reason.
	rec := doJSON(r, "POST", "/api/v1/admin/workers/"+w.ID.String()+"/reject", `{}`, adminTok)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 when reason omitted, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdmin_BanAndUnban_Cycle(t *testing.T) {
	r, userRepo, workerRepo, cfg := newOnboardingRouter(t)
	userID := uuid.New()
	u := &models.User{Name: "B", Email: "b@test.lk", Phone: "+94 77 888 8888", Role: "worker", IsActive: true}
	u.ID = userID
	userRepo.byEmail[u.Email] = u
	userRepo.byID[u.ID] = u
	w := &models.Worker{UserID: &userID, Email: "b@test.lk", Status: "approved", IsActive: true}
	_ = workerRepo.Create(w)

	adminTok, _ := utils.GenerateJWT(uuid.New(), "a@test.lk", "admin", cfg.JWTSecret, 1)

	// Ban.
	rec := doJSON(r, "POST", "/api/v1/admin/workers/"+w.ID.String()+"/ban",
		`{"reason":"policy violation"}`, adminTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("ban → want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if workerRepo.workers[w.ID].Status != "banned" {
		t.Errorf("after ban: status=%q", workerRepo.workers[w.ID].Status)
	}
	if userRepo.byID[userID].IsActive {
		t.Error("user should be inactive after ban")
	}
	// Unban.
	rec = doJSON(r, "POST", "/api/v1/admin/workers/"+w.ID.String()+"/unban", "", adminTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("unban → want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if workerRepo.workers[w.ID].Status != "approved" {
		t.Errorf("after unban: status=%q", workerRepo.workers[w.ID].Status)
	}
	if !userRepo.byID[userID].IsActive {
		t.Error("user should be active after unban")
	}
}

func TestAdmin_ListPaginated(t *testing.T) {
	r, _, workerRepo, cfg := newOnboardingRouter(t)
	for i := 0; i < 5; i++ {
		uid := uuid.New()
		_ = workerRepo.Create(&models.Worker{UserID: &uid, Email: "p@" + uid.String()[:4] + ".lk", Status: "pending"})
	}
	for i := 0; i < 3; i++ {
		uid := uuid.New()
		_ = workerRepo.Create(&models.Worker{UserID: &uid, Email: "a@" + uid.String()[:4] + ".lk", Status: "approved"})
	}
	adminTok, _ := utils.GenerateJWT(uuid.New(), "a@test.lk", "admin", cfg.JWTSecret, 1)

	rec := doJSON(r, "GET", "/api/v1/admin/workers?status=pending&page=1&page_size=3", "", adminTok)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var env envelope
	_ = json.Unmarshal(rec.Body.Bytes(), &env)
	var paginated utils.PaginatedResponse
	_ = json.Unmarshal(env.Data, &paginated)
	if paginated.TotalCount != 5 {
		t.Errorf("want total 5 pending, got %d", paginated.TotalCount)
	}
	if paginated.PageSize != 3 {
		t.Errorf("want page_size=3, got %d", paginated.PageSize)
	}
}
