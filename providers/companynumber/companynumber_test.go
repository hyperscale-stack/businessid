// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package companynumber_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/companynumber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := companynumber.New()

	assert.Equal(t, businessid.IdentifierKindCompanyNumber, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: false, Registry: false}, p.Capabilities())
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := companynumber.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "eight-digits", value: "12345678", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "leading-zero-digits", value: "00000006", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "scotland-prefix", value: "SC123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "scotland-lower", value: "sc123456", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "ni-prefix", value: "NI654321", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "spaces", value: "  SC 123 456 ", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "too-short", value: "1234567", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "unknown-prefix", value: "ZZ123456", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidFormat},
		{name: "prefix-non-digit-suffix", value: "SC12345X", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
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
