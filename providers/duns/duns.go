// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package duns validates the Dun & Bradstreet D-U-N-S number.
//
// A DUNS is exactly 9 digits. No publicly documented checksum exists.
package duns

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const length = 9

// Provider validates DUNS numbers.
type Provider struct{}

// New returns a new DUNS provider.
func New() *Provider { return &Provider{} }

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindDUNS }

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: false, Registry: false}
}

// Canonicalize strips whitespace, dots and dashes.
func (Provider) Canonicalize(input businessid.IdentifierInput) businessid.IdentifierInput {
	input.Value = businessid.StripSeparators(businessid.StripAllSpaces(input.Value), ".", "-")
	input.CountryCode = businessid.NormalizeCountryCode(input.CountryCode)

	return input
}

// ValidateFormat implements [businessid.FormatValidator].
func (Provider) ValidateFormat(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindDUNS,
		Level:          businessid.ValidationLevelFormat,
		InputValue:     input.Value,
		CanonicalValue: input.Value,
		CountryCode:    input.CountryCode,
	}

	if input.Value == "" {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonEmpty
		res.Message = "empty value"

		return res, nil
	}

	if len(input.Value) != length {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "DUNS must be exactly 9 digits"

		return res, nil
	}

	if !businessid.IsAllDigits(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "DUNS must contain only digits"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}
