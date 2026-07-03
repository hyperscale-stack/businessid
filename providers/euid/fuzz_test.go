// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/euid"
	"github.com/hyperscale-stack/businessid/providers/siren"
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

	// Two providers — one bare, one with SIREN sub-validator wired.
	pBare := euid.New()
	pWithSub := euid.New(euid.WithSubValidator(siren.New()))

	f.Fuzz(func(t *testing.T, s string) {
		for _, prov := range []*euid.Provider{pBare, pWithSub} {
			in := prov.Canonicalize(businessid.IdentifierInput{Value: s})

			res, err := prov.ValidateFormat(context.Background(), in)
			if err == nil && res == nil {
				t.Fatalf("ValidateFormat nil result with nil error, input=%q", s)
			}

			cres, cerr := prov.ValidateChecksum(context.Background(), in)
			if cerr == nil && cres == nil {
				t.Fatalf("ValidateChecksum nil result with nil error, input=%q", s)
			}
		}
	})
}
