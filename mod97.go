// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

// Mod9710 reports whether s satisfies the ISO/IEC 7064 MOD 97-10 checksum
// (as used by LEI and IBAN).
//
// Letters are folded to numbers with A=10..Z=35. The check passes when the
// resulting integer, computed digit by digit, is congruent to 1 modulo 97.
// Empty strings and strings containing anything other than digits or
// upper-case ASCII letters return false.
func Mod9710(s string) bool {
	if s == "" {
		return false
	}

	remainder := 0

	for i := range len(s) {
		c := s[i]

		switch {
		case c >= '0' && c <= '9':
			remainder = (remainder*10 + int(c-'0')) % 97
		case c >= 'A' && c <= 'Z':
			v := int(c-'A') + 10
			remainder = (remainder*100 + v) % 97
		default:
			return false
		}
	}

	return remainder == 1
}
