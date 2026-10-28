package service

import (
	"strings"
	"testing"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/legal"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// setupTestDB spins up an in-memory SQLite with the ConsentRecord +
// User tables migrated. We don't exercise Postgres-specific features
// in these tests (append-only writes work on both), so SQLite is a
// safe stand-in and gives us <1s test runs.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	// Note: tests only exercise ConsentRecord so we don't migrate the
	// full User schema (its FK relations into bookings/addresses pull
	// in a lot of Postgres-specific DDL that SQLite rejects).
	if err := db.AutoMigrate(&models.ConsentRecord{}); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

// liveHash is the hash the service expects for a given (docID, locale)
// against the current embedded legal bundle. Tests use this to send
// correct payloads unless they're deliberately exercising a mismatch.
func liveHash(t *testing.T, docID, locale string) string {
	t.Helper()
	h, _, err := legal.ComputeDisplayedTextHash(docID, locale)
	if err != nil {
		t.Fatalf("liveHash(%s,%s): %v", docID, locale, err)
	}
	return h
}

func liveVersion(t *testing.T, docID string) string {
	t.Helper()
	v, err := legal.Version(docID)
	if err != nil {
		t.Fatalf("liveVersion(%s): %v", docID, err)
	}
	return v
}

func validSignupConsents(t *testing.T, locale, dob string) []dto.ConsentInput {
	t.Helper()
	cid1 := uuid.New()
	cid2 := uuid.New()
	cid3 := uuid.New()
	return []dto.ConsentInput{
		{
			ConsentID:         &cid1,
			ConsentType:       "user_tc",
			DisplayedTextHash: liveHash(t, legal.DocUserTC, locale),
			DisplayedLanguage: locale,
			DocumentVersion:   liveVersion(t, legal.DocUserTC),
			Channel:           "mobile",
		},
		{
			ConsentID:         &cid2,
			ConsentType:       "user_privacy",
			DisplayedTextHash: liveHash(t, legal.DocUserPrivacy, locale),
			DisplayedLanguage: locale,
			DocumentVersion:   liveVersion(t, legal.DocUserPrivacy),
			Channel:           "mobile",
		},
		{
			ConsentID:         &cid3,
			ConsentType:       "age_confirmation",
			DisplayedTextHash: legal.ComputeAgeHash(dob),
			DisplayedLanguage: locale,
			DocumentVersion:   legal.AgeVersion,
			Channel:           "mobile",
		},
	}
}

func TestVerifyAndNormalizeInput_RejectsBadInputs(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	cases := []struct {
		name     string
		mutate   func(c *dto.ConsentInput)
		wantSubs string
	}{
		{
			"unknown consent_type",
			func(c *dto.ConsentInput) { c.ConsentType = "cookies_mystery" },
			"unknown_consent_type",
		},
		{
			"unsupported language",
			func(c *dto.ConsentInput) { c.DisplayedLanguage = "fr" },
			"unsupported_language",
		},
		{
			"missing version",
			func(c *dto.ConsentInput) { c.DocumentVersion = "" },
			"missing_document_version",
		},
		{
			"short hash",
			func(c *dto.ConsentInput) { c.DisplayedTextHash = "deadbeef" },
			"malformed_hash",
		},
		{
			"version drift",
			func(c *dto.ConsentInput) { c.DocumentVersion = "9.9.9" },
			"version_drift",
		},
		{
			"hash mismatch",
			func(c *dto.ConsentInput) {
				c.DisplayedTextHash = strings.Repeat("a", 64)
			},
			"hash_mismatch",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cid := uuid.New()
			in := dto.ConsentInput{
				ConsentID:         &cid,
				ConsentType:       "user_tc",
				DisplayedTextHash: liveHash(t, legal.DocUserTC, "en"),
				DisplayedLanguage: "en",
				DocumentVersion:   liveVersion(t, legal.DocUserTC),
				Channel:           "mobile",
				Granted:           true,
			}
			tc.mutate(&in)
			err := svc.VerifyAndNormalizeInput(&in)
			if err == nil {
				t.Fatalf("expected rejection, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantSubs) {
				t.Fatalf("expected substring %q in %q", tc.wantSubs, err.Error())
			}
		})
	}
}

func TestGrant_WritesAppendOnlyRow(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	cid := uuid.New()
	in := dto.ConsentInput{
		ConsentID:         &cid,
		ConsentType:       "user_tc",
		DisplayedTextHash: liveHash(t, legal.DocUserTC, "en"),
		DisplayedLanguage: "en",
		DocumentVersion:   liveVersion(t, legal.DocUserTC),
		Channel:           "mobile",
	}
	rec, err := svc.Grant(userID, in)
	if err != nil {
		t.Fatalf("grant failed: %v", err)
	}
	if rec.ConsentID != cid {
		t.Errorf("ConsentID round-trip failed")
	}
	if !rec.Granted || rec.Withdrawn {
		t.Error("state flags wrong on fresh grant")
	}
	expectedSubject := repository.HashSubjectID(userID)
	if rec.SubjectIDHash != expectedSubject {
		t.Error("SubjectIDHash should be derived server-side")
	}

	// Row count: exactly one
	var count int64
	db.Model(&models.ConsentRecord{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestGrant_IsIdempotentOnConsentID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	cid := uuid.New()
	in := dto.ConsentInput{
		ConsentID:         &cid,
		ConsentType:       "user_tc",
		DisplayedTextHash: liveHash(t, legal.DocUserTC, "en"),
		DisplayedLanguage: "en",
		DocumentVersion:   liveVersion(t, legal.DocUserTC),
	}
	rec1, err := svc.Grant(userID, in)
	if err != nil {
		t.Fatalf("first grant: %v", err)
	}
	rec2, err := svc.Grant(userID, in)
	if err != nil {
		t.Fatalf("retry grant: %v", err)
	}
	if rec1.ConsentID != rec2.ConsentID {
		t.Error("idempotent grant returned a different consent_id")
	}
	var count int64
	db.Model(&models.ConsentRecord{}).Count(&count)
	if count != 1 {
		t.Errorf("retry inserted a duplicate row: count=%d", count)
	}
}

func TestWithdraw_AppendsRowNotMutatesOriginal(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	cid := uuid.New()
	original, err := svc.Grant(userID, dto.ConsentInput{
		ConsentID:         &cid,
		ConsentType:       "user_tc",
		DisplayedTextHash: liveHash(t, legal.DocUserTC, "en"),
		DisplayedLanguage: "en",
		DocumentVersion:   liveVersion(t, legal.DocUserTC),
	})
	if err != nil {
		t.Fatalf("grant: %v", err)
	}

	withdrawal, err := svc.Withdraw(userID, dto.ConsentWithdrawRequest{
		ConsentType:       "user_tc",
		DisplayedLanguage: "en",
		DocumentVersion:   liveVersion(t, legal.DocUserTC),
	})
	if err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	if withdrawal.ConsentID == original.ConsentID {
		t.Error("withdrawal reused the original consent_id instead of creating a new row")
	}

	// Verify the original row is untouched
	var reread models.ConsentRecord
	if err := db.First(&reread, "consent_id = ?", original.ConsentID).Error; err != nil {
		t.Fatalf("re-read original: %v", err)
	}
	if !reread.Granted || reread.Withdrawn {
		t.Error("original grant row was mutated — append-only invariant violated")
	}

	// And the withdrawal row has the expected shape
	if withdrawal.Granted || !withdrawal.Withdrawn || withdrawal.WithdrawnAt == nil {
		t.Error("withdrawal row shape wrong")
	}
	var count int64
	db.Model(&models.ConsentRecord{}).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 rows after grant+withdraw, got %d", count)
	}
}

func TestStatus_FlagsReconsentWhenVersionDrifts(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	// Write a row claiming an OLD version (simulating a bump since).
	oldVersion := "0.9.0"
	rec := models.ConsentRecord{
		ConsentType:       "user_tc",
		SubjectIDHash:     repository.HashSubjectID(userID),
		DisplayedTextHash: liveHash(t, legal.DocUserTC, "en"),
		DisplayedLanguage: "en",
		DocumentVersion:   &oldVersion,
		Granted:           true,
		Channel:           "mobile",
	}
	if err := repo.Create(nil, &rec); err != nil {
		t.Fatalf("seed: %v", err)
	}

	resp, err := svc.Status(userID, []string{"user_tc"})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	entry, ok := resp.Consents["user_tc"]
	if !ok {
		t.Fatalf("user_tc missing from status response")
	}
	if !entry.NeedsReconsent {
		t.Error("expected needs_reconsent=true when stored version is older than live")
	}
}

func TestStatus_NeverGrantedMarkedNeedsReconsent(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	resp, err := svc.Status(userID, []string{"user_tc", "user_privacy"})
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	for _, k := range []string{"user_tc", "user_privacy"} {
		e := resp.Consents[k]
		if e.Granted {
			t.Errorf("%s: fresh user should not be marked granted", k)
		}
		if !e.NeedsReconsent {
			t.Errorf("%s: fresh user should need re-consent", k)
		}
	}
}

func TestHistory_ReturnsNewestFirst(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	// Three rows, deliberately out of chronological insertion order.
	subject := repository.HashSubjectID(userID)
	v := liveVersion(t, legal.DocUserTC)
	for i, offset := range []time.Duration{-2 * time.Hour, 0, -1 * time.Hour} {
		verStr := v
		rec := models.ConsentRecord{
			ConsentType:       "user_tc",
			SubjectIDHash:     subject,
			DisplayedTextHash: liveHash(t, legal.DocUserTC, "en"),
			DisplayedLanguage: "en",
			DocumentVersion:   &verStr,
			Granted:           true,
			Channel:           "mobile",
			GrantedAt:         time.Now().UTC().Add(offset),
		}
		if err := repo.Create(nil, &rec); err != nil {
			t.Fatalf("seed row %d: %v", i, err)
		}
	}

	resp, err := svc.History(userID)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(resp.Records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(resp.Records))
	}
	for i := 1; i < len(resp.Records); i++ {
		if resp.Records[i-1].GrantedAt.Before(resp.Records[i].GrantedAt) {
			t.Errorf("history not sorted newest first: idx %d (%v) < idx %d (%v)",
				i-1, resp.Records[i-1].GrantedAt, i, resp.Records[i].GrantedAt)
		}
	}
}

func TestPersistInTx_RollsBackIfAnyConsentInvalid(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	inputs := validSignupConsents(t, "en", "2000-01-15")
	// Corrupt one entry so verification fails
	inputs[1].DisplayedTextHash = strings.Repeat("f", 64)

	err := db.Transaction(func(tx *gorm.DB) error {
		_, err := svc.PersistInTx(tx, userID, inputs, "mobile")
		return err
	})
	if err == nil {
		t.Fatal("expected PersistInTx to fail due to corrupt row")
	}
	var count int64
	db.Model(&models.ConsentRecord{}).Count(&count)
	if count != 0 {
		t.Errorf("transaction should have rolled back; got %d rows persisted", count)
	}
}

func TestPersistInTx_HappyPath(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewConsentRepository(db)
	svc := NewConsentService(repo)

	userID := uuid.New()
	dob := "2000-01-15"
	inputs := validSignupConsents(t, "en", dob)

	err := db.Transaction(func(tx *gorm.DB) error {
		_, err := svc.PersistInTx(tx, userID, inputs, "mobile")
		return err
	})
	if err != nil {
		t.Fatalf("persist: %v", err)
	}

	var count int64
	db.Model(&models.ConsentRecord{}).Count(&count)
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}
