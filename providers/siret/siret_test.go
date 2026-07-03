// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siret_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/siret"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := siret.New()

	assert.Equal(t, businessid.IdentifierKindSIRET, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: true, Registry: false}, p.Capabilities())
}

func TestCanonicalize(t *testing.T) {
	t.Parallel()

	p := siret.New()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "spaces", in: "732 829 320 00074", want: "73282932000074"},
		{name: "dashes", in: "732-829-320-000-74", want: "73282932000074"},
		{name: "no-op", in: "73282932000074", want: "73282932000074"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := p.Canonicalize(businessid.IdentifierInput{Value: tc.in})
			assert.Equal(t, tc.want, got.Value)
		})
	}
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := siret.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "too-short", value: "1234567890123", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "too-long", value: "123456789012345", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "letters", value: "1234567890123A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "letter-mid-string", value: "7328A932000074", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "underscore", value: "1234567890123_", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		// NIC 00000 is reserved by INSEE and never attributed.
		// Source: https://www.insee.fr/fr/information/2408687 (identifiants Sirene).
		{name: "nic-00000-reserved", value: "73282932000000", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "digits-only", value: "73282932000074", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateFormat(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, tc.wantReason, res.ReasonCode)
		})
	}
}

func TestValidateChecksum(t *testing.T) {
	t.Parallel()

	p := siret.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		// Real-world head-office SIRETs (SIREN + NIC 0007x / 0001x) sourced
		// from company legal imprints and https://annuaire-entreprises.data.gouv.fr/.
		{name: "known-valid", value: "73282932000074", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "known-valid-lvmh", value: "55210055400013", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "known-valid-totalenergies", value: "54205118000074", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "known-valid-renault", value: "44163946500018", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},

		// Off-by-one on the NIC.
		{name: "known-invalid", value: "73282932000075", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		{name: "lvmh-off-by-one", value: "55210055400014", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},

		// La Poste (SIREN 356000000). A SIRET is accepted when either
		// Luhn passes (many head-office SIRETs do) or the plain digit
		// sum is divisible by 5 (INSEE dérogatoire rule for La Poste
		// establishments).
		{name: "la-poste-luhn-valid", value: "35600000000048", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "la-poste-rule-only", value: "35600000000060", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		// A SIRET starting with 356000000 that satisfies neither rule
		// is still rejected.
		{name: "la-poste-both-rules-fail", value: "35600000000002", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		// The La Poste rule only applies to SIREN 356000000; another
		// digit-sum-divisible-by-5 SIRET still needs standard Luhn.
		{name: "non-la-poste-digit-sum-only", value: "12345678900032", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},

		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateChecksum(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, tc.wantReason, res.ReasonCode)
		})
	}
}

func TestValidateChecksumWithDerogation(t *testing.T) {
	t.Parallel()

	// Rule accepts any SIRET whose 14 digits sum to a multiple of 7. Only
	// applied to SIRETs whose SIREN prefix is 999999999 (a made-up value
	// used only to exercise the WithDerogation plumbing).
	divisibleBy7 := func(s string) bool {
		sum := 0
		for i := range len(s) {
			sum += int(s[i] - '0')
		}

		return sum%7 == 0
	}

	p := siret.New(siret.WithDerogation("999999999", divisibleBy7))

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
	}{
		// digit sum = 81 + 3 = 84 → divisible by 7 → accepted by the rule
		// even though Luhn fails.
		{name: "custom-rule-accepts", value: "99999999900003", wantStatus: businessid.ValidationStatusValid},
		// digit sum = 81 + 4 = 85 → not divisible; both Luhn and rule fail.
		{name: "custom-rule-rejects", value: "99999999900004", wantStatus: businessid.ValidationStatusInvalid},
		// Different SIREN prefix: rule not consulted; falls back to Luhn only.
		{name: "different-prefix-uses-luhn", value: "73282932000074", wantStatus: businessid.ValidationStatusValid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := p.ValidateChecksum(context.Background(), p.Canonicalize(businessid.IdentifierInput{Value: tc.value}))
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tc.wantStatus, res.Status)
		})
	}
}
