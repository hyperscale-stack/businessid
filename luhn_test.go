// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid_test

import (
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/stretchr/testify/assert"
)

func TestLuhn(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty", input: "", want: false},
		{name: "single-zero", input: "0", want: true},
		{name: "all-zeros", input: "000000000", want: true},
		{name: "known-siren-danone", input: "732829320", want: true},
		{name: "known-siren-lvmh", input: "552100554", want: true},
		{name: "known-siret-danone", input: "73282932000074", want: true},
		{name: "off-by-one", input: "732829321", want: false},
		{name: "letter", input: "73282932A", want: false},
		{name: "space", input: "7328 29320", want: false},
		{name: "credit-card-valid", input: "4532015112830366", want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, businessid.Luhn(tc.input))
		})
	}
}
