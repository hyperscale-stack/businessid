// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ein_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/ein"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := ein.New()

	assert.Equal(t, businessid.IdentifierKindEIN, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: false, Registry: false}, p.Capabilities())
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := ein.New()

	cases := []struct {
		name       string
		value      string
		wantStatus businessid.ValidationStatus
		wantReason string
	}{
		{name: "empty", value: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "canonical", value: "123456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "with-dash", value: "12-3456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "with-space", value: "12 3456789", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "leading-zero", value: "01-2345678", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "letters", value: "12-345678A", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
		{name: "too-short", value: "12345678", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
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
