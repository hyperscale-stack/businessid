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
