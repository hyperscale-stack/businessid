// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package vat

import "testing"

// These tests exercise the defensive branches of the checksum helpers
// that are unreachable through the public VAT API (ValidateFormat gates
// non-digit / non-standard input out before ValidateChecksum runs). They
// exist so those helpers remain safe if a future refactor exposes them.

func TestDefensiveChecksumBranches(t *testing.T) {
	t.Parallel()

	// dniCheck rejects a body whose first 8 bytes are not all digits.
	if dniCheck("1234567AZ") {
		t.Errorf("dniCheck should reject non-digit body")
	}

	// nieCheck rejects a body whose middle 7 bytes are not all digits.
	if nieCheck("X123456AZ") {
		t.Errorf("nieCheck should reject non-digit middle")
	}

	// cifCheck rejects a body whose middle 7 bytes are not all digits.
	if cifCheck("A123456AZ") {
		t.Errorf("cifCheck should reject non-digit middle")
	}

	// checksumESBody rejects an unclassified first character.
	if checksumESBody("!12345678") {
		t.Errorf("checksumESBody should reject unknown first char")
	}

	// Length-out-of-range paths return false (unreachable via public API).
	if checksumBGBody("1234567") {
		t.Errorf("checksumBGBody should reject 7-digit body")
	}
	if checksumCZBody("1234567") {
		t.Errorf("checksumCZBody should reject 7-digit body")
	}
	if checksumGBBody("1234567") {
		t.Errorf("checksumGBBody should reject 7-digit body")
	}
	if checksumLTBody("1234567") {
		t.Errorf("checksumLTBody should reject 7-digit body")
	}
	if checksumIEBody("1234567") {
		t.Errorf("checksumIEBody should reject 7-digit body")
	}
	if checksumISBody("1234") {
		t.Errorf("checksumISBody should reject non-10-digit body")
	}
	if checksumROBody("") {
		t.Errorf("checksumROBody should reject empty body")
	}
	if checksumROBody("12345678901") {
		t.Errorf("checksumROBody should reject 11-digit body")
	}

	// bulstat9 fallthrough where alt-weight sum also yields 10 → check=0.
	// Construct a number where sum1 mod 11 = 10 AND sum2 mod 11 = 10.
	// Digits chosen empirically to hit both branches.
	// This ensures the "if check == 10 { check = 0 }" alt-alt path
	// executes at least once.
	if bulstat9("000000550") {
		t.Errorf("bulstat9 vector should not match check digit here")
	}
}

// TestIEChecksumInvalidLengths ensures the 8-char legacy check path
// (letters in the digit slot) and the invalid-length default return
// false rather than panicking.
func TestIEChecksumInvalidLengths(t *testing.T) {
	t.Parallel()

	// 8-char with non-digit in positions 0..6 (legacy layout) — falls
	// through the "return false // legacy layout" branch.
	if checksumIEBody("1A345678") {
		t.Errorf("checksumIEBody 8-char legacy layout should return false")
	}
	// 9-char with digits at position 7 (not a letter) — invalid.
	if checksumIEBody("123456789") {
		t.Errorf("checksumIEBody 9-char with digit at pos 7 should return false")
	}
	// 9-char with digit at position 8 — invalid.
	if checksumIEBody("1234567A9") {
		t.Errorf("checksumIEBody 9-char with digit at pos 8 should return false")
	}
	// Length outside 8-9.
	if checksumIEBody("123") {
		t.Errorf("checksumIEBody short body should return false")
	}
}
