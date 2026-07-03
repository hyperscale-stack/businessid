// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package vat

import (
	"slices"

	"github.com/hyperscale-stack/businessid"
)

// countryValidator checks the body of a VAT number for a given country and
// returns (true, "") when it matches the national layout, or (false, reason)
// where reason is [businessid.ReasonInvalidLength] or
// [businessid.ReasonInvalidCharacters].
type countryValidator func(body string) (bool, string)

// countryChecksum returns true when body satisfies the national checksum
// algorithm. A nil value means the country only has format validation.
type countryChecksum func(body string) bool

// countrySpec pairs a format validator with an optional checksum validator
// for a given VAT country code.
type countrySpec struct {
	format   countryValidator
	checksum countryChecksum // nil = checksum not implemented
}

// vatCountrySpecs maps a 2-letter VAT prefix to a format+checksum
// specification.
//
// EU-27 (except FR which has a dedicated code path with checksum): AT, BE,
// BG, HR, CY, CZ, DE, DK, EE, EL, ES, FI, HU, IE, IT, LT, LU, LV, MT, NL,
// PL, PT, RO, SE, SI, SK.
//
// EEA and post-Brexit: NO, IS, LI, GB, XI (Northern Ireland).
//
// Sources: national tax authorities and the European Commission VIES portal
// (https://ec.europa.eu/taxation_customs/vies/). Layouts are cross-checked
// against https://en.wikipedia.org/wiki/VAT_identification_number.
// Checksum algorithms are in checksums.go with per-country references.
//
// NO/IS/LI: no standardized checksum published in a way we can verify
// against VIES vectors; kept format-only.
var vatCountrySpecs = map[string]countrySpec{
	"AT": {format: matchATBody, checksum: checksumATBody},
	"BE": {format: matchBEBody, checksum: checksumBEBody},
	"BG": {format: matchBGBody, checksum: checksumBGBody},
	"HR": {format: matchHRBody, checksum: checksumHRBody},
	"CY": {format: matchCYBody, checksum: checksumCYBody},
	"CZ": {format: matchCZBody, checksum: checksumCZBody},
	"DE": {format: matchDEBody, checksum: checksumDEBody},
	"DK": {format: matchDKBody, checksum: checksumDKBody},
	"EE": {format: matchEEBody, checksum: checksumEEBody},
	"EL": {format: matchELBody, checksum: checksumELBody},
	"ES": {format: matchESBody, checksum: checksumESBody},
	"FI": {format: matchFIBody, checksum: checksumFIBody},
	"HU": {format: matchHUBody, checksum: checksumHUBody},
	"IE": {format: matchIEBody, checksum: checksumIEBody},
	"IT": {format: matchITBody, checksum: checksumITBody},
	"LT": {format: matchLTBody, checksum: checksumLTBody},
	"LU": {format: matchLUBody, checksum: checksumLUBody},
	"LV": {format: matchLVBody, checksum: checksumLVBody},
	"MT": {format: matchMTBody, checksum: checksumMTBody},
	"NL": {format: matchNLBody, checksum: checksumNLBody},
	"PL": {format: matchPLBody, checksum: checksumPLBody},
	"PT": {format: matchPTBody, checksum: checksumPTBody},
	"RO": {format: matchROBody, checksum: checksumROBody},
	"SE": {format: matchSEBody, checksum: checksumSEBody},
	"SI": {format: matchSIBody, checksum: checksumSIBody},
	"SK": {format: matchSKBody, checksum: checksumSKBody},
	"GB": {format: matchGBBody, checksum: checksumGBBody},
	"XI": {format: matchGBBody, checksum: checksumGBBody},
	"NO": {format: matchNOBody},
	"IS": {format: matchISBody},
	"LI": {format: matchLIBody},
}

// matchFixedDigits reports whether body is exactly n ASCII digits.
func matchFixedDigits(body string, n int) (bool, string) {
	if len(body) != n {
		return false, businessid.ReasonInvalidLength
	}

	if !businessid.IsAllDigits(body) {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// matchDigitsInSet reports whether body is all ASCII digits and its length is
// one of the values in allowed.
func matchDigitsInSet(body string, allowed ...int) (bool, string) {
	if !slices.Contains(allowed, len(body)) {
		return false, businessid.ReasonInvalidLength
	}

	if !businessid.IsAllDigits(body) {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// matchDigitsRange reports whether body is all ASCII digits and its length is
// in [minLen, maxLen] inclusive.
func matchDigitsRange(body string, minLen, maxLen int) (bool, string) {
	if len(body) < minLen || len(body) > maxLen {
		return false, businessid.ReasonInvalidLength
	}

	if !businessid.IsAllDigits(body) {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// AT: 'U' + 8 digits.
func matchATBody(body string) (bool, string) {
	if len(body) != 9 {
		return false, businessid.ReasonInvalidLength
	}

	if body[0] != 'U' || !businessid.IsAllDigits(body[1:]) {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// BE: 10 digits; the first digit is 0 or 1.
func matchBEBody(body string) (bool, string) {
	if len(body) != 10 {
		return false, businessid.ReasonInvalidLength
	}

	if !businessid.IsAllDigits(body) || (body[0] != '0' && body[0] != '1') {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// BG: 9 or 10 digits.
func matchBGBody(body string) (bool, string) { return matchDigitsInSet(body, 9, 10) }

// HR: 11 digits (OIB).
func matchHRBody(body string) (bool, string) { return matchFixedDigits(body, 11) }

// CY: 8 digits + 1 upper-case letter.
func matchCYBody(body string) (bool, string) {
	if len(body) != 9 {
		return false, businessid.ReasonInvalidLength
	}

	if !businessid.IsAllDigits(body[:8]) || !isUpperLetter(body[8]) {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// CZ: 8, 9 or 10 digits.
func matchCZBody(body string) (bool, string) { return matchDigitsInSet(body, 8, 9, 10) }

// DE: 9 digits.
func matchDEBody(body string) (bool, string) { return matchFixedDigits(body, 9) }

// DK: 8 digits.
func matchDKBody(body string) (bool, string) { return matchFixedDigits(body, 8) }

// EE: 9 digits.
func matchEEBody(body string) (bool, string) { return matchFixedDigits(body, 9) }

// EL (Greece): 9 digits.
func matchELBody(body string) (bool, string) { return matchFixedDigits(body, 9) }

// ES: [A-Z0-9] + 7 digits + [A-Z0-9]. The first, last, or both characters may
// be letters depending on the taxpayer type (natural person, entity,
// non-resident, foreign national).
func matchESBody(body string) (bool, string) {
	if len(body) != 9 {
		return false, businessid.ReasonInvalidLength
	}

	if !isAlnumByte(body[0]) || !businessid.IsAllDigits(body[1:8]) || !isAlnumByte(body[8]) {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// FI: 8 digits.
func matchFIBody(body string) (bool, string) { return matchFixedDigits(body, 8) }

// HU: 8 digits.
func matchHUBody(body string) (bool, string) { return matchFixedDigits(body, 8) }

// matchIEBodyLegacy is matchIEBody plus acceptance of '+' and '*' as the
// position-2 character of the 8-char legacy layout. Only reachable through
// [WithLegacy]. Ireland issued these numbers before the 2013 renumbering;
// pre-2013 entities that were never renumbered still emit them.
//
// Source: Revenue.ie legacy VAT number guidance.
func matchIEBodyLegacy(body string) (bool, string) {
	switch len(body) {
	case 8:
		if businessid.IsAllDigits(body[:7]) && isUpperLetter(body[7]) {
			return true, ""
		}

		if isDigit(body[0]) && isLegacyIEAlt(body[1]) && businessid.IsAllDigits(body[2:7]) && isUpperLetter(body[7]) {
			return true, ""
		}

		return false, businessid.ReasonInvalidCharacters
	case 9:
		if businessid.IsAllDigits(body[:7]) && isUpperLetter(body[7]) && isUpperLetter(body[8]) {
			return true, ""
		}

		return false, businessid.ReasonInvalidCharacters
	default:
		return false, businessid.ReasonInvalidLength
	}
}

// isLegacyIEAlt reports whether b is an upper-case ASCII letter, '+', or '*'
// — the alphabet allowed in position 2 of the pre-2013 IE VAT layout.
func isLegacyIEAlt(b byte) bool { return isUpperLetter(b) || b == '+' || b == '*' }

// IE accepts three historical layouts:
//   - 7 digits + 1 letter                    (8 chars, current)
//   - 1 digit + 1 letter + 5 digits + 1 letter (8 chars, legacy)
//   - 7 digits + 2 letters                   (9 chars, current since 2013)
func matchIEBody(body string) (bool, string) {
	switch len(body) {
	case 8:
		if businessid.IsAllDigits(body[:7]) && isUpperLetter(body[7]) {
			return true, ""
		}

		if isDigit(body[0]) && isUpperLetter(body[1]) && businessid.IsAllDigits(body[2:7]) && isUpperLetter(body[7]) {
			return true, ""
		}

		return false, businessid.ReasonInvalidCharacters
	case 9:
		if businessid.IsAllDigits(body[:7]) && isUpperLetter(body[7]) && isUpperLetter(body[8]) {
			return true, ""
		}

		return false, businessid.ReasonInvalidCharacters
	default:
		return false, businessid.ReasonInvalidLength
	}
}

// IT: 11 digits.
func matchITBody(body string) (bool, string) { return matchFixedDigits(body, 11) }

// LT: 9 or 12 digits.
func matchLTBody(body string) (bool, string) { return matchDigitsInSet(body, 9, 12) }

// LU: 8 digits.
func matchLUBody(body string) (bool, string) { return matchFixedDigits(body, 8) }

// LV: 11 digits.
func matchLVBody(body string) (bool, string) { return matchFixedDigits(body, 11) }

// MT: 8 digits.
func matchMTBody(body string) (bool, string) { return matchFixedDigits(body, 8) }

// NL: 9 digits + 'B' + 2 digits.
func matchNLBody(body string) (bool, string) {
	if len(body) != 12 {
		return false, businessid.ReasonInvalidLength
	}

	if !businessid.IsAllDigits(body[:9]) || body[9] != 'B' || !businessid.IsAllDigits(body[10:]) {
		return false, businessid.ReasonInvalidCharacters
	}

	return true, ""
}

// PL: 10 digits.
func matchPLBody(body string) (bool, string) { return matchFixedDigits(body, 10) }

// PT: 9 digits.
func matchPTBody(body string) (bool, string) { return matchFixedDigits(body, 9) }

// RO: 2 to 10 digits.
func matchROBody(body string) (bool, string) { return matchDigitsRange(body, 2, 10) }

// SE: 12 digits.
func matchSEBody(body string) (bool, string) { return matchFixedDigits(body, 12) }

// SI: 8 digits.
func matchSIBody(body string) (bool, string) { return matchFixedDigits(body, 8) }

// SK: 10 digits.
func matchSKBody(body string) (bool, string) { return matchFixedDigits(body, 10) }

// GB / XI: 9 or 12 digits. GD### and HA### government/health authority
// numbers are out of scope.
func matchGBBody(body string) (bool, string) { return matchDigitsInSet(body, 9, 12) }

// NO: 9 digits (Norwegian organization number). The optional trailing "MVA"
// suffix must be stripped by the caller.
func matchNOBody(body string) (bool, string) { return matchFixedDigits(body, 9) }

// IS: 5 or 6 digits.
func matchISBody(body string) (bool, string) { return matchDigitsRange(body, 5, 6) }

// LI: 5 digits (Liechtenstein internal enterprise number).
func matchLIBody(body string) (bool, string) { return matchFixedDigits(body, 5) }

// isAlnumByte reports whether b is an upper-case ASCII letter or a digit.
func isAlnumByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'A' && b <= 'Z')
}

// isUpperLetter reports whether b is an upper-case ASCII letter.
func isUpperLetter(b byte) bool { return b >= 'A' && b <= 'Z' }

// isDigit reports whether b is an ASCII digit.
func isDigit(b byte) bool { return b >= '0' && b <= '9' }
