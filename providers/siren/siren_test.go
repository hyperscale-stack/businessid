// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siren_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/siren"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := siren.New()

	assert.Equal(t, businessid.IdentifierKindSIREN, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: true, Registry: false}, p.Capabilities())
}

func TestCanonicalize(t *testing.T) {
	t.Parallel()

	p := siren.New()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "no-op", in: "552100554", want: "552100554"},
		{name: "spaces", in: "552 100 554", want: "552100554"},
		{name: "dots", in: "552.100.554", want: "552100554"},
		{name: "dashes", in: "552-100-554", want: "552100554"},
		{name: "leading-zero-preserved", in: "000-000-018", want: "000000018"},
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

	p := siren.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "too-short", value: "12345678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "too-long", value: "1234567890", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "letters", value: "12345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "digits-only", value: "552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "leading-zero", value: "000000018", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
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

	p := siren.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "known-valid-danone", value: "732 829 320", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "known-valid-lvmh", value: "552-100-554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "known-invalid", value: "732829321", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidChecksum},
		{name: "leading-zero-valid", value: "000000018", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
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
