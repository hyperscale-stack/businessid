// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/euid"
)

func FuzzValidateFormat(f *testing.F) {
	seeds := []string{
		"", "FRRCS.552100554", "DEHRB.HAMBURG/B-12345",
		"FR3102.552100554", "FRRCS.ABC", "IECRO.12+34",
		"XX.", ".", "..", "\x00.\x00", "\xff\xff.\xff",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	p := euid.New()

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
