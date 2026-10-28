package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConsentRepository provides append-only writes and read queries against
// the consent_records table. It never issues UPDATE or DELETE against
// existing rows — withdrawals are new rows, not in-place changes.
type ConsentRepository interface {
	// Create inserts a new row verbatim. The caller is responsible for
	// supplying all fields correctly (including subject_id_hash).
	Create(tx *gorm.DB, record *models.ConsentRecord) error

	// CreateMany inserts N rows in one SQL statement. Fails the whole
	// batch if any row violates a constraint.
	CreateMany(tx *gorm.DB, records []models.ConsentRecord) error

	// LatestForSubject returns the most recent row for a given user
	// and consent type. Returns (nil, nil) if the user has never
	// granted or withdrawn this type.
	LatestForSubject(subjectIDHash, consentType string) (*models.ConsentRecord, error)

	// LatestForSubjectMulti returns the most recent row per consent
	// type for the given user in one query.
	LatestForSubjectMulti(subjectIDHash string, consentTypes []string) (map[string]*models.ConsentRecord, error)

	// HistoryForSubject returns every row for a user in descending
	// granted_at order. Used by Settings → Privacy and by DSR export.
	HistoryForSubject(subjectIDHash string) ([]models.ConsentRecord, error)

	// FindByConsentID returns the row with the given ConsentID, or
	// nil if none exists. Used for idempotency checks.
	FindByConsentID(consentID uuid.UUID) (*models.ConsentRecord, error)

	// DB returns the underlying gorm.DB so services can start their
	// own transactions when they need to combine consent writes with
	// other writes (e.g. user creation at register time).
	DB() *gorm.DB
}

type consentRepository struct {
	db *gorm.DB
}

func NewConsentRepository(db *gorm.DB) ConsentRepository {
	return &consentRepository{db: db}
}

func (r *consentRepository) DB() *gorm.DB { return r.db }

func (r *consentRepository) Create(tx *gorm.DB, record *models.ConsentRecord) error {
	if tx == nil {
		tx = r.db
	}
	if record == nil {
		return errors.New("consent record is nil")
	}
	if record.GrantedAt.IsZero() {
		record.GrantedAt = time.Now().UTC()
	}
	return tx.Create(record).Error
}

func (r *consentRepository) CreateMany(tx *gorm.DB, records []models.ConsentRecord) error {
	if tx == nil {
		tx = r.db
	}
	if len(records) == 0 {
		return nil
	}
	now := time.Now().UTC()
	for i := range records {
		if records[i].GrantedAt.IsZero() {
			records[i].GrantedAt = now
		}
	}
	return tx.Create(&records).Error
}

func (r *consentRepository) LatestForSubject(subjectIDHash, consentType string) (*models.ConsentRecord, error) {
	var row models.ConsentRecord
	err := r.db.
		Where("subject_id_hash = ? AND consent_type = ?", subjectIDHash, consentType).
		Order("granted_at DESC").
		Limit(1).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *consentRepository) LatestForSubjectMulti(subjectIDHash string, consentTypes []string) (map[string]*models.ConsentRecord, error) {
	out := make(map[string]*models.ConsentRecord, len(consentTypes))
	if len(consentTypes) == 0 {
		return out, nil
	}
	// Dialect-agnostic approach: for each type, fetch the latest row.
	// v1 scale makes this cheap (handful of types per status request,
	// each query hits the (subject_id_hash, consent_type, granted_at)
	// index). If N grows past ~10 we can swap to a Postgres-specific
	// DISTINCT ON path at that point.
	for _, ct := range consentTypes {
		row, err := r.LatestForSubject(subjectIDHash, ct)
		if err != nil {
			return nil, err
		}
		if row != nil {
			out[ct] = row
		}
	}
	return out, nil
}

func (r *consentRepository) HistoryForSubject(subjectIDHash string) ([]models.ConsentRecord, error) {
	var rows []models.ConsentRecord
	err := r.db.
		Where("subject_id_hash = ?", subjectIDHash).
		Order("granted_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *consentRepository) FindByConsentID(consentID uuid.UUID) (*models.ConsentRecord, error) {
	var row models.ConsentRecord
	err := r.db.Where("consent_id = ?", consentID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// HashSubjectID returns SHA-256(user.id.String()) as hex. Keeps raw
// user IDs off the consent_records table so operator queries can't
// pivot to PII without a users-table join (which has stricter access
// control). Package-level helper so services and tests share one
// implementation.
func HashSubjectID(userID uuid.UUID) string {
	sum := sha256.Sum256([]byte(userID.String()))
	return hex.EncodeToString(sum[:])
}
