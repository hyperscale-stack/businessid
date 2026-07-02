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
		{name: "fr-valid", value: "FR44732829320", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-alnum-key", value: "FRAB732829320", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-wrong-length", value: "FR44732829", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "fr-bad-siren", value: "FR44ABCDEFGHI", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "generic-valid", value: "DE123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "generic-too-short", value: "DE1", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "country-mismatch", value: "FR44732829320", countryCode: "DE", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonCountryMismatch},
		{name: "generic-too-long", value: "DE12345678901234", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "generic-non-alnum", value: "DE123456*89", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "fr-key-non-alnum", value: "FR$$732829320", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
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

func TestValidateChecksum(t *testing.T) {
	t.Parallel()

	p := vat.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "fr-valid-numeric-key", value: "FR44732829320", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-valid-numeric-key-lvmh", value: "FR96552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "fr-bad-numeric-key", value: "FR45732829320", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		{name: "fr-bad-siren-luhn", value: "FR44732829321", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		{name: "fr-alnum-key-unsupported", value: "FRAB732829320", wantStatus: businessid.ValidationStatusUnsupported, wantReason: businessid.ReasonUnsupportedChecksum},
		{name: "non-fr-unsupported", value: "DE123456789", wantStatus: businessid.ValidationStatusUnsupported, wantReason: businessid.ReasonUnsupportedChecksum},
		{name: "no-prefix-at-all", value: "44732829320", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonMissingCountryCode},
		{name: "fr-wrong-length-checksum", value: "FR44732829", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
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
