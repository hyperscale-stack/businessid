// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package defaults wires every provider shipped by this module into a
// single ready-to-use [businessid.Validator].
//
// This package exists as a separate module path to avoid an import cycle
// between the root package (which defines the Provider interface) and the
// per-kind provider sub-packages (which implement it).
package defaults

import (
	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/companynumber"
	"github.com/hyperscale-stack/businessid/providers/duns"
	"github.com/hyperscale-stack/businessid/providers/ein"
	"github.com/hyperscale-stack/businessid/providers/eori"
	"github.com/hyperscale-stack/businessid/providers/euid"
	"github.com/hyperscale-stack/businessid/providers/lei"
	"github.com/hyperscale-stack/businessid/providers/nationalregistration"
	"github.com/hyperscale-stack/businessid/providers/siren"
	"github.com/hyperscale-stack/businessid/providers/siret"
	"github.com/hyperscale-stack/businessid/providers/vat"
)

// Providers returns freshly-constructed instances of every default provider.
//
// EUID has native validators for every EU-27 register — no wiring is
// needed at the caller level.
func Providers() []businessid.Provider {
	return []businessid.Provider{
		siren.New(),
		siret.New(),
		lei.New(),
		duns.New(),
		ein.New(),
		companynumber.New(),
		euid.New(),
		eori.New(),
		vat.New(),
		nationalregistration.New(),
	}
}

// WithAll returns an option that registers every default provider.
func WithAll() businessid.Option {
	return businessid.WithProviders(Providers()...)
}

// New returns a Validator with every default provider registered.
func New() *businessid.Validator {
	return businessid.NewValidator(WithAll())
}
