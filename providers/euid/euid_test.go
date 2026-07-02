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
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: false, Registry: false}, p.Capabilities())
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
		{name: "valid-simple", value: "FRRCS.552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "valid-with-separators", value: "DEHRB.HAMBURG/B-12345", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "lower-canonicalized", value: "frrcs.552100554", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
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
