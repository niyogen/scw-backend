// Dev helper: emits a valid /auth/register payload (consents + hashes
// derived from the live embedded bundle). Not shipped. Used by
// scripts/smoke_consent.sh and ad-hoc curl testing.
package main
import (
	"encoding/json"
	"fmt"
	"os"
	"delivery-backend/internal/legal"
)
func main() {
	dob := "2000-01-15"
	email := "smoke-test-2026@invalid.lk"
	phone := "0771234567"
	if len(os.Args) > 1 { email = os.Args[1] }
	if len(os.Args) > 2 { phone = os.Args[2] }
	tcHash, _, _ := legal.ComputeDisplayedTextHash(legal.DocUserTC, "en")
	privHash, _, _ := legal.ComputeDisplayedTextHash(legal.DocUserPrivacy, "en")
	ageHash := legal.ComputeAgeHash(dob)
	tcVer, _ := legal.Version(legal.DocUserTC)
	privVer, _ := legal.Version(legal.DocUserPrivacy)
	payload := map[string]interface{}{
		"name":             "Test SmokeUser",
		"email":            email,
		"phone":            phone,
		"password":         "password123",
		"confirm_password": "password123",
		"date_of_birth":    dob,
		"consents": []map[string]interface{}{
			{"consent_type":"user_tc","displayed_text_hash":tcHash,"displayed_language":"en","document_version":tcVer,"granted":true,"channel":"mobile"},
			{"consent_type":"user_privacy","displayed_text_hash":privHash,"displayed_language":"en","document_version":privVer,"granted":true,"channel":"mobile"},
			{"consent_type":"age_confirmation","displayed_text_hash":ageHash,"displayed_language":"en","document_version":legal.AgeVersion,"granted":true,"channel":"mobile"},
		},
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Fprintln(os.Stdout, string(b))
}
