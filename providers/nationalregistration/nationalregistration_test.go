// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package nationalregistration_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/nationalregistration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCapabilities(t *testing.T) {
	t.Parallel()

	p := nationalregistration.New()

	assert.Equal(t, businessid.IdentifierKindNationalRegistrationNumber, p.Kind())
	assert.Equal(t, businessid.Capabilities{Format: true, Checksum: false, Registry: false}, p.Capabilities())
}

func TestValidateFormat(t *testing.T) {
	t.Parallel()

	p := nationalregistration.New()

	cases := []struct {
		name        string
		value       string
		countryCode string
		wantStatus  businessid.ValidationStatus
		wantReason  string
	}{
		{name: "empty-country", value: "ABC123", countryCode: "", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonMissingCountryCode},
		{name: "empty-value", value: "", countryCode: "BE", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonEmpty},
		{name: "simple", value: "ABC.123-45", countryCode: "BE", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "lower-canonicalized", value: "abc-123", countryCode: "be", wantStatus: businessid.ValidationStatusValid, wantReason: businessid.ReasonOK},
		{name: "too-long", value: strings.Repeat("A", 65), countryCode: "BE", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidLength},
		{name: "invalid-char", value: "ABC*123", countryCode: "BE", wantStatus: businessid.ValidationStatusInvalid, wantReason: businessid.ReasonInvalidCharacters},
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
