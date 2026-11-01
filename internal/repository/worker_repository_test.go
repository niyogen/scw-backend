package repository

import (
	"testing"

	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Workers in the customer-facing /workers listing MUST be filtered to
// (status='approved' OR legacy NULL/empty) AND is_active=true. This test
// backs that invariant up with an in-memory sqlite DB — if the filter
// regresses, the mobile app would start showing pending/rejected/banned
// applicants.
//
// We don't use Postgres jsonb features here; the test exercises the
// status + is_active guard only (which was the Task #25 regression risk).

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Each test gets its own in-memory sqlite DB (no shared cache so
	// parallel test runs don't collide).
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// AutoMigrate on the real models fails on sqlite because the GORM
	// tags include Postgres-specific DDL (gen_random_uuid(), jsonb).
	// We create a sqlite-shaped schema by hand that matches the
	// columns the FindAll + ListByStatus queries touch.
	stmt := `
	CREATE TABLE workers (
		id TEXT PRIMARY KEY,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		user_id TEXT,
		worker_code TEXT,
		name TEXT,
		avatar_url TEXT,
		age INTEGER,
		experience_years INTEGER,
		rating REAL DEFAULT 5.0,
		reviews_count INTEGER DEFAULT 0,
		service_types TEXT,
		specializations TEXT,
		phone TEXT,
		email TEXT,
		bio TEXT,
		completed_jobs INTEGER DEFAULT 0,
		hourly_rate TEXT,
		rate_value REAL,
		latitude REAL,
		longitude REAL,
		operating_address TEXT,
		is_open_for_work INTEGER,
		radius_km REAL,
		is_verified INTEGER,
		is_active INTEGER,
		is_available_today INTEGER,
		badges TEXT,
		status TEXT DEFAULT 'approved',
		nic_number TEXT,
		city TEXT,
		service_areas TEXT,
		kyc_documents TEXT,
		approved_at DATETIME,
		approved_by TEXT,
		rejected_at DATETIME,
		rejected_reason TEXT,
		banned_at DATETIME,
		banned_reason TEXT,
		banned_by TEXT
	);
	`
	if err := db.Exec(stmt).Error; err != nil {
		t.Fatalf("manual schema: %v", err)
	}
	return db
}

func seedWorker(t *testing.T, db *gorm.DB, status string, isActive bool, name string) *models.Worker {
	t.Helper()
	w := &models.Worker{
		UserID:          nil,
		WorkerCode:      "wrk_" + uuid.New().String()[:8],
		Name:            name,
		AvatarURL:       "https://example.com/a.jpg",
		Age:             30,
		ExperienceYears: 5,
		Rating:          5.0,
		ServiceTypes:    []string{"plumbing"},
		Specializations: []string{"General"},
		Phone:           "+94 77 000 0000",
		Email:           name + "@test.lk",
		HourlyRate:      "Rs. 1,500 / hr",
		RateValue:       1500,
		IsActive:        isActive,
		IsVerified:      true,
		Status:          status,
	}
	if err := db.Create(w).Error; err != nil {
		t.Fatalf("seed %s: %v", name, err)
	}
	return w
}

func TestFindAll_FiltersToApprovedActive(t *testing.T) {
	db := newTestDB(t)
	repo := NewWorkerRepository(db)

	seedWorker(t, db, "approved", true, "ok1")
	seedWorker(t, db, "approved", false, "inactive")   // approved but inactive — filtered out
	seedWorker(t, db, "pending", true, "pending-live") // pending — filtered out
	seedWorker(t, db, "rejected", true, "rejected")    // rejected — filtered out
	seedWorker(t, db, "banned", true, "banned")        // banned — filtered out
	seedWorker(t, db, "", true, "legacy-null")         // legacy NULL/empty status — kept (back-compat)
	seedWorker(t, db, "approved", true, "ok2")

	workers, err := repo.FindAll("", false, false)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	names := make(map[string]bool)
	for _, w := range workers {
		names[w.Name] = true
	}
	// Must include approved+active and legacy empty-status rows.
	for _, want := range []string{"ok1", "ok2", "legacy-null"} {
		if !names[want] {
			t.Errorf("FindAll missing expected worker %q (got %v)", want, names)
		}
	}
	// Must exclude everything else.
	for _, exclude := range []string{"inactive", "pending-live", "rejected", "banned"} {
		if names[exclude] {
			t.Errorf("FindAll leaked %s row to customer-facing listing", exclude)
		}
	}
}

func TestFindAll_AllFlagBypasses(t *testing.T) {
	db := newTestDB(t)
	repo := NewWorkerRepository(db)

	seedWorker(t, db, "approved", true, "a")
	seedWorker(t, db, "pending", true, "p")
	seedWorker(t, db, "banned", false, "b")

	all, err := repo.FindAll("", false, true)
	if err != nil {
		t.Fatalf("FindAll(all=true): %v", err)
	}
	if len(all) != 3 {
		t.Errorf("admin listing with all=true should see every row; got %d", len(all))
	}
}

func TestListByStatus_PaginatesAndCountsCorrectly(t *testing.T) {
	db := newTestDB(t)
	repo := NewWorkerRepository(db)

	for i := 0; i < 7; i++ {
		seedWorker(t, db, "pending", false, "p"+uuid.New().String()[:4])
	}
	for i := 0; i < 3; i++ {
		seedWorker(t, db, "approved", true, "a"+uuid.New().String()[:4])
	}

	// Status filter + pagination.
	page1, total, err := repo.ListByStatus("pending", 0, 3)
	if err != nil {
		t.Fatalf("ListByStatus: %v", err)
	}
	if total != 7 {
		t.Errorf("want total=7 pending, got %d", total)
	}
	if len(page1) != 3 {
		t.Errorf("want 3 items on page 1, got %d", len(page1))
	}

	// Last page may be short.
	page3, _, _ := repo.ListByStatus("pending", 6, 3)
	if len(page3) != 1 {
		t.Errorf("want 1 item on last page, got %d", len(page3))
	}

	// Empty-status filter returns every row.
	all, allTotal, _ := repo.ListByStatus("", 0, 100)
	if allTotal != 10 {
		t.Errorf("want total=10 across all statuses, got %d", allTotal)
	}
	if len(all) != 10 {
		t.Errorf("want 10 rows, got %d", len(all))
	}
}
