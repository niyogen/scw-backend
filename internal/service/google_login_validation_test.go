package service

import (
	"context"
	"errors"
	"testing"

	"delivery-backend/internal/config"
	"delivery-backend/internal/dto"
)

// fakeGoogleValidator lets each test case control what the validator
// returns without needing a live Google-signed JWT. It replaces the
// package-level validator for the test's duration.
type fakeGoogleValidator struct {
	tokens map[string]*googleTokenPayload
	err    error
}

func (f *fakeGoogleValidator) Validate(ctx context.Context, idToken, audience string) (*googleTokenPayload, error) {
	if f.err != nil {
		return nil, f.err
	}
	if p, ok := f.tokens[idToken]; ok {
		return p, nil
	}
	return nil, errors.New("fake validator: unknown token")
}

// swapValidator installs the given fake validator and returns a cleanup
// closure that restores the production validator. Tests should call the
// cleanup via t.Cleanup to avoid leaking test state across tests.
func swapValidator(t *testing.T, v googleTokenValidator) {
	t.Helper()
	prev := pkgGoogleValidator
	pkgGoogleValidator = v
	t.Cleanup(func() { pkgGoogleValidator = prev })
}

func newAuthServiceForGoogleTest(cfg *config.Config) (AuthService, *mockUserRepo, *mockWorkerRepo) {
	userRepo := newMockUserRepo()
	workerRepo := newMockWorkerRepo()
	return NewAuthService(userRepo, workerRepo, nil, nil, nil, cfg), userRepo, workerRepo
}

func baseCfg(env string) *config.Config {
	return &config.Config{
		Environment:    env,
		JWTSecret:      "test-secret-dont-use-anywhere",
		JWTExpiresIn:   72,
		GoogleClientID: "test-client-id-12345.apps.googleusercontent.com",
	}
}

// --- Production branch tests ---

func TestGoogleLogin_Production_RequiresIDToken(t *testing.T) {
	svc, _, _ := newAuthServiceForGoogleTest(baseCfg("production"))
	// No validator swap needed — we never reach it when id_token is empty.
	_, err := svc.GoogleLogin(&dto.GoogleLoginRequest{
		Email: "attacker@invalid.lk",
		Name:  "Attacker",
	})
	if err == nil {
		t.Fatalf("expected error when id_token missing in production")
	}
	if err.Error() != "google id_token is required" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGoogleLogin_Production_ValidTokenUsesVerifiedClaims(t *testing.T) {
	swapValidator(t, &fakeGoogleValidator{
		tokens: map[string]*googleTokenPayload{
			"valid-token-xyz": {
				Email:         "verified@gmail.com",
				EmailVerified: true,
				Name:          "Verified Name",
				Picture:       "https://example.com/pic.jpg",
				Subject:       "google-user-sub-id",
			},
		},
	})
	svc, _, _ := newAuthServiceForGoogleTest(baseCfg("production"))
	resp, err := svc.GoogleLogin(&dto.GoogleLoginRequest{
		IDToken: "valid-token-xyz",
		// Client-supplied email+name should be ignored in favour of
		// the verified claims — this is the whole point of validation.
		Email: "attacker-tries-this@invalid.lk",
		Name:  "Attacker",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Email != "verified@gmail.com" {
		t.Errorf("expected verified email to be used, got %q", resp.User.Email)
	}
	if resp.User.Name != "Verified Name" {
		t.Errorf("expected verified name to be used, got %q", resp.User.Name)
	}
	if resp.Token == "" {
		t.Error("expected JWT to be issued")
	}
}

func TestGoogleLogin_Production_InvalidTokenRejected(t *testing.T) {
	swapValidator(t, &fakeGoogleValidator{err: errors.New("signature verification failed")})
	svc, _, _ := newAuthServiceForGoogleTest(baseCfg("production"))
	_, err := svc.GoogleLogin(&dto.GoogleLoginRequest{
		IDToken: "forged-token",
	})
	if err == nil {
		t.Fatalf("expected rejection for invalid token")
	}
	if err.Error() == "" || !containsSubstr(err.Error(), "invalid google id token") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGoogleLogin_Production_UnverifiedEmailRejected(t *testing.T) {
	swapValidator(t, &fakeGoogleValidator{
		tokens: map[string]*googleTokenPayload{
			"unverified": {
				Email:         "unverified@gmail.com",
				EmailVerified: false,
				Name:          "Unverified",
				Subject:       "sub-id",
			},
		},
	})
	svc, _, _ := newAuthServiceForGoogleTest(baseCfg("production"))
	_, err := svc.GoogleLogin(&dto.GoogleLoginRequest{IDToken: "unverified"})
	if err == nil {
		t.Fatalf("expected rejection when email is not verified")
	}
}

func TestGoogleLogin_Production_MissingClientID(t *testing.T) {
	cfg := baseCfg("production")
	cfg.GoogleClientID = "" // misconfigured deployment
	svc, _, _ := newAuthServiceForGoogleTest(cfg)
	_, err := svc.GoogleLogin(&dto.GoogleLoginRequest{IDToken: "doesnt-matter"})
	if err == nil {
		t.Fatalf("expected error when GOOGLE_CLIENT_ID is empty")
	}
}

// --- Non-production (staging / dev) branch tests ---

func TestGoogleLogin_Staging_FallbackMockWorks(t *testing.T) {
	// In non-prod, no id_token is OK — the client-supplied email/name
	// are accepted as a convenience for CI and manual test flows.
	svc, _, _ := newAuthServiceForGoogleTest(baseCfg("staging"))
	resp, err := svc.GoogleLogin(&dto.GoogleLoginRequest{
		Email: "staging-test@invalid.lk",
		Name:  "Staging Test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Email != "staging-test@invalid.lk" {
		t.Errorf("expected mock email to be used in non-prod, got %q", resp.User.Email)
	}
}

func TestGoogleLogin_Staging_RealTokenStillPreferred(t *testing.T) {
	// Even in non-prod, if a real token is provided it must still
	// validate — we don't let staging users forge tokens.
	swapValidator(t, &fakeGoogleValidator{
		tokens: map[string]*googleTokenPayload{
			"ok": {
				Email:         "real-verified@gmail.com",
				EmailVerified: true,
				Name:          "Real User",
				Subject:       "sub-id",
			},
		},
	})
	svc, _, _ := newAuthServiceForGoogleTest(baseCfg("staging"))
	resp, err := svc.GoogleLogin(&dto.GoogleLoginRequest{
		IDToken: "ok",
		Email:   "ignored-mock-email@invalid.lk",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Email != "real-verified@gmail.com" {
		t.Errorf("even in non-prod the verified claims should win; got %q", resp.User.Email)
	}
}

func containsSubstr(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
