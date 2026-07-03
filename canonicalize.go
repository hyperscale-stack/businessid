// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

import (
	"strings"
	"unicode"
)

// StripSeparators removes every occurrence of any rune in seps from s.
//
// It never removes digits or letters, so leading zeros are preserved.
func StripSeparators(s string, seps ...string) string {
	if len(seps) == 0 {
		return s
	}

	set := make(map[rune]struct{}, len(seps))

	for _, sep := range seps {
		for _, r := range sep {
			set[r] = struct{}{}
		}
	}

	var b strings.Builder

	b.Grow(len(s))

	for _, r := range s {
		if _, drop := set[r]; drop {
			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}

// StripAllSpaces removes any Unicode whitespace, including non-breaking spaces.
func StripAllSpaces(s string) string {
	var b strings.Builder

	b.Grow(len(s))

	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}

// TrimUpper trims ASCII whitespace and upper-cases the ASCII letters in s.
func TrimUpper(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// NormalizeCountryCode returns the trimmed upper-case country code.
//
// If the resulting value is not exactly two ASCII letters, it returns "".
func NormalizeCountryCode(cc string) string {
	cc = TrimUpper(cc)

	if len(cc) != 2 {
		return ""
	}

	for i := range 2 {
		if cc[i] < 'A' || cc[i] > 'Z' {
			return ""
		}
	}

	return cc
}

// IsAllDigits reports whether s is non-empty and contains only ASCII digits.
func IsAllDigits(s string) bool {
	if s == "" {
		return false
	}

	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

// IsAlnumUpper reports whether s is non-empty and contains only
// upper-case ASCII letters or digits.
func IsAlnumUpper(s string) bool {
	if s == "" {
		return false
	}

	for i := range len(s) {
		c := s[i]
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') {
			continue
		}

		return false
	}

	return true
}

// IsRegistrationCharset reports whether s is non-empty and contains only
// characters from the set A-Z, 0-9, dot, slash, dash, space and plus — the
// character class shared by EUID registration segments and generic
// national registration numbers. The plus sign is accepted because some
// national registers (e.g. legacy Irish, some Nordic registrations) emit
// it as part of the local identifier.
func IsRegistrationCharset(s string) bool {
	if s == "" {
		return false
	}

	for i := range len(s) {
		c := s[i]

		switch {
		case c >= '0' && c <= '9',
			c >= 'A' && c <= 'Z',
			c == '.', c == '/', c == '-', c == ' ', c == '+':
			continue
		default:
			return false
		}
	}

	return true
}

// IsASCIICountryPrefix reports whether s begins with two upper-case ASCII
// letters. Strings shorter than 2 bytes return false.
func IsASCIICountryPrefix(s string) bool {
	if len(s) < 2 {
		return false
	}

	return s[0] >= 'A' && s[0] <= 'Z' && s[1] >= 'A' && s[1] <= 'Z'
}
