// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package vat

import (
	"github.com/hyperscale-stack/businessid"
)

// This file implements the national checksum algorithms for VAT numbers.
//
// Every function operates on the VAT body (everything after the 2-letter
// country prefix) and assumes the caller has already verified the layout
// via the matching entry in [vatCountrySpecs]. Callers therefore may
// assume digit-at-position invariants without re-checking.
//
// Primary source for each algorithm: the corresponding national tax
// authority documentation and the python-stdnum library
// (https://arthurdejong.org/python-stdnum/), which cross-verifies against
// VIES test vectors. Individual sources are noted per function.

// checksumATBody validates the AT checksum: U + 7 digits + 1 check digit.
// Algorithm: S = C1 + digit_sum(2*C2) + C3 + digit_sum(2*C4) + C5 +
// digit_sum(2*C6) + C7. Check = (10 - (S+4) mod 10) mod 10.
// Source: BMF Austria, UStG § 27 Abs. 4.
func checksumATBody(body string) bool {
	// body = "U" + 8 digits ; digits at body[1..8]
	sum := 0

	for i := range 7 {
		d := int(body[1+i] - '0')
		if i%2 == 1 { // 0-indexed: even positions (0,2,4,6) undoubled; odd (1,3,5) doubled
			doubled := 2 * d
			sum += doubled/10 + doubled%10
		} else {
			sum += d
		}
	}

	check := (10 - (sum+4)%10) % 10

	return check == int(body[8]-'0')
}

// checksumBEBody validates BE: 10 digits, check = 97 - (first 8 mod 97).
// Source: SPF Finances (finances.belgium.be) — VAT number layout.
func checksumBEBody(body string) bool {
	first8 := parseDigits(body[:8])
	last2 := parseDigits(body[8:])

	return 97-first8%97 == last2
}

// checksumBGBody validates BG. Three variants:
//   - 9-digit BULSTAT (legal entity): weights 1..8 mod 11, fallback 3..10
//   - 10-digit EGN (natural person): birth-based weights
//   - 10-digit foreigner (LNCh): distinct weights
//
// Source: NRA Bulgaria, ЗДДС Article 94.
func checksumBGBody(body string) bool {
	switch len(body) {
	case 9:
		return bulstat9(body)
	case 10:
		return egn10(body) || foreignerBG10(body)
	}

	return false
}

func bulstat9(body string) bool {
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

func egn10(body string) bool {
	weights := [9]int{2, 4, 8, 5, 10, 9, 7, 3, 6}
	sum := 0

	for i := range 9 {
		sum += int(body[i]-'0') * weights[i]
	}

	check := sum % 11 % 10

	return check == int(body[9]-'0')
}

func foreignerBG10(body string) bool {
	weights := [9]int{21, 19, 17, 13, 11, 9, 7, 3, 1}
	sum := 0

	for i := range 9 {
		sum += int(body[i]-'0') * weights[i]
	}

	return sum%10 == int(body[9]-'0')
}

// checksumHRBody validates HR: 11 digits, ISO 7064 MOD 11,10.
// Source: Porezna Uprava (porezna-uprava.hr) — OIB structure.
func checksumHRBody(body string) bool {
	p := 10

	for i := range 10 {
		s := (int(body[i]-'0') + p) % 10
		if s == 0 {
			s = 10
		}

		p = (s * 2) % 11
	}

	check := (11 - p) % 10

	return check == int(body[10]-'0')
}

// checksumCYBody validates CY: 8 digits + 1 letter check.
// Odd positions (0-indexed: 0,2,4,6) mapped: {0→1,1→0,2→5,3→7,4→9,
// 5→13,6→15,7→17,8→19,9→21}. Even positions summed as-is. Total mod 26
// + 'A' == check letter.
// Source: Cyprus Tax Department (mof.gov.cy).
func checksumCYBody(body string) bool {
	oddMap := [10]int{1, 0, 5, 7, 9, 13, 15, 17, 19, 21}
	sum := 0

	for i := range 8 {
		d := int(body[i] - '0')
		if i%2 == 0 {
			sum += oddMap[d]
		} else {
			sum += d
		}
	}

	return byte('A'+sum%26) == body[8]
}

// checksumCZBody validates CZ. Handled: 8-digit legal entity only.
// Weights 8-7-6-5-4-3-2 on first 7 digits; check = 11 - sum mod 11,
// then mapped: {0,10 → invalid ; 1 → 0 ; else stay}.
// Source: GFR Czech Republic (financnisprava.cz) DIČ.
// 9- and 10-digit variants (rodné číslo) not implemented — those cover
// natural persons whose VAT number is their birth-number, out of scope
// for a business identifier library.
func checksumCZBody(body string) bool {
	if len(body) != 8 {
		return false
	}

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
		check = 1 // 11-10=1
	default:
		check = 11 - r
	}

	return check == int(body[7]-'0')
}

// checksumDEBody validates DE: 9 digits, ISO 7064 MOD 11,10 iterative.
// Source: BZSt (bzst.de) — UStIdNr Aufbau.
func checksumDEBody(body string) bool {
	p := 10

	for i := range 8 {
		s := (int(body[i]-'0') + p) % 10
		if s == 0 {
			s = 10
		}

		p = (s * 2) % 11
	}

	check := (11 - p) % 10

	return check == int(body[8]-'0')
}

// checksumDKBody validates DK: 8 digits, weights 2-7-6-5-4-3-2-1, sum
// mod 11 = 0. First digit must not be 0.
// Source: SKAT (skat.dk) — CVR-nummer.
func checksumDKBody(body string) bool {
	if body[0] == '0' {
		return false
	}

	weights := [8]int{2, 7, 6, 5, 4, 3, 2, 1}
	sum := 0

	for i := range 8 {
		sum += int(body[i]-'0') * weights[i]
	}

	return sum%11 == 0
}

// checksumEEBody validates EE: 9 digits, weights 3-7-1-3-7-1-3-7 on
// first 8. Check = (10 - S mod 10) mod 10.
// Source: EMTA (emta.ee) — käibemaksukohuslasena registreerimise number.
func checksumEEBody(body string) bool {
	weights := [8]int{3, 7, 1, 3, 7, 1, 3, 7}
	sum := 0

	for i := range 8 {
		sum += int(body[i]-'0') * weights[i]
	}

	check := (10 - sum%10) % 10

	return check == int(body[8]-'0')
}

// checksumELBody validates EL: 9 digits, weights 256-128-64-32-16-8-4-2
// on first 8, check = S mod 11 mod 10.
// Source: AADE Greece (aade.gr) — ΑΦΜ.
func checksumELBody(body string) bool {
	weights := [8]int{256, 128, 64, 32, 16, 8, 4, 2}
	sum := 0

	for i := range 8 {
		sum += int(body[i]-'0') * weights[i]
	}

	return sum%11%10 == int(body[8]-'0')
}

// checksumESBody validates ES: 3 variants (DNI, NIE, CIF).
//   - DNI: 8 digits + letter, letter derived from digits mod 23 mapped
//     via "TRWAGMYFPDXBNJZSQVHLCKE".
//   - NIE: X|Y|Z + 7 digits + letter. X→0, Y→1, Z→2, then compute like DNI.
//   - CIF: entity letter + 7 digits + digit/letter check.
//
// Source: Agencia Tributaria (agenciatributaria.es).
func checksumESBody(body string) bool {
	switch {
	case body[0] >= '0' && body[0] <= '9':
		return dniCheck(body)
	case body[0] == 'X' || body[0] == 'Y' || body[0] == 'Z':
		return nieCheck(body)
	case body[0] >= 'A' && body[0] <= 'Z':
		return cifCheck(body)
	}

	return false
}

const dniLetters = "TRWAGMYFPDXBNJZSQVHLCKE"

func dniCheck(body string) bool {
	if !businessid.IsAllDigits(body[:8]) {
		return false
	}

	n := parseDigits(body[:8])

	return body[8] == dniLetters[n%23]
}

func nieCheck(body string) bool {
	if !businessid.IsAllDigits(body[1:8]) {
		return false
	}

	prefix := 0

	switch body[0] {
	case 'X':
		prefix = 0
	case 'Y':
		prefix = 1
	case 'Z':
		prefix = 2
	}

	n := prefix*10_000_000 + parseDigits(body[1:8])

	return body[8] == dniLetters[n%23]
}

func cifCheck(body string) bool {
	if !businessid.IsAllDigits(body[1:8]) {
		return false
	}

	// Odd positions (1,3,5,7) doubled with digit-sum, even positions
	// summed as-is. Note: positions here are 1-based in the CIF body
	// (7 middle digits), 0-based in code we skip body[0] entity letter.
	sum := 0

	for i := 1; i <= 7; i++ {
		d := int(body[i] - '0')
		if i%2 == 1 { // odd 1,3,5,7 → doubled
			doubled := 2 * d
			sum += doubled/10 + doubled%10
		} else {
			sum += d
		}
	}

	check := (10 - sum%10) % 10

	// Entity letter dictates whether check is digit or letter.
	// Digit-only: A B E H
	// Letter-only: P Q R S W (mapped to JABCDEFGHI: 0→J,1→A..9→I)
	// Either: C D F G J L M N U V (default to digit but letter also accepted)
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
		return last == digitCheck || last == letterCheck
	}
}

// checksumFIBody validates FI: 8 digits, weights 7-9-10-5-8-4-2 on
// first 7; check = 11 - S mod 11; reject if 10.
// Source: Vero (vero.fi) — Y-tunnus / ALV-numero.
func checksumFIBody(body string) bool {
	weights := [7]int{7, 9, 10, 5, 8, 4, 2}
	sum := 0

	for i := range 7 {
		sum += int(body[i]-'0') * weights[i]
	}

	r := sum % 11
	if r == 1 {
		return false
	}

	check := 0
	if r > 1 {
		check = 11 - r
	}

	return check == int(body[7]-'0')
}

// checksumHUBody validates HU: 8 digits, weights 9-7-3-1-9-7-3 on
// first 7; check = (10 - S mod 10) mod 10.
// Source: NAV Hungary (nav.gov.hu) — adószám.
func checksumHUBody(body string) bool {
	weights := [7]int{9, 7, 3, 1, 9, 7, 3}
	sum := 0

	for i := range 7 {
		sum += int(body[i]-'0') * weights[i]
	}

	check := (10 - sum%10) % 10

	return check == int(body[7]-'0')
}

// checksumIEBody validates IE. Two supported layouts:
//   - 7D + 1L: weights 8-7-6-5-4-3-2 on 7 digits, mod 23 → letter table
//   - 7D + 2L: as above but the 2nd letter adds its (position-in-alphabet)*9
//
// The legacy D+L+5D+L layout is Format-accepted with WithLegacy() but
// its checksum is skipped (algorithm not stably documented).
// Source: Revenue Ireland (revenue.ie) — VAT number structure.
func checksumIEBody(body string) bool {
	switch len(body) {
	case 8:
		if !businessid.IsAllDigits(body[:7]) {
			return false // legacy layout — no checksum
		}

		return ieChecksum7D(body[:7], body[7], 0)
	case 9:
		if !businessid.IsAllDigits(body[:7]) || !isUpperLetter(body[7]) || !isUpperLetter(body[8]) {
			return false
		}

		letter2Val := int(body[8]-'A') + 1
		if body[8] == 'W' {
			letter2Val = 0
		}

		return ieChecksum7D(body[:7], body[7], letter2Val)
	}

	return false
}

func ieChecksum7D(digits string, checkLetter byte, letter2Extra int) bool {
	weights := [7]int{8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := range 7 {
		sum += int(digits[i]-'0') * weights[i]
	}

	sum += letter2Extra * 9

	r := sum % 23

	var expected byte
	if r == 0 {
		expected = 'W'
	} else {
		expected = byte('A' + r - 1)
	}

	return expected == checkLetter
}

// checksumITBody validates IT: 11 digits, Luhn (identical algorithm to SIREN).
// Source: Agenzia delle Entrate (agenziaentrate.gov.it) — partita IVA.
func checksumITBody(body string) bool { return businessid.Luhn(body) }

// checksumLTBody validates LT. Weighted sum including the check digit
// must be 0 mod 11; if the primary pass yields 10, a secondary pass with
// shifted weights is applied.
// Source: VMI Lithuania (vmi.lt) and python-stdnum lt.vat.
func checksumLTBody(body string) bool {
	// 12-digit LT numbers extend the weight sequence with 1,2,3.
	weights1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 1, 2, 3}
	weights2 := []int{3, 4, 5, 6, 7, 8, 9, 1, 2, 3, 4, 5}

	n := len(body)
	if n != 9 && n != 12 {
		return false
	}

	sum := 0
	for i := range n {
		sum += int(body[i]-'0') * weights1[i]
	}

	r := sum % 11

	if r == 10 {
		sum2 := 0
		for i := range n {
			sum2 += int(body[i]-'0') * weights2[i]
		}

		r = sum2 % 11 % 10
	}

	return r == 0
}

// checksumLUBody validates LU: 8 digits, (first 6) mod 89 = last 2.
// Source: AED Luxembourg (aed.public.lu) — numéro de TVA.
func checksumLUBody(body string) bool {
	first6 := parseDigits(body[:6])
	last2 := parseDigits(body[6:])

	return first6%89 == last2
}

// checksumLVBody validates LV. Two variants:
//   - Legal entity (first digit > 3): weights 9-1-4-8-3-10-2-5-7-6 on
//     first 10; check = (3 - S mod 11) mod 11; reject if 10.
//   - Natural person (first digit ≤ 3): birth-date encoded number.
//     Only legal entity path is implemented; natural persons return false.
//
// Source: VID Latvia (vid.gov.lv) — PVN reģistrācijas numurs.
func checksumLVBody(body string) bool {
	if body[0] <= '3' {
		return false
	}

	weights := [10]int{9, 1, 4, 8, 3, 10, 2, 5, 7, 6}
	sum := 0

	for i := range 10 {
		sum += int(body[i]-'0') * weights[i]
	}

	r := sum % 11

	check := (3 - r + 22) % 11 // ensure positive
	if check == 10 {
		return false
	}

	return check == int(body[10]-'0')
}

// checksumMTBody validates MT: 8 digits, weights 3-4-6-7-8-9 on first 6,
// check = 37 - S mod 37; check == last 2 digits as integer.
// Source: CFR Malta (cfr.gov.mt) — VAT number.
func checksumMTBody(body string) bool {
	weights := [6]int{3, 4, 6, 7, 8, 9}
	sum := 0

	for i := range 6 {
		sum += int(body[i]-'0') * weights[i]
	}

	check := 37 - sum%37
	last2 := parseDigits(body[6:])

	return check == last2
}

// checksumNLBody validates NL: 9-digit body + B + 2-digit branch. The
// checksum is computed on the 9-digit body using weights 9-8-7-6-5-4-3-2;
// S mod 11 must equal the 9th digit. Reject if S mod 11 == 10.
// NB: the new post-2020 format (OB-nummer) uses letters — validated at
// format level only.
// Source: Belastingdienst (belastingdienst.nl) — btw-nummer.
func checksumNLBody(body string) bool {
	weights := [8]int{9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := range 8 {
		sum += int(body[i]-'0') * weights[i]
	}

	r := sum % 11
	if r == 10 {
		return false
	}

	return r == int(body[8]-'0')
}

// checksumPLBody validates PL: 10 digits, weights 6-5-7-2-3-4-5-6-7 on
// first 9; check = S mod 11; reject if 10.
// Source: KAS Poland (podatki.gov.pl) — NIP.
func checksumPLBody(body string) bool {
	weights := [9]int{6, 5, 7, 2, 3, 4, 5, 6, 7}
	sum := 0

	for i := range 9 {
		sum += int(body[i]-'0') * weights[i]
	}

	r := sum % 11
	if r == 10 {
		return false
	}

	return r == int(body[9]-'0')
}

// checksumPTBody validates PT: 9 digits, weights 9-8-7-6-5-4-3-2 on
// first 8; if S mod 11 < 2 → check = 0, else check = 11 - S mod 11.
// Source: AT Portugal (portaldasfinancas.gov.pt) — NIF.
func checksumPTBody(body string) bool {
	weights := [8]int{9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := range 8 {
		sum += int(body[i]-'0') * weights[i]
	}

	r := sum % 11

	check := 0
	if r >= 2 {
		check = 11 - r
	}

	return check == int(body[8]-'0')
}

// checksumROBody validates RO: 2..10 digits. Weights [7,5,3,2,1,7,5,3,2]
// right-aligned to length-1 digits; sum × 10 mod 11 mod 10 = last digit.
// Source: ANAF Romania (anaf.ro) — CUI/CIF.
func checksumROBody(body string) bool {
	weights := [9]int{7, 5, 3, 2, 1, 7, 5, 3, 2}
	n := len(body)
	offset := 9 - (n - 1) // right-align (skip leading weights)

	if offset < 0 || offset > 9 {
		return false
	}

	sum := 0

	for i := range n - 1 {
		sum += int(body[i]-'0') * weights[offset+i]
	}

	check := (sum * 10) % 11 % 10

	return check == int(body[n-1]-'0')
}

// checksumSEBody validates SE: 12 digits, Luhn on first 10, last 2 are
// branch code (01..94 conventionally).
// Source: Skatteverket (skatteverket.se) — momsregistreringsnummer.
func checksumSEBody(body string) bool {
	if !businessid.Luhn(body[:10]) {
		return false
	}

	// Accept any 2-digit branch code (01..99). The spec limits to 01-94
	// but 95-99 have appeared for specific administrative cases.
	branch := parseDigits(body[10:])

	return branch >= 1 && branch <= 99
}

// checksumSIBody validates SI: 8 digits, weights 8-7-6-5-4-3-2 on
// first 7; check = 11 - S mod 11; if 10, invalid; if 11, check = 0.
// Source: FURS (fu.gov.si) — davčna številka.
func checksumSIBody(body string) bool {
	weights := [7]int{8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := range 7 {
		sum += int(body[i]-'0') * weights[i]
	}

	r := sum % 11

	check := 11 - r
	if check == 10 {
		return false
	}

	if check == 11 {
		check = 0
	}

	return check == int(body[7]-'0')
}

// checksumSKBody validates SK: 10 digits, the number as integer must be
// divisible by 11.
// Source: Finančná Správa (financnasprava.sk) — IČ DPH.
func checksumSKBody(body string) bool {
	// The number can be very large; compute mod 11 digit-wise.
	r := 0

	for i := range 10 {
		r = (r*10 + int(body[i]-'0')) % 11
	}

	return r == 0
}

// checksumGBBody validates GB. Two algorithms coexist historically:
// "97" (used until ~2010) and "9755" (used since). Accept when either
// passes. 12-digit VAT groups: first 9 digits follow the algo, last 3
// are the branch code (accepted as any 3 digits).
// Source: HMRC (gov.uk/vat-registration) — VAT number check.
func checksumGBBody(body string) bool {
	switch len(body) {
	case 9:
		return gbAlgo(body, false) || gbAlgo(body, true)
	case 12:
		return gbAlgo(body[:9], false) || gbAlgo(body[:9], true)
	}

	return false
}

func gbAlgo(nine string, use9755 bool) bool {
	weights := [7]int{8, 7, 6, 5, 4, 3, 2}
	sum := 0

	for i := range 7 {
		sum += int(nine[i]-'0') * weights[i]
	}

	last2 := parseDigits(nine[7:])

	if use9755 {
		sum += 55
	}

	// Modulo-97 check: (sum + last2) mod 97 == 0.
	return (sum+last2)%97 == 0
}
