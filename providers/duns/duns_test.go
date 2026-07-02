// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package duns_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/duns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := duns.New()

	assert.Equal(t, businessid.IdentifierKindDUNS, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: false, Registry: false}, p.Capabilities())
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := duns.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "spaces-removed", value: "12 345 678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "valid", value: "123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "dashes-and-dots-stripped", value: "12-34.567-89", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "leading-zero", value: "003456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "too-short", value: "12345678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "letters", value: "12345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
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
