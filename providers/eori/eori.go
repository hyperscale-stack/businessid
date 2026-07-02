// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package eori validates the Economic Operators Registration and
// Identification number.
//
// An EORI is a 2-letter country code followed by 1..15 upper-case
// alphanumeric characters.
package eori

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const (
	minLen = 3
	maxLen = 17
)

// Provider validates EORI numbers.
type Provider struct{}

// New returns a new EORI provider.
func New() *Provider { return &Provider{} }

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindEORI }

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: false, Registry: false}
}

// Canonicalize trims, upper-cases, and strips whitespace.
func (Provider) Canonicalize(input businessid.IdentifierInput) businessid.IdentifierInput {
	input.Value = businessid.StripAllSpaces(businessid.TrimUpper(input.Value))
	input.CountryCode = businessid.NormalizeCountryCode(input.CountryCode)

	return input
}

// ValidateFormat implements [businessid.FormatValidator].
func (Provider) ValidateFormat(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindEORI,
		Level:          businessid.ValidationLevelFormat,
		CanonicalValue: input.Value,
		CountryCode:    input.CountryCode,
	}

	if input.Value == "" {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonEmpty
		res.Message = "empty value"

		return res, nil
	}

	if len(input.Value) < minLen || len(input.Value) > maxLen {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "EORI must be 3..17 characters"

		return res, nil
	}

	if !businessid.IsASCIICountryPrefix(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidFormat
		res.Message = "EORI must begin with a 2-letter country code"

		return res, nil
	}

	if !businessid.IsAlnumUpper(input.Value[2:]) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "EORI suffix must be alphanumeric"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}
