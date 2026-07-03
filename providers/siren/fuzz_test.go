// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siren_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/siren"
)

func FuzzValidateFormat(f *testing.F) {
	seeds := []string{
		"", "552100554", "552 100 554", "552-100-554", "000000000",
		"12345678", "1234567890", "12345678A", "\x00", "\xff\xff",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	p := siren.New()

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
