package dto

import (
	"time"

	"github.com/google/uuid"
)

// ConsentInput is the shape the Flutter mobile app and web app send
// when capturing a consent (signup, re-consent, settings re-grant).
// All fields are validated server-side; the server also independently
// recomputes DisplayedTextHash from its embedded legal_content copy
// and rejects mismatches.
type ConsentInput struct {
	// ConsentID is a client-generated UUID used as an idempotency key.
	// Replaying a grant with the same ConsentID is a no-op, letting
	// the mobile local ledger safely replay without duplicating.
	ConsentID *uuid.UUID `json:"consent_id"`

	ConsentType       string `json:"consent_type" binding:"required"`
	DisplayedTextHash string `json:"displayed_text_hash" binding:"required,len=64"`
	DisplayedLanguage string `json:"displayed_language" binding:"required,oneof=en si ta"`
	DocumentVersion   string `json:"document_version" binding:"required"`

	Granted bool `json:"granted"`

	// GrantedAt may be past-dated when replayed from a local ledger.
	// If zero, the server stamps now().
	GrantedAt *time.Time `json:"granted_at"`

	// Channel: web | mobile | admin | support. Defaults to "mobile"
	// if the request comes from a mobile client, "web" otherwise.
	Channel string `json:"channel"`
}

// ConsentGrantRequest is the standalone endpoint shape for recording
// a consent outside of /auth/register (re-consent on version bump,
// local-ledger replay, settings re-grant).
type ConsentGrantRequest struct {
	ConsentInput
}

// ConsentWithdrawRequest is the shape for recording a withdrawal.
// A withdrawal creates a NEW append-only row with granted=false and
// withdrawn=true. The user's subject_id_hash is derived from the
// authenticated session on the server.
type ConsentWithdrawRequest struct {
	ConsentType       string `json:"consent_type" binding:"required"`
	DisplayedLanguage string `json:"displayed_language" binding:"required,oneof=en si ta"`
	DocumentVersion   string `json:"document_version" binding:"required"`
}

// ConsentStatusEntry is one row of ConsentStatusResponse.
type ConsentStatusEntry struct {
	ConsentType       string     `json:"consent_type"`
	Granted           bool       `json:"granted"`
	DocumentVersion   *string    `json:"document_version,omitempty"`
	DisplayedLanguage string     `json:"displayed_language,omitempty"`
	GrantedAt         *time.Time `json:"granted_at,omitempty"`
	WithdrawnAt       *time.Time `json:"withdrawn_at,omitempty"`
	NeedsReconsent    bool       `json:"needs_reconsent"`
}

// ConsentStatusResponse tells the client the current grant state for
// each consent type requested, plus the server's current live version
// for the two main versioned documents so the client can decide
// whether to force a re-consent cascade.
type ConsentStatusResponse struct {
	Consents        map[string]ConsentStatusEntry `json:"consents"`
	CurrentVersions map[string]string             `json:"current_versions"`
}

// ConsentHistoryResponse returns the full append-only history for a
// user. Used by Settings → Privacy and by DSR export.
type ConsentHistoryResponse struct {
	Records []ConsentRecordResponse `json:"records"`
}

// ConsentRecordResponse mirrors models.ConsentRecord minus the
// subject_id_hash (which the user already knows is theirs) and
// the IP/UA hashes (defence-in-depth, not useful to show).
type ConsentRecordResponse struct {
	ConsentID         uuid.UUID  `json:"consent_id"`
	ConsentType       string     `json:"consent_type"`
	DisplayedTextHash string     `json:"displayed_text_hash"`
	DisplayedLanguage string     `json:"displayed_language"`
	DocumentVersion   *string    `json:"document_version,omitempty"`
	Granted           bool       `json:"granted"`
	GrantedAt         time.Time  `json:"granted_at"`
	Channel           string     `json:"channel"`
	Withdrawn         bool       `json:"withdrawn"`
	WithdrawnAt       *time.Time `json:"withdrawn_at,omitempty"`
}
