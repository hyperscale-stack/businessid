// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siret

const laPosteSIREN = "356000000"

// knownSIRETDerogations returns a fresh copy of the built-in table of
// non-Luhn SIRET rules.
//
// La Poste (SIREN 356000000): the INSEE dérogatoire rule accepts a SIRET
// whose 14 digits sum to a multiple of 5. Historical background: La Poste
// operates more than 17 000 establishments and would have exhausted
// Luhn-valid NICs; INSEE granted this exception so that new establishments
// can continue to be numbered.
//
// The table is returned as a copy so mutation by callers cannot leak
// across providers.
func knownSIRETDerogations() map[string]DerogationRule {
	return map[string]DerogationRule{
		laPosteSIREN: laPosteRule,
	}
}

// laPosteRule reports whether s satisfies La Poste's divisible-by-5 sum
// rule. Callers must have already verified digit-only shape and length.
func laPosteRule(s string) bool { return digitSum(s)%5 == 0 }
