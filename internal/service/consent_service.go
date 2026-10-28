package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/legal"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConsentService handles grant / withdraw / status / history for the
// append-only consent ledger. Every write verifies the client-provided
// displayed_text_hash against what the server would compute from its
// own embedded copy of legal_content — a tampered or bugged client
// cannot persist a bogus hash.
type ConsentService interface {
	// Grant records a single consent. Idempotent on ConsentID.
	// Returns the persisted record (or the existing one on no-op).
	Grant(userID uuid.UUID, input dto.ConsentInput) (*models.ConsentRecord, error)

	// Withdraw records a withdrawal (new row with granted=false,
	// withdrawn=true, withdrawn_at=now). No ConsentID is accepted —
	// withdrawals always produce a fresh row.
	Withdraw(userID uuid.UUID, req dto.ConsentWithdrawRequest) (*models.ConsentRecord, error)

	// Status returns the current (latest-row) state for the given
	// consent types plus the server's current live versions for the
	// versioned documents, so clients can detect re-consent needs.
	Status(userID uuid.UUID, consentTypes []string) (*dto.ConsentStatusResponse, error)

	// History returns every row for the user, newest first. Used by
	// Settings → Privacy and DSR export.
	History(userID uuid.UUID) (*dto.ConsentHistoryResponse, error)

	// PersistInTx is used by AuthService.Register to insert all signup
	// consents in the same transaction as the user row. Expects the
	// subject's userID to already be assigned (post-Create).
	PersistInTx(tx *gorm.DB, userID uuid.UUID, inputs []dto.ConsentInput, channel string) ([]models.ConsentRecord, error)

	// VerifyAndNormalizeInput checks that the client's hash matches
	// what the server would compute, and that required fields are
	// consistent for the consent_type. Returns an error with a code
	// clients can switch on (e.g. "consent_validation_failed:hash_mismatch").
	VerifyAndNormalizeInput(input *dto.ConsentInput) error
}

type consentService struct {
	repo repository.ConsentRepository
}

func NewConsentService(repo repository.ConsentRepository) ConsentService {
	return &consentService{repo: repo}
}

// Known consent types this backend accepts. Keeping the enum small
// and explicit for v1; expanding later for KYC, biometric, cookies,
// marketing requires an entry here.
var allowedConsentTypes = map[string]struct{}{
	"user_tc":          {},
	"user_privacy":     {},
	"age_confirmation": {},
}

// consentTypeToDocID maps a consent_type to the document in the
// legal bundle whose hash must match. age_confirmation uses a
// synthetic salt (age/1.0.0 + "dob:YYYY-MM-DD") rather than a
// stored document.
var consentTypeToDocID = map[string]string{
	"user_tc":      legal.DocUserTC,
	"user_privacy": legal.DocUserPrivacy,
}

func (s *consentService) VerifyAndNormalizeInput(input *dto.ConsentInput) error {
	if input == nil {
		return errors.New("consent_validation_failed:nil_input")
	}
	input.ConsentType = strings.TrimSpace(input.ConsentType)
	input.DisplayedLanguage = strings.TrimSpace(input.DisplayedLanguage)
	input.DocumentVersion = strings.TrimSpace(input.DocumentVersion)
	input.DisplayedTextHash = strings.ToLower(strings.TrimSpace(input.DisplayedTextHash))

	if _, ok := allowedConsentTypes[input.ConsentType]; !ok {
		return fmt.Errorf("consent_validation_failed:unknown_consent_type:%s", input.ConsentType)
	}
	if !legal.IsSupportedLocale(input.DisplayedLanguage) {
		return fmt.Errorf("consent_validation_failed:unsupported_language:%s", input.DisplayedLanguage)
	}
	if input.DocumentVersion == "" {
		return errors.New("consent_validation_failed:missing_document_version")
	}
	if len(input.DisplayedTextHash) != 64 {
		return errors.New("consent_validation_failed:malformed_hash")
	}

	// Server-side hash verification. For versioned docs: compute from
	// the embedded body + version and compare. For age_confirmation:
	// we don't know the DOB here (the AuthService does) — caller
	// verifies it separately. Document version must match the live
	// bundle's version for the hash to possibly match.
	switch input.ConsentType {
	case "user_tc", "user_privacy":
		docID := consentTypeToDocID[input.ConsentType]
		liveVersion, err := legal.Version(docID)
		if err != nil {
			return fmt.Errorf("consent_validation_failed:legal_bundle_error:%v", err)
		}
		if input.DocumentVersion != liveVersion {
			return fmt.Errorf(
				"consent_validation_failed:version_drift:client=%s server=%s",
				input.DocumentVersion, liveVersion,
			)
		}
		wantHash, _, err := legal.ComputeDisplayedTextHash(docID, input.DisplayedLanguage)
		if err != nil {
			return fmt.Errorf("consent_validation_failed:hash_computation:%v", err)
		}
		if input.DisplayedTextHash != wantHash {
			return errors.New("consent_validation_failed:hash_mismatch")
		}
	case "age_confirmation":
		// Format sanity only; the authoritative check is in
		// AuthService.Register which hashes against the submitted DOB.
		if input.DocumentVersion != legal.AgeVersion {
			return fmt.Errorf(
				"consent_validation_failed:age_version_mismatch:client=%s server=%s",
				input.DocumentVersion, legal.AgeVersion,
			)
		}
	}
	return nil
}

func (s *consentService) toRecord(userID uuid.UUID, input dto.ConsentInput, channel string) models.ConsentRecord {
	ch := strings.TrimSpace(input.Channel)
	if ch == "" {
		ch = channel
	}
	if ch == "" {
		ch = "web"
	}
	grantedAt := time.Now().UTC()
	if input.GrantedAt != nil && !input.GrantedAt.IsZero() {
		grantedAt = input.GrantedAt.UTC()
	}
	rec := models.ConsentRecord{
		ConsentType:       input.ConsentType,
		SubjectIDHash:     repository.HashSubjectID(userID),
		DisplayedTextHash: input.DisplayedTextHash,
		DisplayedLanguage: input.DisplayedLanguage,
		Granted:           input.Granted,
		GrantedAt:         grantedAt,
		Channel:           ch,
	}
	if input.DocumentVersion != "" {
		v := input.DocumentVersion
		rec.DocumentVersion = &v
	}
	if input.ConsentID != nil {
		rec.ConsentID = *input.ConsentID
	}
	return rec
}

func (s *consentService) Grant(userID uuid.UUID, input dto.ConsentInput) (*models.ConsentRecord, error) {
	if err := s.VerifyAndNormalizeInput(&input); err != nil {
		return nil, err
	}
	// Idempotency: if ConsentID already exists, no-op.
	if input.ConsentID != nil {
		if existing, err := s.repo.FindByConsentID(*input.ConsentID); err == nil && existing != nil {
			return existing, nil
		}
	}
	record := s.toRecord(userID, input, "web")
	record.Granted = true
	if err := s.repo.Create(nil, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *consentService) Withdraw(userID uuid.UUID, req dto.ConsentWithdrawRequest) (*models.ConsentRecord, error) {
	consentType := strings.TrimSpace(req.ConsentType)
	if _, ok := allowedConsentTypes[consentType]; !ok {
		return nil, fmt.Errorf("consent_validation_failed:unknown_consent_type:%s", consentType)
	}
	// Build a new append-only row representing the withdrawal. We do
	// not require a hash on withdrawal — the user is revoking consent
	// on whatever text they previously saw, not re-affirming it.
	now := time.Now().UTC()
	version := strings.TrimSpace(req.DocumentVersion)
	rec := models.ConsentRecord{
		ConsentType:       consentType,
		SubjectIDHash:     repository.HashSubjectID(userID),
		DisplayedTextHash: "withdrawal:" + consentType, // sentinel; non-hash
		DisplayedLanguage: strings.TrimSpace(req.DisplayedLanguage),
		Granted:           false,
		GrantedAt:         now,
		Channel:           "web", // overridden by handler if from mobile
		Withdrawn:         true,
		WithdrawnAt:       &now,
	}
	if version != "" {
		rec.DocumentVersion = &version
	}
	// Pad sentinel to 64 chars so column size holds (defensive).
	rec.DisplayedTextHash = strings.Repeat("w", 64-len(rec.DisplayedTextHash))[:64-len(rec.DisplayedTextHash)] + rec.DisplayedTextHash
	if err := s.repo.Create(nil, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *consentService) Status(userID uuid.UUID, consentTypes []string) (*dto.ConsentStatusResponse, error) {
	if len(consentTypes) == 0 {
		consentTypes = []string{"user_tc", "user_privacy", "age_confirmation"}
	}
	subject := repository.HashSubjectID(userID)
	latest, err := s.repo.LatestForSubjectMulti(subject, consentTypes)
	if err != nil {
		return nil, err
	}
	currentVersions := map[string]string{}
	for _, docID := range []string{legal.DocUserTC, legal.DocUserPrivacy} {
		if v, err := legal.Version(docID); err == nil {
			currentVersions[docID] = v
		}
	}
	resp := &dto.ConsentStatusResponse{
		Consents:        make(map[string]dto.ConsentStatusEntry, len(consentTypes)),
		CurrentVersions: currentVersions,
	}
	for _, ct := range consentTypes {
		entry := dto.ConsentStatusEntry{ConsentType: ct}
		row, ok := latest[ct]
		if !ok || row == nil {
			entry.Granted = false
			entry.NeedsReconsent = true
			resp.Consents[ct] = entry
			continue
		}
		entry.Granted = row.Granted && !row.Withdrawn
		entry.DocumentVersion = row.DocumentVersion
		entry.DisplayedLanguage = row.DisplayedLanguage
		gt := row.GrantedAt
		entry.GrantedAt = &gt
		if row.Withdrawn && row.WithdrawnAt != nil {
			entry.WithdrawnAt = row.WithdrawnAt
		}
		// Needs re-consent iff: not currently granted OR the version
		// on the latest grant is older than the live version. The
		// client decides whether to force the UI flow.
		if !entry.Granted {
			entry.NeedsReconsent = true
		} else if liveV, ok := currentVersions[ct]; ok {
			if row.DocumentVersion == nil || *row.DocumentVersion != liveV {
				entry.NeedsReconsent = true
			}
		}
		resp.Consents[ct] = entry
	}
	return resp, nil
}

func (s *consentService) History(userID uuid.UUID) (*dto.ConsentHistoryResponse, error) {
	subject := repository.HashSubjectID(userID)
	rows, err := s.repo.HistoryForSubject(subject)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ConsentRecordResponse, 0, len(rows))
	for i := range rows {
		r := rows[i]
		out = append(out, dto.ConsentRecordResponse{
			ConsentID:         r.ConsentID,
			ConsentType:       r.ConsentType,
			DisplayedTextHash: r.DisplayedTextHash,
			DisplayedLanguage: r.DisplayedLanguage,
			DocumentVersion:   r.DocumentVersion,
			Granted:           r.Granted,
			GrantedAt:         r.GrantedAt,
			Channel:           r.Channel,
			Withdrawn:         r.Withdrawn,
			WithdrawnAt:       r.WithdrawnAt,
		})
	}
	return &dto.ConsentHistoryResponse{Records: out}, nil
}

func (s *consentService) PersistInTx(tx *gorm.DB, userID uuid.UUID, inputs []dto.ConsentInput, channel string) ([]models.ConsentRecord, error) {
	if tx == nil {
		return nil, errors.New("PersistInTx requires an active transaction")
	}
	if len(inputs) == 0 {
		return nil, nil
	}
	records := make([]models.ConsentRecord, 0, len(inputs))
	for i := range inputs {
		in := inputs[i]
		if err := s.VerifyAndNormalizeInput(&in); err != nil {
			return nil, err
		}
		// Idempotency check — if a client retries register with the
		// same consent IDs, don't double-insert.
		if in.ConsentID != nil {
			if existing, err := s.repo.FindByConsentID(*in.ConsentID); err == nil && existing != nil {
				records = append(records, *existing)
				continue
			}
		}
		rec := s.toRecord(userID, in, channel)
		rec.Granted = true
		if err := s.repo.Create(tx, &rec); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, nil
}
