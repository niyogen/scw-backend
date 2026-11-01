package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"delivery-backend/internal/middleware"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// adminRouteCase describes one admin endpoint we want to prove is
// RBAC-protected. The table drives both the matrix of HTTP methods and
// the auth-level matrix (anon / customer / admin).
type adminRouteCase struct {
	Name       string
	Method     string
	Path       string
	Body       string // JSON body or empty
	OKStatuses []int  // acceptable success codes for admin — e.g. 200/201/204; handler bodies here are stubs returning 200
}

func TestAdminRouteGroup_RBAC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-rbac-secret-key"

	// Build a router with the exact middleware chain main.go uses. The
	// handlers themselves are stubbed so the test proves the gating,
	// not the business logic (that lives in service-layer tests).
	r := gin.New()
	v1 := r.Group("/api/v1")
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(secret))
	adminGroup.Use(middleware.RequireRole("admin"))
	{
		stub := func(c *gin.Context) {
			utils.JSONSuccess(c, http.StatusOK, "ok", gin.H{"path": c.FullPath()})
		}
		// Representative sample of the real admin routes. If a new
		// admin endpoint is added, extend the admin group and the
		// cases table below.
		adminGroup.GET("/bookings", stub)
		adminGroup.PATCH("/bookings/:id/status", stub)
		adminGroup.GET("/customers", stub)
		adminGroup.POST("/customers", stub)
		adminGroup.PUT("/customers/:id", stub)
		adminGroup.PATCH("/customers/:id/status", stub)
		adminGroup.DELETE("/customers/:id", stub)
		adminGroup.GET("/reviews", stub)
		adminGroup.PATCH("/reviews/:id/status", stub)
		adminGroup.DELETE("/reviews/:id", stub)
		adminGroup.GET("/workers", stub)
		adminGroup.GET("/workers/:id", stub)
		adminGroup.POST("/workers/:id/approve", stub)
		adminGroup.POST("/workers/:id/reject", stub)
		adminGroup.POST("/workers/:id/ban", stub)
		adminGroup.POST("/workers/:id/unban", stub)
	}
	// Admin-invite lives on /auth, not /admin, but shares the same gate.
	adminAuthGroup := v1.Group("/auth")
	adminAuthGroup.Use(middleware.AuthMiddleware(secret))
	adminAuthGroup.Use(middleware.RequireRole("admin"))
	adminAuthGroup.POST("/admin-invite", func(c *gin.Context) {
		utils.JSONSuccess(c, http.StatusCreated, "invited", nil)
	})

	cases := []adminRouteCase{
		{Name: "list bookings", Method: "GET", Path: "/api/v1/admin/bookings"},
		{Name: "update booking status", Method: "PATCH", Path: "/api/v1/admin/bookings/abc/status", Body: `{"status":"accepted"}`},
		{Name: "list customers", Method: "GET", Path: "/api/v1/admin/customers"},
		{Name: "create customer", Method: "POST", Path: "/api/v1/admin/customers", Body: `{"name":"x"}`},
		{Name: "update customer", Method: "PUT", Path: "/api/v1/admin/customers/abc", Body: `{}`},
		{Name: "toggle customer", Method: "PATCH", Path: "/api/v1/admin/customers/abc/status", Body: `{"is_active":false}`},
		{Name: "delete customer", Method: "DELETE", Path: "/api/v1/admin/customers/abc"},
		{Name: "list reviews", Method: "GET", Path: "/api/v1/admin/reviews"},
		{Name: "update review status", Method: "PATCH", Path: "/api/v1/admin/reviews/abc/status", Body: `{"is_approved":true}`},
		{Name: "delete review", Method: "DELETE", Path: "/api/v1/admin/reviews/abc"},
		{Name: "list workers", Method: "GET", Path: "/api/v1/admin/workers"},
		{Name: "get worker", Method: "GET", Path: "/api/v1/admin/workers/abc"},
		{Name: "approve worker", Method: "POST", Path: "/api/v1/admin/workers/abc/approve"},
		{Name: "reject worker", Method: "POST", Path: "/api/v1/admin/workers/abc/reject", Body: `{"reason":"bad nic"}`},
		{Name: "ban worker", Method: "POST", Path: "/api/v1/admin/workers/abc/ban", Body: `{"reason":"abuse"}`},
		{Name: "unban worker", Method: "POST", Path: "/api/v1/admin/workers/abc/unban"},
		{Name: "admin-invite", Method: "POST", Path: "/api/v1/auth/admin-invite", Body: `{"email":"x@y.lk","name":"Y"}`},
	}

	customerToken, err := utils.GenerateJWT(uuid.New(), "cust@test.lk", "customer", secret, 1)
	if err != nil {
		t.Fatalf("failed to issue customer JWT: %v", err)
	}
	adminToken, err := utils.GenerateJWT(uuid.New(), "admin@test.lk", "admin", secret, 1)
	if err != nil {
		t.Fatalf("failed to issue admin JWT: %v", err)
	}
	workerToken, err := utils.GenerateJWT(uuid.New(), "worker@test.lk", "worker", secret, 1)
	if err != nil {
		t.Fatalf("failed to issue worker JWT: %v", err)
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name+"_anon_401", func(t *testing.T) {
			rec := doRequest(r, tc.Method, tc.Path, tc.Body, "")
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("anon → want 401, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
		t.Run(tc.Name+"_customer_403", func(t *testing.T) {
			rec := doRequest(r, tc.Method, tc.Path, tc.Body, customerToken)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("customer → want 403, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
		t.Run(tc.Name+"_worker_403", func(t *testing.T) {
			rec := doRequest(r, tc.Method, tc.Path, tc.Body, workerToken)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("worker → want 403, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
		t.Run(tc.Name+"_admin_2xx", func(t *testing.T) {
			rec := doRequest(r, tc.Method, tc.Path, tc.Body, adminToken)
			if rec.Code < 200 || rec.Code >= 300 {
				t.Fatalf("admin → want 2xx, got %d body=%s", rec.Code, rec.Body.String())
			}
			// Verify the body is actually from our stub (not an error).
			var env utils.StandardResponse
			_ = json.Unmarshal(rec.Body.Bytes(), &env)
			if !env.Success {
				t.Errorf("admin got non-success body: %s", rec.Body.String())
			}
		})
	}
}

// TestAdminRoute_InvalidToken_401 is a smoke test for the middleware
// behaviour on malformed / tampered JWTs. The RequireRole chain should
// never see them — the AuthMiddleware aborts first with 401.
func TestAdminRoute_InvalidToken_401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-rbac-secret-key"
	r := gin.New()
	v1 := r.Group("/api/v1")
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(secret))
	adminGroup.Use(middleware.RequireRole("admin"))
	adminGroup.GET("/ping", func(c *gin.Context) {
		utils.JSONSuccess(c, http.StatusOK, "pong", nil)
	})

	for _, tok := range []string{
		"not-a-jwt",
		"Bearer not-a-jwt",
		"", // handled as missing header
	} {
		rec := doRequest(r, "GET", "/api/v1/admin/ping", "", tok)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("want 401 for token=%q, got %d", tok, rec.Code)
		}
	}
}

// TestAdminRoute_WrongSecret_401 proves that a JWT signed with a
// different secret is rejected (impersonation defence).
func TestAdminRoute_WrongSecret_401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rightSecret := "right-secret"
	wrongSecret := "wrong-secret"

	r := gin.New()
	v1 := r.Group("/api/v1")
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(rightSecret))
	adminGroup.Use(middleware.RequireRole("admin"))
	adminGroup.GET("/ping", func(c *gin.Context) {
		utils.JSONSuccess(c, http.StatusOK, "pong", nil)
	})

	forgedAdmin, _ := utils.GenerateJWT(uuid.New(), "a@a.lk", "admin", wrongSecret, 1)
	rec := doRequest(r, "GET", "/api/v1/admin/ping", "", forgedAdmin)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 for token with wrong secret, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// doRequest is the shared test helper. token="" sends no Authorization
// header; otherwise the token is wrapped in "Bearer <t>" automatically
// unless the string already starts with "Bearer " (so tests can send
// deliberately malformed headers too).
func doRequest(r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
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
		if len(token) >= 7 && token[:7] == "Bearer " {
			req.Header.Set("Authorization", token)
		} else {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}
