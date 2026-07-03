// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid

import "github.com/hyperscale-stack/businessid"

// countryConfig routes an EUID's REGISTRATION segment to the national
// sub-validator that understands the local business identifier format.
type countryConfig struct {
	// subValidatorKind identifies the [businessid.IdentifierKind] used to
	// look up the sub-validator injected via [WithSubValidator]. An empty
	// value means no delegation; the EUID validator only enforces the
	// generic BRIS charset check.
	subValidatorKind businessid.IdentifierKind
}

// euidCountryConfigs binds each EU country to its national identifier
// scheme. Today only FR is wired (REGISTRATION = SIREN). Extension pattern
// per country: add the sub-validator kind here, wire the concrete
// provider via [defaults.New] with [WithSubValidator].
//
// Non-EU codes (XI, GB, NO, IS, LI) are absent on purpose — BRIS does not
// cover them, so the generic charset check is the only meaningful format
// validation.
var euidCountryConfigs = map[string]countryConfig{
	"AT": {},
	"BE": {},
	"BG": {},
	"HR": {},
	"CY": {},
	"CZ": {},
	"DE": {},
	"DK": {},
	"EE": {},
	"EL": {},
	"ES": {},
	"FI": {},
	"FR": {subValidatorKind: businessid.IdentifierKindSIREN},
	"HU": {},
	"IE": {},
	"IT": {},
	"LT": {},
	"LU": {},
	"LV": {},
	"MT": {},
	"NL": {},
	"PL": {},
	"PT": {},
	"RO": {},
	"SE": {},
	"SI": {},
	"SK": {},
}
