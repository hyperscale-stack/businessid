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
	"github.com/hyperscale-stack/businessid/providers/siren"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := euid.New()

	assert.Equal(t, businessid.IdentifierKindEUID, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: true, Registry: false}, p.Capabilities())
}

func TestValidateFormatWithSIRENSubValidator(t *testing.T) {
	t.Parallel()

	// EUID configured as the defaults package would wire it: SIREN
	// sub-validator injected for FR.
	p := euid.New(euid.WithSubValidator(siren.New()))

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		// FR EUID: registration must be a 9-digit SIREN. Basic BRIS layout
		// passes but SIREN sub-validator rejects non-digit / wrong-length.
		{name: "fr-valid-siren-shape", value: "FRRCS.552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-invalid-registration-alpha", value: "FRRCS.ABC", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "fr-invalid-registration-too-short", value: "FRRCS.55210055", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "fr-invalid-registration-non-digit", value: "FRRCS.55210055A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},

		// Non-FR EUID falls through to generic BRIS check (no sub-validator
		// configured for DE), so meta-format shape is enough.
		{name: "de-shape-ok", value: "DEHRB.HAMBURG/B-12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// Plus sign is now allowed in registration charset.
		{name: "registration-with-plus", value: "IECRO.12+34", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
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

func TestValidateChecksumViaSubValidator(t *testing.T) {
	t.Parallel()

	p := euid.New(euid.WithSubValidator(siren.New()))

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		// LVMH SIREN 552100554 → Luhn valid → EUID checksum valid.
		{name: "fr-valid-luhn", value: "FRRCS.552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// Off-by-one on the SIREN → Luhn fails → EUID checksum fails.
		{name: "fr-invalid-luhn", value: "FRRCS.552100555", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		// DE has no sub-validator configured → Unsupported.
		{name: "de-unsupported", value: "DEHRB.HAMBURG/B-12345", wantStatus: businessid.ValidationStatusUnsupported, wantReason: businessid.ReasonUnsupportedChecksum},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateChecksum(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, tc.wantReason, res.ReasonCode)
		})
	}
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := euid.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "lower-canonicalized", value: "frrcs.552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// EU-27 valid EUIDs. Register codes are those documented for the
		// national business register in the BRIS specification (Regulation
		// (EU) 2015/884). Registration segments are shape-correct examples
		// derived from public company data on each national register.

		// AT — Firmenbuch (FN). Source: justiz.gv.at/firmenbuch.
		{name: "at-valid", value: "ATFN.123456A", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// BE — Banque-Carrefour des Entreprises (BCE). Source: kbopub.economie.fgov.be.
		{name: "be-valid", value: "BEBCE.0417497106", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// BG — Trade Register EIK. Source: brra.bg.
		{name: "bg-valid", value: "BGEIK.123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// HR — Court register MBS. Source: sudreg.pravosudje.hr.
		{name: "hr-valid", value: "HRMBS.080000001", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// CY — Registrar of Companies HE. Source: efiling.drcor.mcit.gov.cy.
		{name: "cy-valid", value: "CYHE.123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// CZ — Obchodní rejstřík (OR). Source: or.justice.cz.
		{name: "cz-valid", value: "CZOR.123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// DE — Handelsregister B (HRB) with local court. Source: unternehmensregister.de.
		{name: "de-valid", value: "DEHRB.HAMBURG/B-12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// DK — Central Business Register (CVR). Source: cvr.dk.
		{name: "dk-valid", value: "DKCVR.61056416", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// EE — Business Register (RIK). Source: ariregister.rik.ee.
		{name: "ee-valid", value: "EERIK.10030555", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// EL — General Commercial Registry (GEMI). Source: businessregistry.gr.
		{name: "el-valid", value: "ELGEMI.123456789000", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// ES — Registro Mercantil Central (RMC). Source: rmc.es.
		{name: "es-valid", value: "ESRMC.M12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// FI — Patentti- ja rekisterihallitus (PRH). Source: prh.fi.
		{name: "fi-valid", value: "FIPRH.1120389", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// FR — Registre du Commerce et des Sociétés (RCS). Source: infogreffe.fr.
		{name: "fr-valid", value: "FRRCS.552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// HU — Cégjegyzék (CG). Source: e-cegjegyzek.hu.
		{name: "hu-valid", value: "HUCG.01-09-123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// IE — Companies Registration Office (CRO). Source: cro.ie.
		{name: "ie-valid", value: "IECRO.123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// IT — Registro Imprese (RI). Source: registroimprese.it.
		{name: "it-valid", value: "ITRI.MI-1234567", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// LT — Juridinių asmenų registras (JAR). Source: registrucentras.lt.
		{name: "lt-valid", value: "LTJAR.123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// LU — Registre de Commerce et des Sociétés (RCSL). Source: lbr.lu.
		{name: "lu-valid", value: "LURCSL.B12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// LV — Uzņēmumu Reģistrs (URE). Source: ur.gov.lv.
		{name: "lv-valid", value: "LVURE.40003032949", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// MT — Malta Business Registry (MBR). Source: mbr.mt.
		{name: "mt-valid", value: "MTMBR.C12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// NL — Kamer van Koophandel (KVK). Source: kvk.nl.
		{name: "nl-valid", value: "NLKVK.12345678", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// PL — Krajowy Rejestr Sądowy (KRS). Source: krs.ms.gov.pl.
		{name: "pl-valid", value: "PLKRS.0000123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// PT — Registo Nacional de Pessoas Colectivas (RN). Source: portaldocidadao.pt.
		{name: "pt-valid", value: "PTRN.504499777", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// RO — Oficiul Naţional al Registrului Comerţului (ORCT). Source: onrc.ro.
		{name: "ro-valid", value: "ROORCT.J40/12345/2020", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// SE — Bolagsverket (BR). Source: bolagsverket.se.
		{name: "se-valid", value: "SEBR.556012-5799", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// SI — Sodni register (SRG). Source: ejn.gov.si.
		{name: "si-valid", value: "SISRG.1234567000", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// SK — Obchodný register (OR). Source: orsr.sk.
		{name: "sk-valid", value: "SKOR.SR12345B", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// Non-EU codes (XI, GB, NO, IS, LI) are not part of BRIS but are
		// accepted by the generic BRIS-shaped validator.
		{name: "xi-valid", value: "XICH.NI123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "gb-valid", value: "GBCH.12345678", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "no-valid", value: "NOBRREG.923609016", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "is-valid", value: "ISRSK.1234567", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "li-valid", value: "LIHR.FL0001234567", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// Per-country invalid cases: one representative bad shape per country
		// (empty registration, non-alnum in register, char outside the
		// registration charset, or length overflow).
		{name: "at-invalid-empty-registration", value: "ATFN.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "be-invalid-register-non-alnum", value: "BEB-CE.0417497106", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "bg-invalid-registration-underscore", value: "BGEIK.ABC_123", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "hr-invalid-no-dot", value: "HRMBS080000001", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "cy-invalid-bad-country", value: "C1HE.123456", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "cz-invalid-registration-too-long", value: "CZOR." + strings.Repeat("A", 65), wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "de-invalid-registration-invalid-char", value: "DEHRB.HAMBURG@B-12345", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "dk-invalid-register-non-alnum", value: "DKCV_R.61056416", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "ee-invalid-empty-registration", value: "EERIK.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "el-invalid-register-non-alnum", value: "ELGE!MI.123456789", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "es-invalid-registration-invalid-char", value: "ESRMC.M12*45", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "fi-invalid-register-non-alnum", value: "FIPR_H.1120389", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "fr-invalid-registration-underscore", value: "FRRCS.552_100_554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "hu-invalid-registration-invalid-char", value: "HUCG.01_09_123456", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "ie-invalid-bad-country", value: "I3CRO.123456", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "it-invalid-registration-invalid-char", value: "ITRI.MI_1234567", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "lt-invalid-empty-registration", value: "LTJAR.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "lu-invalid-register-too-long", value: "LU" + strings.Repeat("A", 21) + ".B12345", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "lv-invalid-register-non-alnum", value: "LVU-RE.40003032949", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "mt-invalid-empty-registration", value: "MTMBR.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "nl-invalid-registration-invalid-char", value: "NLKVK.1234@5678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "pl-invalid-register-non-alnum", value: "PL@KRS.0000123456", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "pt-invalid-empty-registration", value: "PTRN.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "ro-invalid-registration-invalid-char", value: "ROORCT.J40_12345_2020", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "se-invalid-empty-registration", value: "SEBR.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "si-invalid-register-non-alnum", value: "SISR-G.1234567000", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "sk-invalid-empty-registration", value: "SKOR.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},

		// Cross-cutting shape errors.
		{name: "no-dot", value: "FRRCS552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "bad-country", value: "F1RCS.552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "register-non-alnum", value: "FRR-CS.552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "empty-registration", value: "FRRCS.", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "register-too-long", value: "FRVERYLONGREGISTERNAMEXX.552100554", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "registration-too-long", value: "FRRCS." + strings.Repeat("A", 65), wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "registration-invalid-char", value: "FRRCS.ABC_123", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateFormat(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, tc.wantReason, res.ReasonCode)
		})
	}
}
