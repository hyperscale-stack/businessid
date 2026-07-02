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
		{name: "known-valid", value: "73282932000074", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "known-invalid", value: "73282932000075", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
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
