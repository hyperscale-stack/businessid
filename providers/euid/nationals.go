// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid

import (
	"strings"

	"github.com/hyperscale-stack/businessid"
)

// registerValidator holds the native rules for one country's national
// business register. All fields are optional — a nil field means "no
// national-level check at that step, treat as valid".
type registerValidator struct {
	// canonicalize normalises the REGISTRATION segment before format /
	// checksum checks (strip separators, apply casing). If nil the
	// registration is used as-is.
	canonicalize func(registration string) string

	// validateFormat checks the layout of the (canonicalized) REGISTRATION
	// segment. Returns (ok, reason, message) where reason is one of
	// businessid.ReasonInvalid*.
	validateFormat func(registration string) (bool, string, string)

	// validateChecksum verifies the national checksum on the canonicalized
	// REGISTRATION segment. nil means the register has no publicly
	// documented checksum.
	validateChecksum func(registration string) bool
}

// NewRegisterValidator constructs a register-validator from user-provided
// functions. Pass nil for any step that should be skipped. This is the
// public entry point used by [WithCountryValidator] to inject validators
// for non-EU codes or overrides.
func NewRegisterValidator(
	canonicalize func(registration string) string,
	validateFormat func(registration string) (bool, string, string),
	validateChecksum func(registration string) bool,
) registerValidator { //nolint:revive // deliberate: rv is opaque to callers
	return registerValidator{
		canonicalize:     canonicalize,
		validateFormat:   validateFormat,
		validateChecksum: validateChecksum,
	}
}

// euidRegisterValidators binds each EU country to the native rules for
// its national business register. Sourced from BRIS (Regulation (EU)
// 2015/884), the national commercial register documentation, and
// cross-verified against the python-stdnum library where applicable.
//
// Non-EU codes (XI, GB, NO, IS, LI) are absent: BRIS does not cover them.
// Callers who need to validate EUIDs with those prefixes can inject
// custom validators via [WithCountryValidator].
var euidRegisterValidators = map[string]registerValidator{
	"AT": {canonicalize: stripSeparators, validateFormat: atRegisterFormat},                                       // Firmenbuchnummer: 1-6 digits + 1 letter
	"BE": {canonicalize: stripSeparators, validateFormat: beRegisterFormat, validateChecksum: beRegisterChecksum}, // KBO: 10 digits + mod-97
	"BG": {canonicalize: stripSeparators, validateFormat: bgRegisterFormat, validateChecksum: bgRegisterChecksum}, // EIK: 9 or 13 digits
	"HR": {canonicalize: stripSeparators, validateFormat: hrRegisterFormat},                                       // MBS: 8 digits
	"CY": {canonicalize: stripSeparators, validateFormat: cyRegisterFormat},                                       // HE number: 6 digits
	"CZ": {canonicalize: stripSeparators, validateFormat: czRegisterFormat, validateChecksum: czRegisterChecksum}, // IČO: 8 digits + mod 11
	"DE": {validateFormat: deRegisterFormat},                                                                      // Handelsregister: local court + type + number
	"DK": {canonicalize: stripSeparators, validateFormat: dkRegisterFormat, validateChecksum: dkRegisterChecksum}, // CVR: 8 digits + mod 11
	"EE": {canonicalize: stripSeparators, validateFormat: eeRegisterFormat, validateChecksum: eeRegisterChecksum}, // Registrikood: 8 digits + mod 11
	"EL": {canonicalize: stripSeparators, validateFormat: elRegisterFormat},                                       // GEMI: 12 digits
	"ES": {canonicalize: stripSeparators, validateFormat: esRegisterFormat, validateChecksum: esRegisterChecksum}, // NIF/CIF
	"FI": {canonicalize: stripSeparators, validateFormat: fiRegisterFormat, validateChecksum: fiRegisterChecksum}, // Y-tunnus: 7 digits + check
	"FR": {canonicalize: stripSeparators, validateFormat: frRegisterFormat, validateChecksum: frRegisterChecksum}, // SIREN: 9 digits + Luhn
	"HU": {validateFormat: huRegisterFormat},                                                                      // Cégjegyzékszám: NN-NN-NNNNNN
	"IE": {canonicalize: stripSeparators, validateFormat: ieRegisterFormat},                                       // CRO: 5-7 digits
	"IT": {canonicalize: stripSeparators, validateFormat: itRegisterFormat, validateChecksum: itRegisterChecksum}, // Codice Fiscale entità: 11 digits + Luhn
	"LT": {canonicalize: stripSeparators, validateFormat: ltRegisterFormat, validateChecksum: ltRegisterChecksum}, // 9 digits + mod 11
	"LU": {canonicalize: stripSeparators, validateFormat: luRegisterFormat},                                       // RCSL: B + digits
	"LV": {canonicalize: stripSeparators, validateFormat: lvRegisterFormat, validateChecksum: lvRegisterChecksum}, // 11 digits
	"MT": {canonicalize: stripSeparators, validateFormat: mtRegisterFormat},                                       // C + digits
	"NL": {canonicalize: stripSeparators, validateFormat: nlRegisterFormat},                                       // KVK: 8 digits (no publicly documented checksum)
	"PL": {canonicalize: stripSeparators, validateFormat: plRegisterFormat},                                       // KRS: 10 digits
	"PT": {canonicalize: stripSeparators, validateFormat: ptRegisterFormat, validateChecksum: ptRegisterChecksum}, // NIPC: 9 digits + mod 11
	"RO": {canonicalize: stripSeparators, validateFormat: roRegisterFormat, validateChecksum: roRegisterChecksum}, // CUI: 2-10 digits + mod 11
	"SE": {canonicalize: stripSeparators, validateFormat: seRegisterFormat, validateChecksum: seRegisterChecksum}, // Organisationsnummer: 10 digits + Luhn
	"SI": {canonicalize: stripSeparators, validateFormat: siRegisterFormat},                                       // Matična številka: 7 digits
	"SK": {canonicalize: stripSeparators, validateFormat: skRegisterFormat, validateChecksum: skRegisterChecksum}, // IČO: 8 digits + mod 11
}

// stripSeparators removes spaces, dashes and dots from a registration
// segment before format / checksum checks. Countries whose registration
// is a simple digit or letter run use this canonicalizer.
func stripSeparators(s string) string {
	return businessid.StripSeparators(businessid.StripAllSpaces(s), ".", "-", "/")
}

// ----------------------------------------------------------------------
// AT — Firmenbuchnummer (justiz.gv.at/firmenbuch).
// Format: 1..6 digits + 1 upper-case letter (e.g. "123456A", "78J").
// No publicly documented checksum for the letter suffix.
// ----------------------------------------------------------------------

func atRegisterFormat(r string) (bool, string, string) {
	if len(r) < 2 || len(r) > 7 {
		return false, businessid.ReasonInvalidLength, "AT Firmenbuchnummer must be 2-7 chars"
	}

	digits := r[:len(r)-1]
	letter := r[len(r)-1]

	if !businessid.IsAllDigits(digits) || letter < 'A' || letter > 'Z' {
		return false, businessid.ReasonInvalidCharacters, "AT Firmenbuchnummer must be digits + one letter"
	}

	return true, businessid.ReasonOK, ""
}

// ----------------------------------------------------------------------
// BE — BCE/KBO (kbopub.economie.fgov.be).
// 10 digits, first is 0 or 1. Checksum: 97 - (first_8 mod 97) == last_2.
// ----------------------------------------------------------------------

func beRegisterFormat(r string) (bool, string, string) {
	if len(r) != 10 {
		return false, businessid.ReasonInvalidLength, "BE KBO number must be 10 digits"
	}

	if !businessid.IsAllDigits(r) || (r[0] != '0' && r[0] != '1') {
		return false, businessid.ReasonInvalidCharacters, "BE KBO number must be 10 digits starting with 0 or 1"
	}

	return true, businessid.ReasonOK, ""
}

func beRegisterChecksum(r string) bool {
	first8 := parseInt(r[:8])
	last2 := parseInt(r[8:])

	return 97-first8%97 == last2
}

// ----------------------------------------------------------------------
// BG — EIK / ЕИК (brra.bg). 9 digits (legal entity) or 13 digits (branch:
// 9-digit parent + 4-digit branch serial). Checksum on 9-digit parent
// only, mirrors VAT BG BULSTAT.
// ----------------------------------------------------------------------

func bgRegisterFormat(r string) (bool, string, string) {
	if len(r) != 9 && len(r) != 13 {
		return false, businessid.ReasonInvalidLength, "BG EIK must be 9 or 13 digits"
	}

	if !businessid.IsAllDigits(r) {
		return false, businessid.ReasonInvalidCharacters, "BG EIK must be all digits"
	}

	return true, businessid.ReasonOK, ""
}

func bgRegisterChecksum(r string) bool { return bgBulstat9Checksum(r[:9]) }

func bgBulstat9Checksum(body string) bool {
	weights := [8]int{1, 2, 3, 4, 5, 6, 7, 8}
	sum := 0

	for i := range 8 {
		sum += int(body[i]-'0') * weights[i]
	}

	check := sum % 11
	if check == 10 {
		altWeights := [8]int{3, 4, 5, 6, 7, 8, 9, 10}
		sum2 := 0

		for i := range 8 {
			sum2 += int(body[i]-'0') * altWeights[i]
		}

		check = sum2 % 11
		if check == 10 {
			check = 0
		}
	}

	return check == int(body[8]-'0')
}

// ----------------------------------------------------------------------
// HR — MBS (sudreg.pravosudje.hr). 8 digits, no publicly documented checksum.
// ----------------------------------------------------------------------

func hrRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 8, "HR MBS") }

// ----------------------------------------------------------------------
// CY — HE number (efiling.drcor.mcit.gov.cy). 6 digits (leading zeros
// significant). No standardized checksum.
// ----------------------------------------------------------------------

func cyRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 6, "CY HE number")
}

// ----------------------------------------------------------------------
// CZ — IČO (or.justice.cz). 8 digits + weighted mod-11 checksum.
// Source: Zákon č. 227/2000 Sb. and python-stdnum stdnum.cz.dic.
// ----------------------------------------------------------------------

func czRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 8, "CZ IČO") }

func czRegisterChecksum(r string) bool { return icoMod11(r) }

// icoMod11 implements the shared IČO / IČ DPH mod-11 checksum used by
// CZ and SK 8-digit business identifiers.
func icoMod11(body string) bool {
	weights := [7]int{8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := range 7 {
		sum += int(body[i]-'0') * weights[i]
	}

	r := sum % 11

	var check int

	switch r {
	case 0:
		check = 1
	case 1:
		check = 0
	case 10:
		check = 1
	default:
		check = 11 - r
	}

	return check == int(body[7]-'0')
}

// ----------------------------------------------------------------------
// DE — Handelsregister (unternehmensregister.de). The registration
// segment carries the local court, register type (HRA/HRB/GnR/VR/PR)
// and entry number, separated by "/" and "-". Format is loose because
// BRIS allows a broad free-form value; we require at least one
// alphanumeric character.
// ----------------------------------------------------------------------

func deRegisterFormat(r string) (bool, string, string) {
	hasAlnum := false

	for i := range len(r) {
		c := r[i]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			hasAlnum = true

			break
		}
	}

	if !hasAlnum {
		return false, businessid.ReasonInvalidCharacters, "DE Handelsregister must contain at least one alphanumeric character"
	}

	return true, businessid.ReasonOK, ""
}

// ----------------------------------------------------------------------
// DK — CVR (cvr.dk). 8 digits, first ≠ 0. Weighted mod-11.
// ----------------------------------------------------------------------

func dkRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 8, "DK CVR") }

func dkRegisterChecksum(r string) bool {
	if r[0] == '0' {
		return false
	}

	weights := [8]int{2, 7, 6, 5, 4, 3, 2, 1}
	sum := 0

	for i := range 8 {
		sum += int(r[i]-'0') * weights[i]
	}

	return sum%11 == 0
}

// ----------------------------------------------------------------------
// EE — Registrikood (ariregister.rik.ee). 8 digits, weighted mod-11.
// ----------------------------------------------------------------------

func eeRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 8, "EE Registrikood")
}

func eeRegisterChecksum(r string) bool {
	weights := [7]int{1, 2, 3, 4, 5, 6, 7}
	sum := 0

	for i := range 7 {
		sum += int(r[i]-'0') * weights[i]
	}

	check := sum % 11
	if check == 10 {
		alt := [7]int{3, 4, 5, 6, 7, 8, 9}
		sum2 := 0

		for i := range 7 {
			sum2 += int(r[i]-'0') * alt[i]
		}

		check = sum2 % 11 % 10
	}

	return check == int(r[7]-'0')
}

// ----------------------------------------------------------------------
// EL — GEMI (businessregistry.gr). 12 digits, no standardized checksum.
// ----------------------------------------------------------------------

func elRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 12, "EL GEMI") }

// ----------------------------------------------------------------------
// ES — NIF/CIF (agenciatributaria.es). Same format as VAT ES CIF.
// ----------------------------------------------------------------------

func esRegisterFormat(r string) (bool, string, string) {
	if len(r) != 9 {
		return false, businessid.ReasonInvalidLength, "ES NIF/CIF must be 9 characters"
	}

	if !isAlnumUpperByte(r[0]) || !businessid.IsAllDigits(r[1:8]) || !isAlnumUpperByte(r[8]) {
		return false, businessid.ReasonInvalidCharacters, "ES NIF/CIF must be alnum + 7 digits + alnum"
	}

	return true, businessid.ReasonOK, ""
}

func esRegisterChecksum(r string) bool { return esCIFOrDNIChecksum(r) }

// ----------------------------------------------------------------------
// FI — Y-tunnus (prh.fi). 7 digits + 1 check digit (may be joined by "-").
// Weighted mod-11: 7-9-10-5-8-4-2.
// ----------------------------------------------------------------------

func fiRegisterFormat(r string) (bool, string, string) {
	if len(r) != 8 {
		return false, businessid.ReasonInvalidLength, "FI Y-tunnus must be 8 digits"
	}

	if !businessid.IsAllDigits(r) {
		return false, businessid.ReasonInvalidCharacters, "FI Y-tunnus must be all digits"
	}

	return true, businessid.ReasonOK, ""
}

func fiRegisterChecksum(r string) bool {
	weights := [7]int{7, 9, 10, 5, 8, 4, 2}
	sum := 0

	for i := range 7 {
		sum += int(r[i]-'0') * weights[i]
	}

	m := sum % 11
	if m == 1 {
		return false
	}

	check := 0
	if m > 1 {
		check = 11 - m
	}

	return check == int(r[7]-'0')
}

// ----------------------------------------------------------------------
// FR — SIREN (infogreffe.fr). 9 digits + Luhn.
// ----------------------------------------------------------------------

func frRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 9, "FR SIREN") }

func frRegisterChecksum(r string) bool { return businessid.Luhn(r) }

// ----------------------------------------------------------------------
// HU — Cégjegyzékszám (e-cegjegyzek.hu). Format: 2 digits (region) + "-"
// + 2 digits (type) + "-" + 6 digits (serial). We accept both with and
// without dashes (BRIS transmits often un-hyphenated).
// ----------------------------------------------------------------------

func huRegisterFormat(r string) (bool, string, string) {
	compact := stripSeparators(r)
	if len(compact) != 10 {
		return false, businessid.ReasonInvalidLength, "HU Cégjegyzékszám must be 10 digits total (RR-TT-NNNNNN)"
	}

	if !businessid.IsAllDigits(compact) {
		return false, businessid.ReasonInvalidCharacters, "HU Cégjegyzékszám must be digits (and optional dashes)"
	}

	return true, businessid.ReasonOK, ""
}

// ----------------------------------------------------------------------
// IE — CRO number (cro.ie). 5-7 digits (historical range).
// ----------------------------------------------------------------------

func ieRegisterFormat(r string) (bool, string, string) {
	if len(r) < 5 || len(r) > 7 {
		return false, businessid.ReasonInvalidLength, "IE CRO number must be 5-7 digits"
	}

	if !businessid.IsAllDigits(r) {
		return false, businessid.ReasonInvalidCharacters, "IE CRO number must be all digits"
	}

	return true, businessid.ReasonOK, ""
}

// ----------------------------------------------------------------------
// IT — Codice Fiscale entità (registroimprese.it). 11 digits + Luhn.
// ----------------------------------------------------------------------

func itRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 11, "IT Codice Fiscale")
}

func itRegisterChecksum(r string) bool { return businessid.Luhn(r) }

// ----------------------------------------------------------------------
// LT — Juridinio asmens kodas (registrucentras.lt). 9 digits, mod 11.
// ----------------------------------------------------------------------

func ltRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 9, "LT juridinio asmens kodas")
}

func ltRegisterChecksum(r string) bool {
	weights1 := [9]int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	sum := 0

	for i := range 9 {
		sum += int(r[i]-'0') * weights1[i]
	}

	m := sum % 11
	if m == 10 {
		weights2 := [9]int{3, 4, 5, 6, 7, 8, 9, 1, 2}
		sum2 := 0

		for i := range 9 {
			sum2 += int(r[i]-'0') * weights2[i]
		}

		m = sum2 % 11 % 10
	}

	return m == 0
}

// ----------------------------------------------------------------------
// LU — RCS Luxembourg (lbr.lu). Format: "B" + 4-6 digits.
// ----------------------------------------------------------------------

func luRegisterFormat(r string) (bool, string, string) {
	if len(r) < 5 || len(r) > 7 {
		return false, businessid.ReasonInvalidLength, "LU RCSL must be B + 4-6 digits"
	}

	if r[0] != 'B' || !businessid.IsAllDigits(r[1:]) {
		return false, businessid.ReasonInvalidCharacters, "LU RCSL must be B followed by digits"
	}

	return true, businessid.ReasonOK, ""
}

// ----------------------------------------------------------------------
// LV — Reģistrācijas numurs (ur.gov.lv). 11 digits, mod-11 weighted.
// ----------------------------------------------------------------------

func lvRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 11, "LV registrācijas numurs")
}

func lvRegisterChecksum(r string) bool {
	if r[0] > '3' {
		// Legal entity — mirrors VAT LV algorithm.
		weights := [10]int{9, 1, 4, 8, 3, 10, 2, 5, 7, 6}
		sum := 0

		for i := range 10 {
			sum += int(r[i]-'0') * weights[i]
		}

		check := ((3-sum)%11 + 11) % 11
		if check == 10 {
			return false
		}

		return check == int(r[10]-'0')
	}

	// Natural person — personal code with DDMMYY + century + serial + check.
	// Algorithm: weights = [10,5,8,4,2,1,6,3,7,9], check = (1101 - S) mod 11 mod 10.
	weights := [10]int{10, 5, 8, 4, 2, 1, 6, 3, 7, 9}
	sum := 0

	for i := range 10 {
		sum += int(r[i]-'0') * weights[i]
	}

	check := (1101 - sum) % 11 % 10

	return check == int(r[10]-'0')
}

// ----------------------------------------------------------------------
// MT — Company Registration Number (mbr.mt). Format: "C" + 4-6 digits.
// ----------------------------------------------------------------------

func mtRegisterFormat(r string) (bool, string, string) {
	if len(r) < 5 || len(r) > 7 {
		return false, businessid.ReasonInvalidLength, "MT company number must be C + 4-6 digits"
	}

	if r[0] != 'C' || !businessid.IsAllDigits(r[1:]) {
		return false, businessid.ReasonInvalidCharacters, "MT company number must be C followed by digits"
	}

	return true, businessid.ReasonOK, ""
}

// ----------------------------------------------------------------------
// NL — KVK-nummer (kvk.nl). 8 digits. The KVK number is a sequential
// register identifier with no publicly documented checksum algorithm
// (verified against real KVKs of Philips, ASML, Shell, Unilever,
// Rabobank, Heineken — none pass a common mod-11 formula), so we keep
// this register format-only.
// ----------------------------------------------------------------------

func nlRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 8, "NL KVK nummer")
}

// ----------------------------------------------------------------------
// PL — KRS (krs.ms.gov.pl). 10 digits, leading zeros significant.
// No publicly documented checksum for KRS itself.
// ----------------------------------------------------------------------

func plRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 10, "PL KRS") }

// ----------------------------------------------------------------------
// PT — NIPC (portaldocidadao.pt). 9 digits, mod-11 weighted.
// ----------------------------------------------------------------------

func ptRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 9, "PT NIPC") }

func ptRegisterChecksum(r string) bool {
	weights := [8]int{9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := range 8 {
		sum += int(r[i]-'0') * weights[i]
	}

	m := sum % 11
	check := 0

	if m >= 2 {
		check = 11 - m
	}

	return check == int(r[8]-'0')
}

// ----------------------------------------------------------------------
// RO — CUI (onrc.ro). 2-10 digits, mod-11 weighted with right-aligned weights.
// ----------------------------------------------------------------------

func roRegisterFormat(r string) (bool, string, string) {
	if len(r) < 2 || len(r) > 10 {
		return false, businessid.ReasonInvalidLength, "RO CUI must be 2-10 digits"
	}

	if !businessid.IsAllDigits(r) {
		return false, businessid.ReasonInvalidCharacters, "RO CUI must be all digits"
	}

	return true, businessid.ReasonOK, ""
}

func roRegisterChecksum(r string) bool {
	weights := [9]int{7, 5, 3, 2, 1, 7, 5, 3, 2}
	n := len(r)
	offset := 9 - (n - 1)

	if offset < 0 || offset > 9 {
		return false
	}

	sum := 0

	for i := range n - 1 {
		sum += int(r[i]-'0') * weights[offset+i]
	}

	check := sum * 10 % 11 % 10

	return check == int(r[n-1]-'0')
}

// ----------------------------------------------------------------------
// SE — Organisationsnummer (bolagsverket.se). 10 digits + Luhn.
// ----------------------------------------------------------------------

func seRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 10, "SE organisationsnummer")
}

func seRegisterChecksum(r string) bool { return businessid.Luhn(r) }

// ----------------------------------------------------------------------
// SI — Matična številka (ejn.gov.si). 7 digits, no standardized checksum
// (some sources cite mod-11 but implementations differ; kept format-only).
// ----------------------------------------------------------------------

func siRegisterFormat(r string) (bool, string, string) {
	return fixedDigitsFormat(r, 7, "SI matična številka")
}

// ----------------------------------------------------------------------
// SK — IČO (orsr.sk). 8 digits + mod-11 (same algorithm as CZ IČO).
// ----------------------------------------------------------------------

func skRegisterFormat(r string) (bool, string, string) { return fixedDigitsFormat(r, 8, "SK IČO") }

func skRegisterChecksum(r string) bool { return icoMod11(r) }

// ----------------------------------------------------------------------
// Shared helpers.
// ----------------------------------------------------------------------

// fixedDigitsFormat is the common "N digits, all digits" format check.
func fixedDigitsFormat(r string, n int, name string) (bool, string, string) {
	if len(r) != n {
		return false, businessid.ReasonInvalidLength, name + " must be " + itoa(n) + " digits"
	}

	if !businessid.IsAllDigits(r) {
		return false, businessid.ReasonInvalidCharacters, name + " must be all digits"
	}

	return true, businessid.ReasonOK, ""
}

// isAlnumUpperByte reports whether b is an upper-case ASCII letter or digit.
func isAlnumUpperByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z')
}

// parseInt parses an all-digit string as int. Caller must ensure content.
func parseInt(s string) int {
	n := 0

	for i := range len(s) {
		n = n*10 + int(s[i]-'0')
	}

	return n
}

// itoa converts a small non-negative int to decimal (used only in messages).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var buf [10]byte

	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}

	return string(buf[i:])
}

// esCIFOrDNIChecksum validates ES NIF (DNI/NIE/CIF) using the same
// algorithm as the VAT ES layer. Duplicated to keep euid independent of
// the vat package. Source: python-stdnum stdnum.es.cif and stdnum.es.dni.
func esCIFOrDNIChecksum(body string) bool {
	switch {
	case body[0] >= '0' && body[0] <= '9':
		return esDNI(body)
	case body[0] == 'X' || body[0] == 'Y' || body[0] == 'Z':
		return esNIE(body)
	case body[0] >= 'A' && body[0] <= 'Z':
		return esCIF(body)
	}

	return false
}

const esDNILetters = "TRWAGMYFPDXBNJZSQVHLCKE"

func esDNI(body string) bool {
	if !businessid.IsAllDigits(body[:8]) {
		return false
	}

	n := parseInt(body[:8])

	return body[8] == esDNILetters[n%23]
}

func esNIE(body string) bool {
	if !businessid.IsAllDigits(body[1:8]) {
		return false
	}

	prefix := strings.IndexByte("XYZ", body[0])
	n := prefix*10_000_000 + parseInt(body[1:8])

	return body[8] == esDNILetters[n%23]
}

func esCIF(body string) bool {
	if !businessid.IsAllDigits(body[1:8]) {
		return false
	}

	sum := 0

	for i := 1; i <= 7; i++ {
		d := int(body[i] - '0')
		if i%2 == 1 {
			doubled := 2 * d
			sum += doubled/10 + doubled%10
		} else {
			sum += d
		}
	}

	check := (10 - sum%10) % 10

	first := body[0]
	last := body[8]

	digitCheck := byte('0' + check)
	letterCheck := "JABCDEFGHI"[check]

	switch first {
	case 'A', 'B', 'E', 'H':
		return last == digitCheck
	case 'P', 'Q', 'R', 'S', 'W':
		return last == letterCheck
	default:
		// Strict rule: for the "either" middle group (C,D,F,G,J,L,M,N,U,V)
		// domestic entities use digit check. Foreign CIFs (rare) use
		// letter; callers who need that must use WithSubValidator with
		// a permissive override.
		return last == digitCheck
	}
}
