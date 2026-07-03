// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package vat_test

import (
	"context"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/vat"
)

// FuzzValidateFormat verifies that ValidateFormat never panics and always
// returns either a non-nil result or an error, for any input. Seed values
// exercise every country prefix branch.
func FuzzValidateFormat(f *testing.F) {
	seeds := []string{
		"", "FR", "FR44732829320", "DE143454214", "IT07973780013",
		"GB562235987", "XI562235987", "IE6388047V", "ES A39000013",
		"NL010000446B01", "gr094014201", "IE1+23456T", "IE1*23456T",
		"ZZ99999", "\x00\x00", "\xff\xff\xff",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	p := vat.New()
	pLegacy := vat.New(vat.WithLegacy())

	f.Fuzz(func(t *testing.T, s string) {
		for _, prov := range []*vat.Provider{p, pLegacy} {
			in := prov.Canonicalize(businessid.IdentifierInput{Value: s})

			res, err := prov.ValidateFormat(context.Background(), in)
			if err == nil && res == nil {
				t.Fatalf("ValidateFormat: nil result with nil error, input=%q", s)
			}

			cres, cerr := prov.ValidateChecksum(context.Background(), in)
			if cerr == nil && cres == nil {
				t.Fatalf("ValidateChecksum: nil result with nil error, input=%q", s)
			}
		}
	})
}
