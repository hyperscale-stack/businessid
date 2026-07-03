// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid

import "testing"

// TestNationalsDefensiveBranches exercises helper branches that are
// unreachable via the public EUID validator (they're gated by the BRIS
// charset / meta-format checks) so they remain safe under refactors.
func TestNationalsDefensiveBranches(t *testing.T) {
	t.Parallel()

	// esDNI: rejects a body whose first 8 bytes are not all digits.
	if esDNI("1234567AZ") {
		t.Errorf("esDNI should reject non-digit body")
	}

	// esNIE: rejects a body whose middle 7 bytes are not all digits.
	if esNIE("X123456AZ") {
		t.Errorf("esNIE should reject non-digit middle")
	}
	// esNIE: Y prefix and Z prefix branches.
	if !esNIE("Y0000000Z") {
		t.Errorf("esNIE Y-prefix should validate")
	}
	if !esNIE("Z0000000M") {
		t.Errorf("esNIE Z-prefix should validate")
	}

	// esCIF: rejects a body whose middle 7 bytes are not all digits.
	if esCIF("A123456AZ") {
		t.Errorf("esCIF should reject non-digit middle")
	}

	// esCIFOrDNIChecksum: rejects an unclassified first character.
	if esCIFOrDNIChecksum("!12345678") {
		t.Errorf("esCIFOrDNIChecksum should reject unknown first char")
	}

	// bgBulstat9Checksum alt-alt path: both sums yield 10 → check=0.
	// "605000000" was constructed to hit this exact branch.
	if !bgBulstat9Checksum("605000000") {
		t.Errorf("bgBulstat9Checksum should validate 605000000 via alt-alt path")
	}
	// Invalid vector where check != d[8].
	if bgBulstat9Checksum("000000550") {
		t.Errorf("bgBulstat9Checksum should reject wrong check digit")
	}

	// icoMod11 default case where r==10 → check=1.
	if !icoMod11("00000051") {
		t.Errorf("icoMod11 should validate 00000051 (r==10 branch)")
	}

	// dkRegisterChecksum defensive: first digit == 0 returns false.
	if dkRegisterChecksum("01056416") {
		t.Errorf("dkRegisterChecksum should reject first digit 0")
	}

	// itoa(0) branch.
	if got := itoa(0); got != "0" {
		t.Errorf("itoa(0) = %q, want 0", got)
	}

	// fiRegisterChecksum r == 1 branch.
	if fiRegisterChecksum("80000000") {
		t.Errorf("fiRegisterChecksum r==1 should return false")
	}

	// lvRegisterChecksum legal check==10 branch.
	if lvRegisterChecksum("90000000000") {
		t.Errorf("lvRegisterChecksum legal check==10 should return false")
	}

	// roRegisterChecksum defensive offset check (unreachable in practice).
	if roRegisterChecksum("") {
		t.Errorf("roRegisterChecksum empty should return false")
	}

	// esRegisterFormat empty non-alnum first character branch.
	if ok, _, _ := esRegisterFormat("/12345678"); ok {
		t.Errorf("esRegisterFormat should reject non-alnum first char")
	}
}
