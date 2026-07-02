// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package siret validates the French SIRET establishment identifier.
//
// A SIRET is exactly 14 digits (SIREN + 5-digit NIC) and normally validates
// against the Luhn (mod-10) checksum. The La Poste exception (SIREN 356000000)
// is not handled in V1 and returns invalid checksum.
package siret

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const length = 14

// Provider validates SIRET numbers.
type Provider struct{}

// New returns a new SIRET provider.
func New() *Provider { return &Provider{} }

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindSIRET }

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: true, Registry: false}
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
		Kind:           businessid.IdentifierKindSIRET,
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
		res.Message = "SIRET must be exactly 14 digits"

		return res, nil
	}

	if !businessid.IsAllDigits(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "SIRET must contain only digits"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// ValidateChecksum implements [businessid.ChecksumValidator].
func (Provider) ValidateChecksum(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindSIRET,
		Level:          businessid.ValidationLevelChecksum,
		InputValue:     input.Value,
		CanonicalValue: input.Value,
		CountryCode:    input.CountryCode,
	}

	if !businessid.Luhn(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidChecksum
		res.Message = "SIRET Luhn check failed"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}
