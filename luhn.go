// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

// Luhn reports whether s satisfies the Luhn (mod-10) checksum used by SIREN,
// SIRET, and many national identifiers.
//
// Empty strings and strings containing non-digit characters return false.
func Luhn(s string) bool {
	if s == "" {
		return false
	}

	var sum int

	double := false

	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]

		if c < '0' || c > '9' {
			return false
		}

		d := int(c - '0')

		if double {
			d *= 2

			if d > 9 {
				d -= 9
			}
		}

		sum += d
		double = !double
	}

	return sum%10 == 0
}
