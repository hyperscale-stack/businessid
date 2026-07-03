// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/euid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := euid.New()

	assert.Equal(t, businessid.IdentifierKindEUID, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: true, Registry: false}, p.Capabilities())
}

// TestValidateFormatBRISShape covers the meta-format layer: BRIS layout,
// charset, register length. Uses a country whose national validator is
// permissive so we isolate meta-format failures.
func TestValidateFormatBRISShape(t *testing.T) {
	t.Parallel()

	p := euid.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "no-dot", value: "FRRCS552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "bad-country", value: "F1RCS.552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "register-non-alnum", value: "FRR-CS.552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "empty-registration", value: "FRRCS.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "register-too-long", value: "FRVERYLONGREGISTERNAMEXX.552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "registration-too-long", value: "DEHRB." + strings.Repeat("A", 65), wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "registration-invalid-char", value: "DEHRB.ABC_123", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// Lower-case is canonicalized to upper.
		{name: "lower-canonicalized", value: "frrcs.552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// DE has a permissive register format (BRIS accepts free-form
		// court/register/entry strings) — used here as a "shape-only" fixture.
		{name: "de-shape-ok", value: "DEHRB.HAMBURG/B-12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// Plus sign accepted in the registration charset.
		{name: "registration-with-plus", value: "DEHRB.MUNICH+B12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateFormat(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status, "reason=%s message=%s", res.ReasonCode, res.Message)
			assert.Equal(t, tc.wantReason, res.ReasonCode)
		})
	}
}

// TestValidateFormatPerCountry covers the 27 EU native register-format
// checks. Registration values are shape-correct examples referenced from
// national register documentation and python-stdnum test vectors.
func TestValidateFormatPerCountry(t *testing.T) {
	t.Parallel()

	p := euid.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
	}{
		// -- Valid cases -------------------------------------------------

		// AT — Firmenbuchnummer: 1-6 digits + 1 letter. Source: justiz.gv.at.
		{name: "at-valid", value: "ATFN.123456A", wantStatus: businessid.ValidationStatusValid},
		// BE — KBO: 10 digits starting 0/1. Source: kbopub.economie.fgov.be
		//        (AB InBev 0417497106).
		{name: "be-valid", value: "BEBCE.0417497106", wantStatus: businessid.ValidationStatusValid},
		// BG — EIK: 9 digits (BULSTAT valid vector from python-stdnum).
		{name: "bg-valid", value: "BGEIK.040808527", wantStatus: businessid.ValidationStatusValid},
		// HR — MBS: 8 digits. Source: sudreg.pravosudje.hr.
		{name: "hr-valid", value: "HRMBS.08000001", wantStatus: businessid.ValidationStatusValid},
		// CY — HE number: 6 digits. Source: efiling.drcor.mcit.gov.cy.
		{name: "cy-valid", value: "CYHE.123456", wantStatus: businessid.ValidationStatusValid},
		// CZ — IČO: 8 digits (ČD railway 00006947 verified mod-11).
		{name: "cz-valid", value: "CZOR.00006947", wantStatus: businessid.ValidationStatusValid},
		// DE — Handelsregister: free-form. Source: unternehmensregister.de.
		{name: "de-valid", value: "DEHRB.HAMBURG/B-12345", wantStatus: businessid.ValidationStatusValid},
		// DK — CVR: 8 digits (Carlsberg 61056416 verified).
		{name: "dk-valid", value: "DKCVR.61056416", wantStatus: businessid.ValidationStatusValid},
		// EE — Registrikood: 8 digits (stdnum valid vector).
		{name: "ee-valid", value: "EERIK.12345678", wantStatus: businessid.ValidationStatusValid},
		// EL — GEMI: 12 digits. Source: businessregistry.gr.
		{name: "el-valid", value: "ELGEMI.123456789000", wantStatus: businessid.ValidationStatusValid},
		// ES — CIF: 9 chars (Santander A39000013 verified).
		{name: "es-valid", value: "ESRMC.A39000013", wantStatus: businessid.ValidationStatusValid},
		// FI — Y-tunnus: 8 digits (Nokia 01120389 verified).
		{name: "fi-valid", value: "FIPRH.01120389", wantStatus: businessid.ValidationStatusValid},
		// FR — SIREN: 9 digits (LVMH 552100554 verified Luhn).
		{name: "fr-valid", value: "FRRCS.552100554", wantStatus: businessid.ValidationStatusValid},
		// HU — Cégjegyzékszám: 10 digits (with optional dashes).
		{name: "hu-valid", value: "HUCG.01-09-123456", wantStatus: businessid.ValidationStatusValid},
		// IE — CRO: 5-7 digits (Google Ireland 368047 hypothetical).
		{name: "ie-valid", value: "IECRO.368047", wantStatus: businessid.ValidationStatusValid},
		// IT — Codice Fiscale entità: 11 digits (Stellantis 07973780013 verified Luhn).
		{name: "it-valid", value: "ITRI.07973780013", wantStatus: businessid.ValidationStatusValid},
		// LT — 9 digits (stdnum valid vector).
		{name: "lt-valid", value: "LTJAR.100000006", wantStatus: businessid.ValidationStatusValid},
		// LU — RCSL: B + 4-6 digits.
		{name: "lu-valid", value: "LURCSL.B12345", wantStatus: businessid.ValidationStatusValid},
		// LV — 11 digits (Latvenergo 40003032949 verified).
		{name: "lv-valid", value: "LVURE.40003032949", wantStatus: businessid.ValidationStatusValid},
		// MT — Company number: C + 4-6 digits.
		{name: "mt-valid", value: "MTMBR.C12345", wantStatus: businessid.ValidationStatusValid},
		// NL — KVK: 8 digits (stdnum valid vector).
		{name: "nl-valid", value: "NLKVK.68750110", wantStatus: businessid.ValidationStatusValid},
		// PL — KRS: 10 digits (no checksum).
		{name: "pl-valid", value: "PLKRS.0000123456", wantStatus: businessid.ValidationStatusValid},
		// PT — NIPC: 9 digits (Galp 504499777 verified mod-11).
		{name: "pt-valid", value: "PTRN.504499777", wantStatus: businessid.ValidationStatusValid},
		// RO — CUI: 2-10 digits (Petrom 14186770 verified).
		{name: "ro-valid", value: "ROORCT.14186770", wantStatus: businessid.ValidationStatusValid},
		// SE — Organisationsnummer: 10 digits Luhn (Volvo 5560032291 verified).
		{name: "se-valid", value: "SEBR.1234567897", wantStatus: businessid.ValidationStatusValid},
		// SI — Matična številka: 7 digits.
		{name: "si-valid", value: "SISRG.1234567", wantStatus: businessid.ValidationStatusValid},
		// SK — IČO: 8 digits mod-11.
		{name: "sk-valid", value: "SKOR.31333532", wantStatus: businessid.ValidationStatusValid},

		// -- Format-invalid cases: wrong length or wrong charset -----------

		{name: "at-invalid-empty-registration", value: "ATFN.", wantStatus: businessid.ValidationStatusInvalid},
		{name: "be-invalid-length", value: "BEBCE.041749710", wantStatus: businessid.ValidationStatusInvalid},
		{name: "bg-invalid-length", value: "BGEIK.12345678", wantStatus: businessid.ValidationStatusInvalid},
		{name: "hr-invalid-length", value: "HRMBS.0800000", wantStatus: businessid.ValidationStatusInvalid},
		{name: "cy-invalid-length", value: "CYHE.12345", wantStatus: businessid.ValidationStatusInvalid},
		{name: "cz-invalid-length", value: "CZOR.1234567", wantStatus: businessid.ValidationStatusInvalid},
		{name: "de-invalid-non-alnum", value: "DEHRB.___", wantStatus: businessid.ValidationStatusInvalid},
		{name: "dk-invalid-length", value: "DKCVR.6105641", wantStatus: businessid.ValidationStatusInvalid},
		{name: "ee-invalid-length", value: "EERIK.1234567", wantStatus: businessid.ValidationStatusInvalid},
		{name: "el-invalid-length", value: "ELGEMI.12345678900", wantStatus: businessid.ValidationStatusInvalid},
		{name: "es-invalid-length", value: "ESRMC.A3900001", wantStatus: businessid.ValidationStatusInvalid},
		{name: "fi-invalid-length", value: "FIPRH.0112038", wantStatus: businessid.ValidationStatusInvalid},
		{name: "fr-invalid-length", value: "FRRCS.55210055", wantStatus: businessid.ValidationStatusInvalid},
		{name: "hu-invalid-length", value: "HUCG.01-09-12345", wantStatus: businessid.ValidationStatusInvalid},
		{name: "ie-invalid-length", value: "IECRO.1234", wantStatus: businessid.ValidationStatusInvalid},
		{name: "it-invalid-length", value: "ITRI.0797378001", wantStatus: businessid.ValidationStatusInvalid},
		{name: "lt-invalid-length", value: "LTJAR.10000000", wantStatus: businessid.ValidationStatusInvalid},
		{name: "lu-invalid-no-b", value: "LURCSL.X12345", wantStatus: businessid.ValidationStatusInvalid},
		{name: "lv-invalid-length", value: "LVURE.4000303294", wantStatus: businessid.ValidationStatusInvalid},
		{name: "mt-invalid-no-c", value: "MTMBR.X12345", wantStatus: businessid.ValidationStatusInvalid},
		{name: "nl-invalid-length", value: "NLKVK.6875011", wantStatus: businessid.ValidationStatusInvalid},
		{name: "pl-invalid-length", value: "PLKRS.000012345", wantStatus: businessid.ValidationStatusInvalid},
		{name: "pt-invalid-length", value: "PTRN.50449977", wantStatus: businessid.ValidationStatusInvalid},
		{name: "ro-invalid-length", value: "ROORCT.1", wantStatus: businessid.ValidationStatusInvalid},
		{name: "se-invalid-length", value: "SEBR.123456789", wantStatus: businessid.ValidationStatusInvalid},
		{name: "si-invalid-length", value: "SISRG.123456", wantStatus: businessid.ValidationStatusInvalid},
		{name: "sk-invalid-length", value: "SKOR.3133353", wantStatus: businessid.ValidationStatusInvalid},

		// Character-class failures per country.
		{name: "at-invalid-chars", value: "ATFN.12345A_", wantStatus: businessid.ValidationStatusInvalid},
		{name: "be-invalid-first", value: "BEBCE.2417497106", wantStatus: businessid.ValidationStatusInvalid},
		{name: "bg-invalid-chars", value: "BGEIK.12345678A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "cz-invalid-chars", value: "CZOR.1234567A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "dk-invalid-chars", value: "DKCVR.6105641A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "ee-invalid-chars", value: "EERIK.1234567A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "es-invalid-chars", value: "ESRMC.A390000!3", wantStatus: businessid.ValidationStatusInvalid},
		{name: "fi-invalid-chars", value: "FIPRH.0112038A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "hu-invalid-chars", value: "HUCG.01-09-12345A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "ie-invalid-chars", value: "IECRO.12345A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "lu-invalid-chars", value: "LURCSL.B1234A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "mt-invalid-chars", value: "MTMBR.C1234A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "ro-invalid-chars", value: "ROORCT.1418677A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "fixed-digits-non-digit-hr", value: "HRMBS.0800000A", wantStatus: businessid.ValidationStatusInvalid},
		{name: "de-empty-registration", value: "DEHRB.___", wantStatus: businessid.ValidationStatusInvalid},

		// AT invalid chars (digits contain a letter).
		{name: "at-invalid-mixed", value: "ATFN.12A345B", wantStatus: businessid.ValidationStatusInvalid},

		// DE with only separators (no alphanumeric character).
		{name: "de-invalid-no-alnum", value: "DEHRB.- -", wantStatus: businessid.ValidationStatusInvalid},

		// ES with non-alnum first byte.
		{name: "es-invalid-first-slash", value: "ESRMC./12345678", wantStatus: businessid.ValidationStatusInvalid},

		// LU with letters in the digit tail.
		{name: "lu-invalid-tail", value: "LURCSL.B12A4", wantStatus: businessid.ValidationStatusInvalid},

		// MT with letters in the digit tail.
		{name: "mt-invalid-tail", value: "MTMBR.C12A4", wantStatus: businessid.ValidationStatusInvalid},

		// Non-EU code with no override → falls through generic BRIS.
		{name: "xi-no-override", value: "XICH.ANYTHING", wantStatus: businessid.ValidationStatusValid},

		// AT registration too short (1 char) triggers len < 2.
		{name: "at-invalid-too-short-native", value: "ATFN.A", wantStatus: businessid.ValidationStatusInvalid},
		// AT registration too long (8 chars) triggers len > 7.
		{name: "at-invalid-too-long-native", value: "ATFN.12345678", wantStatus: businessid.ValidationStatusInvalid},

		// LU too short.
		{name: "lu-invalid-too-short", value: "LURCSL.B12", wantStatus: businessid.ValidationStatusInvalid},

		// MT too short.
		{name: "mt-invalid-too-short", value: "MTMBR.C12", wantStatus: businessid.ValidationStatusInvalid},

		// FR non-digit rejected by SIREN sub-validator.
		{name: "fr-invalid-non-digit", value: "FRRCS.55210055A", wantStatus: businessid.ValidationStatusInvalid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateFormat(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status, "reason=%s message=%s", res.ReasonCode, res.Message)
		})
	}
}

// TestWithCountryValidator verifies the escape hatch for injecting a
// custom register validator (used for non-EU codes or overrides).
func TestWithCountryValidator(t *testing.T) {
	t.Parallel()

	// A custom XI (Northern Ireland) validator that accepts any
	// registration matching "NI" + 6 digits, without checksum.
	xiValidator := euid.NewRegisterValidator(
		nil,
		func(r string) (bool, string, string) {
			if len(r) != 8 {
				return false, businessid.ReasonInvalidLength, "XI reg must be NI + 6 digits"
			}
			if r[0] != 'N' || r[1] != 'I' {
				return false, businessid.ReasonInvalidCharacters, "XI reg must start with NI"
			}
			for i := 2; i < 8; i++ {
				if r[i] < '0' || r[i] > '9' {
					return false, businessid.ReasonInvalidCharacters, "XI reg tail must be digits"
				}
			}
			return true, businessid.ReasonOK, ""
		},
		nil,
	)

	p := euid.New(euid.WithCountryValidator("XI", xiValidator))

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
	}{
		{name: "xi-custom-valid", value: "XICH.NI123456", wantStatus: businessid.ValidationStatusValid},
		{name: "xi-custom-invalid-length", value: "XICH.NI12345", wantStatus: businessid.ValidationStatusInvalid},
		{name: "xi-custom-invalid-prefix", value: "XICH.XX123456", wantStatus: businessid.ValidationStatusInvalid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateFormat(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status, "reason=%s message=%s", res.ReasonCode, res.Message)
		})
	}
}

// TestValidateChecksumPerCountry covers the countries whose national
// register has a documented checksum algorithm.
func TestValidateChecksumPerCountry(t *testing.T) {
	t.Parallel()

	p := euid.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
	}{
		// BE (mod-97), CZ (mod-11), DK (mod-11), EE (mod-11), ES CIF, FI (mod-11),
		// FR SIREN Luhn, IT Luhn, LT (mod-11), LV, NL (mod-11), PT, RO, SE Luhn,
		// SK IČO.
		{name: "be-valid-checksum", value: "BEBCE.0417497106", wantStatus: businessid.ValidationStatusValid},
		{name: "be-invalid-checksum", value: "BEBCE.0417497100", wantStatus: businessid.ValidationStatusInvalid},

		{name: "cz-valid-checksum", value: "CZOR.00006947", wantStatus: businessid.ValidationStatusValid},
		{name: "cz-invalid-checksum", value: "CZOR.00006940", wantStatus: businessid.ValidationStatusInvalid},

		{name: "dk-valid-checksum", value: "DKCVR.61056416", wantStatus: businessid.ValidationStatusValid},
		{name: "dk-invalid-checksum", value: "DKCVR.61056410", wantStatus: businessid.ValidationStatusInvalid},

		{name: "ee-valid-checksum", value: "EERIK.12345678", wantStatus: businessid.ValidationStatusValid},

		{name: "es-valid-checksum", value: "ESRMC.A39000013", wantStatus: businessid.ValidationStatusValid},

		{name: "fi-valid-checksum", value: "FIPRH.01120389", wantStatus: businessid.ValidationStatusValid},

		{name: "fr-valid-luhn", value: "FRRCS.552100554", wantStatus: businessid.ValidationStatusValid},
		{name: "fr-invalid-luhn", value: "FRRCS.552100555", wantStatus: businessid.ValidationStatusInvalid},

		{name: "it-valid-luhn", value: "ITRI.07973780013", wantStatus: businessid.ValidationStatusValid},

		{name: "lt-valid-checksum", value: "LTJAR.100000006", wantStatus: businessid.ValidationStatusValid},

		{name: "lv-valid-checksum", value: "LVURE.40003032949", wantStatus: businessid.ValidationStatusValid},

		{name: "nl-valid-checksum", value: "NLKVK.68750110", wantStatus: businessid.ValidationStatusValid},

		{name: "pt-valid-checksum", value: "PTRN.504499777", wantStatus: businessid.ValidationStatusValid},

		{name: "ro-valid-checksum", value: "ROORCT.14186770", wantStatus: businessid.ValidationStatusValid},

		{name: "se-valid-luhn", value: "SEBR.1234567897", wantStatus: businessid.ValidationStatusValid},

		{name: "sk-valid-checksum", value: "SKOR.31333532", wantStatus: businessid.ValidationStatusValid},

		// Countries with no documented checksum → Unsupported.
		{name: "hr-unsupported", value: "HRMBS.08000001", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "cy-unsupported", value: "CYHE.123456", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "de-unsupported", value: "DEHRB.HAMBURG/B-12345", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "el-unsupported", value: "ELGEMI.123456789000", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "hu-unsupported", value: "HUCG.01-09-123456", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "ie-unsupported", value: "IECRO.368047", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "lu-unsupported", value: "LURCSL.B12345", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "mt-unsupported", value: "MTMBR.C12345", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "pl-unsupported", value: "PLKRS.0000123456", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "si-unsupported", value: "SISRG.1234567", wantStatus: businessid.ValidationStatusUnsupported},

		// Non-EU code → Unsupported (no native validator).
		{name: "xi-unsupported", value: "XICH.NI123456", wantStatus: businessid.ValidationStatusUnsupported},
		{name: "gb-unsupported", value: "GBCH.12345678", wantStatus: businessid.ValidationStatusUnsupported},

		// -- Coverage-driven additions ---------------------------------

		// BG EIK 9-digit valid checksum.
		{name: "bg-valid-checksum", value: "BGEIK.040808527", wantStatus: businessid.ValidationStatusValid},
		{name: "bg-invalid-checksum", value: "BGEIK.040808520", wantStatus: businessid.ValidationStatusInvalid},

		// ES DNI (digit prefix) and NIE (X/Y/Z prefix).
		{name: "es-valid-dni", value: "ESRMC.12345678Z", wantStatus: businessid.ValidationStatusValid},
		{name: "es-valid-nie", value: "ESRMC.X0000001R", wantStatus: businessid.ValidationStatusValid},
		{name: "es-invalid-first-char", value: "ESRMC.!12345678", wantStatus: businessid.ValidationStatusInvalid},

		// FI invalid (wrong check).
		{name: "fi-invalid-checksum", value: "FIPRH.01120380", wantStatus: businessid.ValidationStatusInvalid},

		// EE r==10 alt-weights path (constructed vector where sum1%11==10).
		// digits 5,5,5,5,5,5,5,X where first-pass sum ≡ 10 mod 11.
		{name: "ee-alt-path-invalid", value: "EERIK.10000015", wantStatus: businessid.ValidationStatusInvalid},

		// IT invalid Luhn.
		{name: "it-invalid-luhn", value: "ITRI.07973780010", wantStatus: businessid.ValidationStatusInvalid},

		// SK invalid.
		{name: "sk-invalid-checksum", value: "SKOR.31333530", wantStatus: businessid.ValidationStatusInvalid},

		// SE invalid.
		{name: "se-invalid-luhn", value: "SEBR.1234567890", wantStatus: businessid.ValidationStatusInvalid},

		// NL invalid.
		{name: "nl-invalid-checksum", value: "NLKVK.68750100", wantStatus: businessid.ValidationStatusInvalid},

		// PT invalid.
		{name: "pt-invalid-checksum", value: "PTRN.504499770", wantStatus: businessid.ValidationStatusInvalid},

		// RO invalid.
		{name: "ro-invalid-checksum", value: "ROORCT.14186771", wantStatus: businessid.ValidationStatusInvalid},

		// LT invalid.
		{name: "lt-invalid-checksum", value: "LTJAR.100000005", wantStatus: businessid.ValidationStatusInvalid},

		// LV natural person (first digit ≤ 3).
		{name: "lv-valid-natural", value: "LVURE.01010120001", wantStatus: businessid.ValidationStatusValid},

		// EE invalid.
		{name: "ee-invalid-checksum", value: "EERIK.12345670", wantStatus: businessid.ValidationStatusInvalid},

		// EE alt-weights path (sum1 % 11 == 10 triggers alt weights).
		{name: "ee-valid-alt", value: "EERIK.00000906", wantStatus: businessid.ValidationStatusValid},

		// LT alt-weights path.
		{name: "lt-valid-alt", value: "LTJAR.110000002", wantStatus: businessid.ValidationStatusValid},

		// BG BULSTAT alt-alt path (both sums mod-11 == 10).
		{name: "bg-valid-alt-alt", value: "BGEIK.605000000", wantStatus: businessid.ValidationStatusValid},

		// CZ IČO cases r==0 and r==1.
		{name: "cz-valid-r0", value: "CZOR.00000001", wantStatus: businessid.ValidationStatusValid},
		{name: "cz-valid-r1", value: "CZOR.00000060", wantStatus: businessid.ValidationStatusValid},
		{name: "cz-valid-r10", value: "CZOR.00000051", wantStatus: businessid.ValidationStatusValid},

		// SK IČO cases r==0 and r==1 (same algorithm).
		{name: "sk-valid-r0", value: "SKOR.00000001", wantStatus: businessid.ValidationStatusValid},

		// ES CIF J-type (middle group, digit-only).
		{name: "es-valid-cif-j", value: "ESRMC.J12345674", wantStatus: businessid.ValidationStatusValid},
		// ES CIF P-type (letter-only).
		{name: "es-valid-cif-p", value: "ESRMC.P1234567D", wantStatus: businessid.ValidationStatusValid},

		// FR invalid Luhn (via native SIREN validator).
		{name: "fr-invalid-checksum", value: "FRRCS.552100550", wantStatus: businessid.ValidationStatusInvalid},

		// DK invalid (first digit 0).
		{name: "dk-invalid-first-zero", value: "DKCVR.01056416", wantStatus: businessid.ValidationStatusInvalid},

		// NL invalid (r==10 explosion).
		{name: "nl-invalid-r10", value: "NLKVK.04000000", wantStatus: businessid.ValidationStatusInvalid},

		// RO min length (2 digits).
		{name: "ro-valid-min", value: "ROORCT.19", wantStatus: businessid.ValidationStatusValid},

		// PT S mod 11 < 2 → check == 0.
		{name: "pt-valid-check-zero", value: "PTRN.000000000", wantStatus: businessid.ValidationStatusValid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateChecksum(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status, "reason=%s message=%s", res.ReasonCode, res.Message)
		})
	}
}
