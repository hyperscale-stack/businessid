// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package lei_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/lei"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := lei.New()

	assert.Equal(t, businessid.IdentifierKindLEI, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: true, Registry: false}, p.Capabilities())
}

func TestCanonicalize(t *testing.T) {
	t.Parallel()

	p := lei.New()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "no-op", in: "529900T8BM49AURSDO55", want: "529900T8BM49AURSDO55"},
		{name: "lower", in: "529900t8bm49aursdo55", want: "529900T8BM49AURSDO55"},
		{name: "spaces", in: " 529900 T8BM49 AURSDO55 ", want: "529900T8BM49AURSDO55"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := p.Canonicalize(businessid.IdentifierInput{Value: tc.in})
			assert.Equal(t, tc.want, got.Value)
		})
	}
}

func TestValidateFormatAndChecksum(t *testing.T) {
	t.Parallel()

	p := lei.New()

	cases := []struct {
		name             string
		value            string
		wantFormatStatus businessid.ValidationStatus
		wantFormatReason string
		wantChecksum     businessid.ValidationStatus
		wantChecksumCode string
	}{
		{
			name:             "known-valid-deutsche-bank",
			value:            "529900T8BM49AURSDO55",
			wantFormatStatus: businessid.ValidationStatusValid,
			wantFormatReason: businessid.ReasonOK,
			wantChecksum:     businessid.ValidationStatusValid,
			wantChecksumCode: businessid.ReasonOK,
		},
		{
			name:             "known-valid-nordea",
			value:            "6SCPQ280AIY8EP3XFW53",
			wantFormatStatus: businessid.ValidationStatusValid,
			wantFormatReason: businessid.ReasonOK,
			wantChecksum:     businessid.ValidationStatusValid,
			wantChecksumCode: businessid.ReasonOK,
		},
		{
			name:             "known-invalid-check",
			value:            "529900T8BM49AURSDO56",
			wantFormatStatus: businessid.ValidationStatusValid,
			wantFormatReason: businessid.ReasonOK,
			wantChecksum:     businessid.ValidationStatusInvalid,
			wantChecksumCode: businessid.ReasonInvalidChecksum,
		},
		{
			name:             "empty",
			value:            "",
			wantFormatStatus: businessid.ValidationStatusInvalid,
			wantFormatReason: businessid.ReasonEmpty,
			wantChecksum:     businessid.ValidationStatusInvalid,
			wantChecksumCode: businessid.ReasonEmpty,
		},
		{
			name:             "too-short",
			value:            "529900T8BM49AURSDO5",
			wantFormatStatus: businessid.ValidationStatusInvalid,
			wantFormatReason: businessid.ReasonInvalidLength,
			wantChecksum:     businessid.ValidationStatusInvalid,
			wantChecksumCode: businessid.ReasonInvalidChecksum,
		},
		{
			name:             "invalid-char",
			value:            "529900T8BM49AURSDO5-",
			wantFormatStatus: businessid.ValidationStatusInvalid,
			wantFormatReason: businessid.ReasonInvalidCharacters,
			wantChecksum:     businessid.ValidationStatusInvalid,
			wantChecksumCode: businessid.ReasonInvalidChecksum,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			canon := p.Canonicalize(businessid.IdentifierInput{Value: tc.value})

			f, err := p.ValidateFormat(context.Background(), canon)
			require.NoError(t, err)
			assert.Equal(t, tc.wantFormatStatus, f.Status)
			assert.Equal(t, tc.wantFormatReason, f.ReasonCode)

			c, err := p.ValidateChecksum(context.Background(), canon)
			require.NoError(t, err)
			assert.Equal(t, tc.wantChecksum, c.Status)
			assert.Equal(t, tc.wantChecksumCode, c.ReasonCode)
		})
	}
}
