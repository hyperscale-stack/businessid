// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siret_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/siret"
)

func FuzzValidateFormat(f *testing.F) {
	seeds := []string{
		"", "73282932000074", "35600000000060", "55210055400013",
		"73282932000000", "1234567890123", "1234567890123A", "\x00", "\xff\xff",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	p := siret.New()

	f.Fuzz(func(t *testing.T, s string) {
		in := p.Canonicalize(businessid.IdentifierInput{Value: s})

		res, err := p.ValidateFormat(context.Background(), in)
		if err == nil && res == nil {
			t.Fatalf("ValidateFormat nil result with nil error, input=%q", s)
		}

		cres, cerr := p.ValidateChecksum(context.Background(), in)
		if cerr == nil && cres == nil {
			t.Fatalf("ValidateChecksum nil result with nil error, input=%q", s)
		}
	})
}
