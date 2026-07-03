// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siren

// knownSIRENDerogations returns a fresh copy of the built-in table of
// non-Luhn SIREN rules.
//
// La Poste's SIREN (356000000) is Luhn-valid on its own, so no dérogation
// is needed at SIREN level. Callers who encounter historical SIRENs that
// pre-date Luhn adoption can inject rules via [WithDerogation].
//
// The table is returned as a copy so mutation by callers cannot leak
// across providers.
func knownSIRENDerogations() map[string]DerogationRule {
	return map[string]DerogationRule{}
}
