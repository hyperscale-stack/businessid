// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid_test

import (
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/stretchr/testify/assert"
)

func TestMod9710(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty", input: "", want: false},
		{name: "known-lei-deutsche-bank", input: "529900T8BM49AURSDO55", want: true},
		{name: "known-lei-nordea", input: "6SCPQ280AIY8EP3XFW53", want: true},
		{name: "off-by-one", input: "529900T8BM49AURSDO56", want: false},
		{name: "lowercase-invalid", input: "529900t8bm49aursdo55", want: false},
		{name: "non-alnum", input: "529900T8BM49AURSDO5-", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, businessid.Mod9710(tc.input))
		})
	}
}
