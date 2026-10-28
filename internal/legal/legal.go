// Package legal loads the shared Terms & Privacy document bundle at build
// time and exposes the exact bytes the Flutter and web apps hash into the
// consent ledger. The source of truth lives in scw/legal_content/ and is
// mirrored into this package via scw/scripts/sync_legal_to_backend.sh. The
// hash function here must stay in sync with:
//
//   - web: mockup/src/lib/consent/hash.ts  (sha256(version + "\n" + body))
//   - mobile: lib/services/legal_docs_service.dart  (same formula)
//
// If any of those three drift, the audit trail stops lining up across
// platforms. A unit test verifies this on every build (see legal_test.go).
package legal

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed legal_content/*
var legalFS embed.FS

const (
	// Document IDs. Must match the keys in legal_content/config.json and
	// the consent_type values written to the consent_records table.
	DocUserTC      = "user_tc"
	DocUserPrivacy = "user_privacy"

	// Age-confirmation version. Not a document in config.json; it's a
	// separate hash salt matching web onboarding (confirmAge records
	// grants under this version, hashing "dob:YYYY-MM-DD").
	AgeVersion = "age/1.0.0"

	// AuthoritativeLang is the fallback when a requested locale is
	// missing a translation. Matches config.json's authoritative_language.
	AuthoritativeLang = "en"
)

type docMeta struct {
	ID            string            `json:"id"`
	Version       string            `json:"version"`
	EffectiveDate string            `json:"effective_date"`
	Paths         map[string]string `json:"paths"`
	Hashes        map[string]string `json:"hashes"`
}

type config struct {
	SchemaVersion         int                `json:"schema_version"`
	DataController        string             `json:"data_controller"`
	AppName               string             `json:"app_name"`
	AuthoritativeLanguage string             `json:"authoritative_language"`
	SupportedLanguages    []string           `json:"supported_languages"`
	PreLaunch             bool               `json:"pre_launch"`
	Documents             map[string]docMeta `json:"documents"`
}

type doc struct {
	id            string
	version       string
	effectiveDate string
	bodies        map[string]string // locale -> body
}

var (
	cfg         *config
	docs        map[string]*doc
	loadErr     error
	loadDoneErr error
)

func init() {
	loadErr = loadBundle()
}

func loadBundle() error {
	raw, err := legalFS.ReadFile("legal_content/config.json")
	if err != nil {
		return fmt.Errorf("legal: config.json missing from embed: %w", err)
	}
	var c config
	if err := json.Unmarshal(raw, &c); err != nil {
		return fmt.Errorf("legal: config.json malformed: %w", err)
	}
	loaded := make(map[string]*doc, len(c.Documents))
	for id, meta := range c.Documents {
		d := &doc{
			id:            id,
			version:       meta.Version,
			effectiveDate: meta.EffectiveDate,
			bodies:        make(map[string]string, len(meta.Paths)),
		}
		for lang, rel := range meta.Paths {
			body, err := legalFS.ReadFile("legal_content/" + rel)
			if err != nil {
				return fmt.Errorf("legal: missing body file %s for %s/%s: %w", rel, id, lang, err)
			}
			d.bodies[lang] = string(body)
		}
		loaded[id] = d
	}
	cfg = &c
	docs = loaded
	return nil
}

// LoadError returns any error that occurred during the embedded-bundle
// load at init time. Call this once at startup and fail-fast on non-nil.
func LoadError() error { return loadErr }

// IsPreLaunch reports whether the bundle is still marked as pre-launch
// in config.json. True means [TBD before launch] placeholders are still
// present; the admin UI should surface this prominently.
func IsPreLaunch() bool {
	if cfg == nil {
		return true
	}
	return cfg.PreLaunch
}

// DataController returns the controlling entity name from config.json
// (e.g. "Sunrise Holdings (Pvt) Ltd").
func DataController() string {
	if cfg == nil {
		return ""
	}
	return cfg.DataController
}

// AppName returns the user-facing product name from config.json.
func AppName() string {
	if cfg == nil {
		return ""
	}
	return cfg.AppName
}

// Version returns the version string for a document ID. Error if the
// document is unknown.
func Version(docID string) (string, error) {
	d, ok := docs[docID]
	if !ok {
		return "", fmt.Errorf("legal: unknown document %q", docID)
	}
	return d.version, nil
}

// EffectiveDate returns the effective date (YYYY-MM-DD) for a doc.
func EffectiveDate(docID string) (string, error) {
	d, ok := docs[docID]
	if !ok {
		return "", fmt.Errorf("legal: unknown document %q", docID)
	}
	return d.effectiveDate, nil
}

// Body returns the document body for the given locale, falling back
// to the authoritative language if the locale is missing a translation.
// The second return value is the locale actually served.
func Body(docID, locale string) (string, string, error) {
	d, ok := docs[docID]
	if !ok {
		return "", "", fmt.Errorf("legal: unknown document %q", docID)
	}
	if body, ok := d.bodies[locale]; ok {
		return body, locale, nil
	}
	if body, ok := d.bodies[AuthoritativeLang]; ok {
		return body, AuthoritativeLang, nil
	}
	return "", "", fmt.Errorf("legal: no body for %q (locale=%q)", docID, locale)
}

// ComputeDisplayedTextHash returns the SHA-256 hex of the document
// body the user would see for (docID, locale), salted with the
// document version. Must match the mobile/web hash formula exactly.
func ComputeDisplayedTextHash(docID, locale string) (string, string, error) {
	version, err := Version(docID)
	if err != nil {
		return "", "", err
	}
	body, servedLocale, err := Body(docID, locale)
	if err != nil {
		return "", "", err
	}
	return sha256Hex(version + "\n" + body), servedLocale, nil
}

// ComputeAgeHash hashes an age-confirmation marker. dob must be
// "YYYY-MM-DD". Matches web contract: sha256("age/1.0.0\ndob:" + dob).
func ComputeAgeHash(dob string) string {
	return sha256Hex(AgeVersion + "\ndob:" + strings.TrimSpace(dob))
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// SupportedLocales returns the locales the bundle has translations for,
// in config-declared order. Useful for request validation.
func SupportedLocales() []string {
	if cfg == nil {
		return nil
	}
	out := make([]string, len(cfg.SupportedLanguages))
	copy(out, cfg.SupportedLanguages)
	return out
}

// IsSupportedLocale checks whether a locale code is in the bundle.
func IsSupportedLocale(locale string) bool {
	for _, l := range SupportedLocales() {
		if l == locale {
			return true
		}
	}
	return false
}
