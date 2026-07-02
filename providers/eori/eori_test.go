// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package eori_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/eori"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := eori.New()

	assert.Equal(t, businessid.IdentifierKindEORI, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: false, Registry: false}, p.Capabilities())
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := eori.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "typical", value: "FR12345678900", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "min-length", value: "FRA", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "max-length", value: "GB123456789012345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "lower", value: "fr12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "spaces", value: " FR 12 345 ", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "too-short", value: "FR", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "too-long", value: "GB1234567890123456", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "digit-prefix", value: "12ABCDE", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "bad-suffix", value: "FR12-34", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
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
