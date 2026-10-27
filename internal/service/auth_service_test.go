package service

import (
	"testing"
	"time"

	"delivery-backend/internal/config"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"

	"github.com/google/uuid"
)

type mockUserRepo struct {
	users map[string]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*models.User)}
}

func (m *mockUserRepo) Create(user *models.User) error {
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByID(id uuid.UUID) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByEmailOrPhone(input string) (*models.User, error) {
	if u, ok := m.users[input]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) FindByEmail(email string) (*models.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) FindByPhone(phone string) (*models.User, error) {
	for _, u := range m.users {
		if u.Phone == phone {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindAll(role string) ([]models.User, error) {
	var list []models.User
	for _, u := range m.users {
		if role == "" || u.Role == role {
			list = append(list, *u)
		}
	}
	return list, nil
}

func (m *mockUserRepo) Update(user *models.User) error {
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) UpdatePassword(id uuid.UUID, newHash string) error {
	for _, u := range m.users {
		if u.ID == id {
			u.PasswordHash = newHash
			return nil
		}
	}
	return nil
}

func (m *mockUserRepo) Delete(id uuid.UUID) error {
	for k, u := range m.users {
		if u.ID == id {
			delete(m.users, k)
			return nil
		}
	}
	return nil
}

func (m *mockUserRepo) ToggleActive(id uuid.UUID, isActive bool) error {
	for _, u := range m.users {
		if u.ID == id {
			u.IsActive = isActive
			return nil
		}
	}
	return nil
}

type mockWorkerRepo struct {
	workers map[string]*models.Worker
}

func newMockWorkerRepo() *mockWorkerRepo {
	return &mockWorkerRepo{workers: make(map[string]*models.Worker)}
}

func (m *mockWorkerRepo) FindAll(serviceType string, availableOnly bool, all bool) ([]models.Worker, error) {
	var list []models.Worker
	for _, w := range m.workers {
		list = append(list, *w)
	}
	return list, nil
}

func (m *mockWorkerRepo) FindByIDOrCode(idOrCode string) (*models.Worker, error) {
	for _, w := range m.workers {
		if w.ID.String() == idOrCode || w.WorkerCode == idOrCode {
			return w, nil
		}
	}
	return nil, nil
}

func (m *mockWorkerRepo) FindByUserID(userID uuid.UUID) (*models.Worker, error) {
	for _, w := range m.workers {
		if w.UserID != nil && *w.UserID == userID {
			return w, nil
		}
	}
	return nil, nil
}

func (m *mockWorkerRepo) FindByEmail(email string) (*models.Worker, error) {
	if w, ok := m.workers[email]; ok {
		return w, nil
	}
	return nil, nil
}

func (m *mockWorkerRepo) Create(w *models.Worker) error {
	w.ID = uuid.New()
	m.workers[w.Email] = w
	return nil
}

func (m *mockWorkerRepo) Update(w *models.Worker) error {
	m.workers[w.Email] = w
	return nil
}

func (m *mockWorkerRepo) UpdateLocationStatus(id uuid.UUID, lat, lng float64, address string, isOpen *bool, radiusKm *float64) error {
	return nil
}

func (m *mockWorkerRepo) ToggleActive(id uuid.UUID, isActive bool) error {
	return nil
}

func (m *mockWorkerRepo) Delete(id uuid.UUID) error {
	for k, w := range m.workers {
		if w.ID == id {
			delete(m.workers, k)
			return nil
		}
	}
	return nil
}

func (m *mockWorkerRepo) AddReview(review *models.WorkerReview) error {
	return nil
}

func (m *mockWorkerRepo) GetReviews(workerID uuid.UUID) ([]models.WorkerReview, error) {
	return nil, nil
}

func (m *mockWorkerRepo) GetAllReviews(status string) ([]models.WorkerReview, error) {
	return nil, nil
}

func (m *mockWorkerRepo) FindReviewByID(id uuid.UUID) (*models.WorkerReview, error) {
	return nil, nil
}

func (m *mockWorkerRepo) UpdateReviewStatus(id uuid.UUID, isApproved bool) error {
	return nil
}

func (m *mockWorkerRepo) DeleteReview(id uuid.UUID) error {
	return nil
}

func (m *mockWorkerRepo) UpdateWorkerRating(workerID uuid.UUID) error {
	return nil
}

func (m *mockWorkerRepo) IncrementCompletedJobs(workerID uuid.UUID) error {
	return nil
}

func TestGoogleLogin_RangaUser(t *testing.T) {
	userRepo := newMockUserRepo()
	workerRepo := newMockWorkerRepo()
	cfg := &config.Config{
		JWTSecret:    "test-jwt-secret-key",
		JWTExpiresIn: 24,
	}

	authSvc := NewAuthService(userRepo, workerRepo, nil, cfg)

	req := &dto.GoogleLoginRequest{
		Email:     "rangap@niyogen.com",
		Name:      "Ranga P",
		AvatarURL: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=300",
		Role:      "customer",
	}

	resp, err := authSvc.GoogleLogin(req)
	if err != nil {
		t.Fatalf("GoogleLogin failed: %v", err)
	}

	if resp.Token == "" {
		t.Errorf("Expected non-empty JWT token")
	}

	if resp.User.Email != "rangap@niyogen.com" {
		t.Errorf("Expected email rangap@niyogen.com, got %s", resp.User.Email)
	}

	if resp.User.Name != "Ranga P" {
		t.Errorf("Expected name Ranga P, got %s", resp.User.Name)
	}

	if !resp.User.IsActive {
		t.Errorf("Expected user to be active")
	}

	// Login again with same email should reuse account and issue new token
	resp2, err := authSvc.GoogleLogin(req)
	if err != nil {
		t.Fatalf("Second GoogleLogin failed: %v", err)
	}

	if resp2.User.ID != resp.User.ID {
		t.Errorf("Expected same user ID on subsequent Google login")
	}
}

func TestGoogleLogin_WorkerMode(t *testing.T) {
	userRepo := newMockUserRepo()
	workerRepo := newMockWorkerRepo()
	cfg := &config.Config{
		JWTSecret:    "test-jwt-secret-key",
		JWTExpiresIn: 24,
	}

	authSvc := NewAuthService(userRepo, workerRepo, nil, cfg)

	req := &dto.GoogleLoginRequest{
		Email:     "rangap@niyogen.com",
		Name:      "Ranga P",
		AvatarURL: "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=300",
		Role:      "worker",
	}

	resp, err := authSvc.GoogleLogin(req)
	if err != nil {
		t.Fatalf("Worker GoogleLogin failed: %v", err)
	}

	if resp.WorkerProfile == nil {
		t.Fatalf("Expected WorkerProfile in response for worker role")
	}

	if resp.WorkerProfile.Email != "rangap@niyogen.com" {
		t.Errorf("Expected worker profile email rangap@niyogen.com, got %s", resp.WorkerProfile.Email)
	}
}
