// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package vat_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/vat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := vat.New()

	assert.Equal(t, businessid.IdentifierKindVAT, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: true, Registry: false}, p.Capabilities())
}

func TestCanonicalize(t *testing.T) {
	t.Parallel()

	p := vat.New()

	cases := []struct {
		name        string
		value       string
		countryCode string
		want        string
	}{
		{name: "already-prefixed", value: "FR44732829320", want: "FR44732829320"},
		{name: "lower", value: "fr44732829320", want: "FR44732829320"},
		{name: "spaces-dots-dashes", value: "FR 44.732-829.320", want: "FR44732829320"},
		{name: "prepend-country", value: "44732829320", countryCode: "fr", want: "FR44732829320"},
		{name: "no-country-no-prefix", value: "44732829320", want: "44732829320"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := p.Canonicalize(businessid.IdentifierInput{Value: tc.value, CountryCode: tc.countryCode})
			assert.Equal(t, tc.want, got.Value)
		})
	}
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := vat.New()

	cases := []struct {
		name        string
		value       string
		countryCode string
		wantStatus  businessid.ValidationStatus
		wantReason  string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "no-prefix-no-country", value: "44732829320", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonMissingCountryCode},
		{name: "country-mismatch", value: "FR44732829320", countryCode: "DE", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonCountryMismatch},

		// FR — France: 2 alnum key + 9 digit SIREN.
		// Source: https://ec.europa.eu/taxation_customs/vies/ (LVMH, SIREN 552100554).
		{name: "fr-valid", value: "FR44732829320", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-alnum-key", value: "FRAB732829320", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-invalid-length", value: "FR44732829", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "fr-invalid-chars", value: "FR44ABCDEFGHI", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "fr-key-non-alnum", value: "FR$$732829320", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// AT — Austria: U + 8 digits.
		// Source: Red Bull GmbH imprint (redbull.com/at-de/energydrink/company).
		{name: "at-valid", value: "ATU19277503", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "at-invalid-missing-u", value: "AT192775034", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "at-invalid-length", value: "ATU1927750", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},

		// BE — Belgium: 10 digits, first digit is 0 or 1.
		// Source: AB InBev SA/NV imprint (ab-inbev.com legal).
		{name: "be-valid", value: "BE0417497106", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "be-invalid-first-digit", value: "BE2417497106", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "be-invalid-length", value: "BE04174971", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},

		// BG — Bulgaria: 9 or 10 digits.
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "bg-valid-9", value: "BG123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "bg-valid-10", value: "BG1234567890", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "bg-invalid-length", value: "BG12345678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "bg-invalid-chars", value: "BG12345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// HR — Croatia: 11 digits (OIB).
		// Source: INA d.d. imprint (ina.hr).
		{name: "hr-valid", value: "HR27759560625", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "hr-invalid-length", value: "HR2775956062", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "hr-invalid-chars", value: "HR2775956062A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// CY — Cyprus: 8 digits + 1 letter.
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "cy-valid", value: "CY10000000P", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "cy-invalid-missing-letter", value: "CY100000000", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "cy-invalid-length", value: "CY1000000P", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},

		// CZ — Czech Republic: 8, 9 or 10 digits.
		// Source: Škoda Auto a.s. imprint (skoda-auto.com legal).
		{name: "cz-valid-8", value: "CZ00177041", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "cz-valid-9", value: "CZ001770410", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "cz-valid-10", value: "CZ0017704100", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "cz-invalid-length", value: "CZ1234567", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},

		// DE — Germany: 9 digits.
		// Source: SAP SE imprint (sap.com/impressum).
		{name: "de-valid", value: "DE143454214", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "de-invalid-length", value: "DE12345678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "de-invalid-chars", value: "DE12345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// DK — Denmark: 8 digits.
		// Source: Carlsberg A/S imprint (carlsberggroup.com).
		{name: "dk-valid", value: "DK61056416", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "dk-invalid-length", value: "DK6105641", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "dk-invalid-chars", value: "DK6105641A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// EE — Estonia: 9 digits.
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "ee-valid", value: "EE100305557", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "ee-invalid-length", value: "EE10030555", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "ee-invalid-chars", value: "EE10030555A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// EL — Greece: 9 digits (note: EL prefix, not GR).
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "el-valid", value: "EL094014298", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "el-invalid-length", value: "EL09401429", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "el-invalid-chars", value: "EL09401429A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		// GR (ISO country code) is aliased to EL by Canonicalize because
		// legacy ERPs and non-VIES systems still emit GR. The alias covers
		// both the value prefix and the CountryCode input field.
		{name: "gr-alias-value", value: "GR094014298", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "gr-alias-cc", value: "094014298", countryCode: "GR", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// ES — Spain: alnum + 7 digits + alnum. Prefix or suffix letter depends
		// on taxpayer type. Source: Santander (ESA39000013), Telefónica (ESA28015865).
		{name: "es-valid-leading-letter", value: "ESA39000013", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "es-valid-all-digits", value: "ES012345678", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "es-invalid-length", value: "ESA3900001", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "es-invalid-middle", value: "ESA390000A3", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// FI — Finland: 8 digits.
		// Source: Nokia Oyj imprint (nokia.com/about-us).
		{name: "fi-valid", value: "FI01120389", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fi-invalid-length", value: "FI0112038", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "fi-invalid-chars", value: "FI0112038A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// HU — Hungary: 8 digits.
		// Source: MOL Nyrt. imprint (molgroup.info).
		{name: "hu-valid", value: "HU10625790", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "hu-invalid-length", value: "HU1062579", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "hu-invalid-chars", value: "HU1062579A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// IE — Ireland: 7D+1L or 7D+2L (current) or D+L+5D+L (legacy).
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "ie-valid-7d1l", value: "IE1234567T", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "ie-valid-7d2l", value: "IE1234567TW", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "ie-valid-legacy", value: "IE1A23456T", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "ie-invalid-length", value: "IE12345T", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "ie-invalid-chars-8", value: "IE12345678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "ie-invalid-chars-9-suffix-digit", value: "IE12345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		// Legacy IE format (pre-2013): position-2 could be '+' or '*'.
		// Rejected by default; accepted with WithLegacy() in a separate test.
		{name: "ie-legacy-plus-rejected-default", value: "IE1+23456T", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "ie-legacy-star-rejected-default", value: "IE1*23456T", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// IT — Italy: 11 digits.
		// Source: Ferrari SpA imprint (ferrari.com/legal).
		{name: "it-valid", value: "IT03032310367", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "it-invalid-length", value: "IT0303231036", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "it-invalid-chars", value: "IT0303231036A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// LT — Lithuania: 9 or 12 digits.
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "lt-valid-9", value: "LT123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "lt-valid-12", value: "LT100002045110", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "lt-invalid-length-10", value: "LT1234567890", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "lt-invalid-chars", value: "LT12345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// LU — Luxembourg: 8 digits.
		// Source: Amazon EU S.à r.l. imprint (aboutamazon.com/eu).
		{name: "lu-valid", value: "LU19647148", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "lu-invalid-length", value: "LU1964714", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "lu-invalid-chars", value: "LU1964714A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// LV — Latvia: 11 digits.
		// Source: Latvenergo AS imprint (latvenergo.lv).
		{name: "lv-valid", value: "LV40003032949", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "lv-invalid-length", value: "LV4000303294", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "lv-invalid-chars", value: "LV4000303294A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// MT — Malta: 8 digits.
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "mt-valid", value: "MT12345678", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "mt-invalid-length", value: "MT1234567", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "mt-invalid-chars", value: "MT1234567A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// NL — Netherlands: 9 digits + 'B' + 2 digits.
		// Source: Koninklijke Philips N.V. imprint (philips.com).
		{name: "nl-valid", value: "NL800106004B01", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "nl-invalid-missing-b", value: "NL800106004001", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "nl-invalid-length", value: "NL800106004B0", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},

		// PL — Poland: 10 digits.
		// Source: PKN Orlen SA imprint (orlen.pl).
		{name: "pl-valid", value: "PL7740001454", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "pl-invalid-length", value: "PL774000145", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "pl-invalid-chars", value: "PL774000145A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// PT — Portugal: 9 digits.
		// Source: Galp Energia SGPS SA imprint (galp.com).
		{name: "pt-valid", value: "PT504499777", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "pt-invalid-length", value: "PT50449977", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "pt-invalid-chars", value: "PT50449977A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// RO — Romania: 2 to 10 digits.
		// Source: https://ec.europa.eu/taxation_customs/vies/ (format doc).
		{name: "ro-valid-min", value: "RO12", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "ro-valid-max", value: "RO1234567890", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "ro-invalid-length-too-short", value: "RO1", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "ro-invalid-length-too-long", value: "RO12345678901", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "ro-invalid-chars", value: "RO123456789A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// SE — Sweden: 12 digits (trailing "01" convention not enforced).
		// Source: AB Volvo imprint (volvogroup.com).
		{name: "se-valid", value: "SE556012579901", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "se-invalid-length", value: "SE55601257990", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "se-invalid-chars", value: "SE55601257990A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// SI — Slovenia: 8 digits.
		// Source: Krka d.d. imprint (krka.si).
		{name: "si-valid", value: "SI82646716", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "si-invalid-length", value: "SI8264671", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "si-invalid-chars", value: "SI8264671A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// SK — Slovakia: 10 digits.
		// Source: Slovnaft a.s. imprint (slovnaft.sk).
		{name: "sk-valid", value: "SK2020348904", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "sk-invalid-length", value: "SK202034890", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "sk-invalid-chars", value: "SK202034890A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// GB — United Kingdom: 9 or 12 digits.
		// Source: HMRC VAT format doc (gov.uk/vat-registration).
		{name: "gb-valid-9", value: "GB123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "gb-valid-12", value: "GB123456789012", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "gb-invalid-length-10", value: "GB1234567890", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "gb-invalid-chars", value: "GB12345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// XI — Northern Ireland: same as GB (post-Brexit protocol for EU goods).
		// Source: gov.uk/guidance/using-the-xi-prefix.
		{name: "xi-valid-9", value: "XI123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "xi-valid-12", value: "XI123456789012", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "xi-invalid-length", value: "XI12345678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},

		// NO — Norway: 9-digit organization number.
		// Source: Skatteetaten (skatteetaten.no) foretaksregisteret.
		{name: "no-valid", value: "NO923609016", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "no-invalid-length", value: "NO92360901", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "no-invalid-chars", value: "NO92360901A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// IS — Iceland: 5 or 6 digits (VSK number).
		// Source: skatturinn.is registration guide.
		{name: "is-valid-5", value: "IS12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "is-valid-6", value: "IS123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "is-invalid-length", value: "IS1234", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "is-invalid-chars", value: "IS1234A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// LI — Liechtenstein: 5-digit internal enterprise number.
		// Source: Liechtenstein Tax Administration (stv.li).
		{name: "li-valid", value: "LI12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "li-invalid-length", value: "LI1234", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "li-invalid-chars", value: "LI1234A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// Generic fallback: unknown 2-letter prefix falls back to
		// 2..13 alphanumeric (preserves prior behaviour for exotic prefixes).
		{name: "generic-valid", value: "ZZ123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "generic-too-short", value: "ZZ1", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "generic-too-long", value: "ZZ12345678901234", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "generic-non-alnum", value: "ZZ123456*89", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateFormat(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value, CountryCode: tc.countryCode}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, tc.wantReason, res.ReasonCode)
		})
	}
}

func TestValidateFormatWithLegacy(t *testing.T) {
	t.Parallel()

	p := vat.New(vat.WithLegacy())

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
	}{
		// Position-2 '+' and '*' were valid in the pre-2013 Irish VAT layout
		// (digit + [A-Z+*] + 5 digits + letter). Source: Revenue.ie legacy
		// numbering guidance.
		{name: "ie-legacy-plus", value: "IE1+23456T", wantStatus: businessid.ValidationStatusValid},
		{name: "ie-legacy-star", value: "IE1*23456T", wantStatus: businessid.ValidationStatusValid},
		{name: "ie-new-still-valid", value: "IE1234567T", wantStatus: businessid.ValidationStatusValid},
		{name: "ie-invalid-still-rejected", value: "IE1@23456T", wantStatus: businessid.ValidationStatusInvalid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateFormat(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status)
		})
	}
}

func TestValidateChecksum(t *testing.T) {
	t.Parallel()

	p := vat.New()

	cases := []struct {
		name        string
		value       string
		countryCode string
		wantStatus  businessid.ValidationStatus
		wantReason  string
	}{
		{name: "fr-valid-numeric-key", value: "FR44732829320", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-valid-numeric-key-lvmh", value: "FR96552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-bad-numeric-key", value: "FR45732829320", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		{name: "fr-bad-siren-luhn", value: "FR44732829321", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		{name: "fr-alnum-key-unsupported", value: "FRAB732829320", wantStatus: businessid.ValidationStatusUnsupported, wantReason: businessid.ReasonUnsupportedChecksum},
		// DE now has a checksum implementation; the synthetic value fails
		// ISO 7064 MOD 11,10 as expected.
		{name: "de-synthetic-fails-checksum", value: "DE123456789", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		// LI has no checksum → still Unsupported.
		{name: "li-unsupported", value: "LI12345", wantStatus: businessid.ValidationStatusUnsupported, wantReason: businessid.ReasonUnsupportedChecksum},
		{name: "no-prefix-at-all", value: "44732829320", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonMissingCountryCode},
		{name: "fr-wrong-length-checksum", value: "FR44732829", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "fr-alnum-key-bad-siren-still-unsupported", value: "FRAB732829321", wantStatus: businessid.ValidationStatusUnsupported, wantReason: businessid.ReasonUnsupportedChecksum},
		{name: "empty-with-country", value: "", countryCode: "FR", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "checksum-country-mismatch", value: "FR96552100554", countryCode: "DE", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonCountryMismatch},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateChecksum(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value, CountryCode: tc.countryCode}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, tc.wantReason, res.ReasonCode)
		})
	}
}

// TestValidateChecksumPerCountry exercises the per-country checksum
// algorithms with vectors verified against python-stdnum test suite
// and national tax authorities. Each country covers ≥1 valid case
// (real-world or documented sample) and 1 obviously-invalid (off-by-one).
func TestValidateChecksumPerCountry(t *testing.T) {
	t.Parallel()

	p := vat.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
	}{
		// AT — Source: Wikipedia AT VAT sample + python-stdnum.
		{name: "at-valid-1", value: "ATU13585627", wantStatus: businessid.ValidationStatusValid},
		{name: "at-valid-2", value: "ATU10223006", wantStatus: businessid.ValidationStatusValid},
		{name: "at-invalid", value: "ATU13585628", wantStatus: businessid.ValidationStatusInvalid},

		// BE — Source: SPF Finances (AB InBev imprint + python-stdnum sample).
		{name: "be-valid-abi", value: "BE0417497106", wantStatus: businessid.ValidationStatusValid},
		{name: "be-valid-2", value: "BE0428759497", wantStatus: businessid.ValidationStatusValid},
		{name: "be-invalid", value: "BE0428759498", wantStatus: businessid.ValidationStatusInvalid},

		// BG — Source: NRA + python-stdnum.
		{name: "bg-valid-9d", value: "BG040808527", wantStatus: businessid.ValidationStatusValid},
		{name: "bg-valid-10d", value: "BG7523169263", wantStatus: businessid.ValidationStatusValid},
		{name: "bg-invalid", value: "BG040808520", wantStatus: businessid.ValidationStatusInvalid},

		// HR — Source: INA d.d. imprint.
		{name: "hr-valid-ina", value: "HR27759560625", wantStatus: businessid.ValidationStatusValid},
		{name: "hr-invalid", value: "HR27759560624", wantStatus: businessid.ValidationStatusInvalid},

		// CY — Source: python-stdnum sample.
		{name: "cy-valid", value: "CY10259033P", wantStatus: businessid.ValidationStatusValid},
		{name: "cy-invalid", value: "CY10259033A", wantStatus: businessid.ValidationStatusInvalid},

		// CZ — Source: python-stdnum + railway operator.
		{name: "cz-valid-1", value: "CZ25123891", wantStatus: businessid.ValidationStatusValid},
		{name: "cz-valid-cd", value: "CZ00006947", wantStatus: businessid.ValidationStatusValid},
		{name: "cz-invalid", value: "CZ25123890", wantStatus: businessid.ValidationStatusInvalid},

		// DE — Source: SAP imprint + Wikipedia DE VAT sample.
		{name: "de-valid-sap", value: "DE143454214", wantStatus: businessid.ValidationStatusValid},
		{name: "de-valid-2", value: "DE136695976", wantStatus: businessid.ValidationStatusValid},
		{name: "de-invalid", value: "DE136695970", wantStatus: businessid.ValidationStatusInvalid},

		// DK — Source: Carlsberg imprint + Wikipedia sample.
		{name: "dk-valid-carlsberg", value: "DK61056416", wantStatus: businessid.ValidationStatusValid},
		{name: "dk-valid-2", value: "DK13585628", wantStatus: businessid.ValidationStatusValid},
		{name: "dk-invalid", value: "DK13585620", wantStatus: businessid.ValidationStatusInvalid},

		// EE — Source: python-stdnum verified sample.
		{name: "ee-valid", value: "EE100931558", wantStatus: businessid.ValidationStatusValid},
		{name: "ee-invalid", value: "EE100931550", wantStatus: businessid.ValidationStatusInvalid},

		// EL — Source: National Bank of Greece imprint.
		{name: "el-valid-nbg", value: "EL094014201", wantStatus: businessid.ValidationStatusValid},
		{name: "el-invalid", value: "EL094014200", wantStatus: businessid.ValidationStatusInvalid},

		// ES — Source: Santander SA imprint.
		{name: "es-valid-santander", value: "ESA39000013", wantStatus: businessid.ValidationStatusValid},
		{name: "es-invalid", value: "ESA39000010", wantStatus: businessid.ValidationStatusInvalid},

		// FI — Source: Nokia imprint + python-stdnum.
		{name: "fi-valid-nokia", value: "FI01120389", wantStatus: businessid.ValidationStatusValid},
		{name: "fi-valid-2", value: "FI09853608", wantStatus: businessid.ValidationStatusValid},
		{name: "fi-invalid", value: "FI09853600", wantStatus: businessid.ValidationStatusInvalid},

		// HU — Source: MOL imprint + python-stdnum.
		{name: "hu-valid-mol", value: "HU10625790", wantStatus: businessid.ValidationStatusValid},
		{name: "hu-valid-2", value: "HU12892312", wantStatus: businessid.ValidationStatusValid},
		{name: "hu-invalid", value: "HU12892310", wantStatus: businessid.ValidationStatusInvalid},

		// IE — Source: Google Ireland imprint + python-stdnum.
		{name: "ie-valid-google", value: "IE6388047V", wantStatus: businessid.ValidationStatusValid},
		{name: "ie-valid-2l", value: "IE9825613N", wantStatus: businessid.ValidationStatusValid},
		{name: "ie-invalid", value: "IE6388047X", wantStatus: businessid.ValidationStatusInvalid},

		// IT — Source: Stellantis imprint + python-stdnum.
		{name: "it-valid-stellantis", value: "IT07973780013", wantStatus: businessid.ValidationStatusValid},
		{name: "it-valid-stdnum", value: "IT12345670017", wantStatus: businessid.ValidationStatusValid},
		{name: "it-invalid", value: "IT12345670010", wantStatus: businessid.ValidationStatusInvalid},

		// LT — Source: python-stdnum synthetic vectors.
		{name: "lt-valid-9d", value: "LT100000006", wantStatus: businessid.ValidationStatusValid},
		{name: "lt-invalid", value: "LT100000005", wantStatus: businessid.ValidationStatusInvalid},

		// LU — Source: Amazon EU + python-stdnum.
		{name: "lu-valid-amazon", value: "LU19647148", wantStatus: businessid.ValidationStatusValid},
		{name: "lu-valid-2", value: "LU10000356", wantStatus: businessid.ValidationStatusValid},
		{name: "lu-invalid", value: "LU10000350", wantStatus: businessid.ValidationStatusInvalid},

		// LV — Source: Latvenergo imprint.
		{name: "lv-valid-latvenergo", value: "LV40003032949", wantStatus: businessid.ValidationStatusValid},
		{name: "lv-invalid", value: "LV40003032940", wantStatus: businessid.ValidationStatusInvalid},

		// MT — Source: python-stdnum.
		{name: "mt-valid", value: "MT12345634", wantStatus: businessid.ValidationStatusValid},
		{name: "mt-invalid", value: "MT12345630", wantStatus: businessid.ValidationStatusInvalid},

		// NL — Source: python-stdnum vectors.
		{name: "nl-valid-1", value: "NL010000446B01", wantStatus: businessid.ValidationStatusValid},
		{name: "nl-valid-2", value: "NL196117306B01", wantStatus: businessid.ValidationStatusValid},
		{name: "nl-invalid", value: "NL010000440B01", wantStatus: businessid.ValidationStatusInvalid},

		// PL — Source: PGE imprint.
		{name: "pl-valid-pge", value: "PL5261040567", wantStatus: businessid.ValidationStatusValid},
		{name: "pl-invalid", value: "PL5261040560", wantStatus: businessid.ValidationStatusInvalid},

		// PT — Source: EDP imprint + python-stdnum.
		{name: "pt-valid-edp", value: "PT500697256", wantStatus: businessid.ValidationStatusValid},
		{name: "pt-valid-2", value: "PT501964843", wantStatus: businessid.ValidationStatusValid},
		{name: "pt-invalid", value: "PT501964840", wantStatus: businessid.ValidationStatusInvalid},

		// RO — Source: python-stdnum.
		{name: "ro-valid", value: "RO14186770", wantStatus: businessid.ValidationStatusValid},
		{name: "ro-invalid", value: "RO14186771", wantStatus: businessid.ValidationStatusInvalid},

		// SE — Source: python-stdnum sample.
		{name: "se-valid", value: "SE123456789701", wantStatus: businessid.ValidationStatusValid},
		{name: "se-invalid", value: "SE123456789801", wantStatus: businessid.ValidationStatusInvalid},

		// SI — Source: Krka imprint + python-stdnum.
		{name: "si-valid-krka", value: "SI82646716", wantStatus: businessid.ValidationStatusValid},
		{name: "si-valid-2", value: "SI50223054", wantStatus: businessid.ValidationStatusValid},
		{name: "si-invalid", value: "SI50223050", wantStatus: businessid.ValidationStatusInvalid},

		// SK — Source: python-stdnum sample.
		{name: "sk-valid", value: "SK2022749619", wantStatus: businessid.ValidationStatusValid},
		{name: "sk-invalid", value: "SK2022749610", wantStatus: businessid.ValidationStatusInvalid},

		// GB — Source: python-stdnum (9755-algo sample).
		{name: "gb-valid", value: "GB562235987", wantStatus: businessid.ValidationStatusValid},
		{name: "gb-invalid", value: "GB562235980", wantStatus: businessid.ValidationStatusInvalid},

		// XI — Same algorithms as GB.
		{name: "xi-valid", value: "XI562235987", wantStatus: businessid.ValidationStatusValid},
		{name: "xi-invalid", value: "XI562235980", wantStatus: businessid.ValidationStatusInvalid},
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
