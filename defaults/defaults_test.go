// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package defaults_test

import (
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviders(t *testing.T) {
	t.Parallel()

	ps := defaults.Providers()
	require.NotEmpty(t, ps)

	seen := make(map[businessid.IdentifierKind]struct{}, len(ps))

	for _, p := range ps {
		seen[p.Kind()] = struct{}{}
	}

	for _, k := range []businessid.IdentifierKind{
		businessid.IdentifierKindSIREN,
		businessid.IdentifierKindSIRET,
		businessid.IdentifierKindLEI,
		businessid.IdentifierKindDUNS,
		businessid.IdentifierKindEIN,
		businessid.IdentifierKindCompanyNumber,
		businessid.IdentifierKindEUID,
		businessid.IdentifierKindEORI,
		businessid.IdentifierKindVAT,
		businessid.IdentifierKindNationalRegistrationNumber,
	} {
		_, ok := seen[k]
		assert.Truef(t, ok, "missing default provider for %q", k)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	v := defaults.New()
	require.NotNil(t, v)

	_, ok := v.Provider(businessid.IdentifierKindSIREN)
	assert.True(t, ok)
}

func TestWithAll(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(defaults.WithAll())

	_, ok := v.Provider(businessid.IdentifierKindLEI)
	assert.True(t, ok)
}
