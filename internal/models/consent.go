package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConsentRecord is a PDPA-defensible, append-only audit entry proving
// what text (hash) the user saw, in which language, at which version,
// granted or withdrawn, when, and over which channel.
//
// Invariants enforced at the application layer:
//   - This table is INSERT-only. Repository code never issues UPDATE
//     or DELETE against existing rows.
//   - A "withdrawal" is a NEW row with granted=false and withdrawn=true.
//   - subject_id_hash = SHA-256(user.id). The raw user id is never
//     stored on this table so operator queries can't pivot back to PII
//     without a join to users (which has its own access controls).
//
// Columns match the web mockup's Postgres schema exactly
// (mockup/src/lib/consent/schema.ts) so the two apps produce one
// unified audit trail.
type ConsentRecord struct {
	// ConsentID is populated by BeforeCreate (Go side) rather than a
	// DB default so the schema stays dialect-agnostic (tests run on
	// SQLite; production is Postgres). Postgres is still a safe
	// backend because every insert goes through this hook.
	ConsentID uuid.UUID `gorm:"type:uuid;primaryKey" json:"consent_id"`

	ConsentType       string `gorm:"size:64;not null;index:idx_consent_subject,priority:2;index:idx_consent_type" json:"consent_type"`
	SubjectIDHash     string `gorm:"size:64;not null;index:idx_consent_subject,priority:1" json:"subject_id_hash"`
	DisplayedTextHash string `gorm:"size:64;not null" json:"displayed_text_hash"`
	DisplayedLanguage string `gorm:"size:8;not null" json:"displayed_language"`

	// DocumentVersion is nullable for consent types that aren't backed
	// by a versioned document (none today — all current types use it —
	// but future types like "cookie_banner_dismissal" might not).
	DocumentVersion *string `gorm:"size:32" json:"document_version,omitempty"`

	// Granted/GrantedAt: BeforeCreate stamps a default so SQLite tests
	// and ad-hoc raw inserts both get sensible values.
	Granted   bool      `gorm:"not null" json:"granted"`
	GrantedAt time.Time `gorm:"not null;index:idx_consent_subject,priority:3,sort:desc" json:"granted_at"`

	// Channel: 'web' | 'mobile' | 'admin' | 'support'. Required so the
	// DPO can tell where the record came from during audit.
	Channel string `gorm:"size:16;not null;default:'web'" json:"channel"`

	// Withdrawal fields. Set on the NEW row that represents the
	// withdrawal, NOT updated in place on the original grant row.
	Withdrawn   bool       `gorm:"not null;default:false" json:"withdrawn"`
	WithdrawnAt *time.Time `json:"withdrawn_at,omitempty"`

	// Optional defence-in-depth fingerprints. Hashed, not raw.
	IPHash        *string `gorm:"size:64" json:"ip_hash,omitempty"`
	UserAgentHash *string `gorm:"size:64" json:"user_agent_hash,omitempty"`
}

// TableName locks the Postgres table name to match the web mockup
// (default GORM pluralisation would give "consent_records" too, but
// being explicit avoids any surprise if someone customises the naming
// strategy later).
func (ConsentRecord) TableName() string {
	return "consent_records"
}

// BeforeCreate ensures ConsentID is populated when the DB driver
// doesn't execute the gen_random_uuid() default (e.g. SQLite in tests,
// or if the row was constructed in-memory without a prior INSERT).
// Postgres production still uses the DB default — this hook only
// fires if ConsentID is still the zero UUID at write time.
func (r *ConsentRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ConsentID == (uuid.UUID{}) {
		r.ConsentID = uuid.New()
	}
	if r.GrantedAt.IsZero() {
		r.GrantedAt = time.Now().UTC()
	}
	return nil
}
