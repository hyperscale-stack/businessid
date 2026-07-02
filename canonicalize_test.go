// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid_test

import (
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/stretchr/testify/assert"
)

func TestStripSeparators(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		seps  []string
		want  string
	}{
		{name: "no-seps", input: "abc", seps: nil, want: "abc"},
		{name: "dashes", input: "12-34-56", seps: []string{"-"}, want: "123456"},
		{name: "mixed", input: "12.34 56-78", seps: []string{".", " ", "-"}, want: "12345678"},
		{name: "preserve-leading-zero", input: "0-01234", seps: []string{"-"}, want: "001234"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, businessid.StripSeparators(tc.input, tc.seps...))
		})
	}
}

func TestStripAllSpaces(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "ascii-space", input: "a b c", want: "abc"},
		{name: "tab", input: "a\tb", want: "ab"},
		{name: "nbsp", input: "a b", want: "ab"},
		{name: "ideographic-space", input: "a　b", want: "ab"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, businessid.StripAllSpaces(tc.input))
		})
	}
}

func TestTrimUpper(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "lower", input: "fr123", want: "FR123"},
		{name: "trim", input: "  hello  ", want: "HELLO"},
		{name: "mixed", input: " Fr-12 ", want: "FR-12"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, businessid.TrimUpper(tc.input))
		})
	}
}

func TestNormalizeCountryCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "lower", input: "fr", want: "FR"},
		{name: "padded", input: "  FR  ", want: "FR"},
		{name: "too-long", input: "FRA", want: ""},
		{name: "digits", input: "F1", want: ""},
		{name: "one-letter", input: "F", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, businessid.NormalizeCountryCode(tc.input))
		})
	}
}

func TestIsAllDigits(t *testing.T) {
	t.Parallel()

	assert.False(t, businessid.IsAllDigits(""))
	assert.True(t, businessid.IsAllDigits("0"))
	assert.True(t, businessid.IsAllDigits("00012345"))
	assert.False(t, businessid.IsAllDigits("12A"))
	assert.False(t, businessid.IsAllDigits("1 2"))
}

func TestIsAlnumUpper(t *testing.T) {
	t.Parallel()

	assert.False(t, businessid.IsAlnumUpper(""))
	assert.True(t, businessid.IsAlnumUpper("ABC123"))
	assert.False(t, businessid.IsAlnumUpper("abc"))
	assert.False(t, businessid.IsAlnumUpper("AB-12"))
}
