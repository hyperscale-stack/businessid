// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package nationalregistration validates opaque national registration
// identifiers.
//
// The generic V1 check only enforces a character-class and length range;
// country-specific validation can be layered on top later.
package nationalregistration

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const maxLen = 64

// Provider validates national registration numbers.
type Provider struct{}

// New returns a new national registration number provider.
func New() *Provider { return &Provider{} }

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind {
	return businessid.IdentifierKindNationalRegistrationNumber
}

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: false, Registry: false}
}

// Canonicalize trims and upper-cases.
func (Provider) Canonicalize(input businessid.IdentifierInput) businessid.IdentifierInput {
	input.Value = businessid.TrimUpper(input.Value)
	input.CountryCode = businessid.NormalizeCountryCode(input.CountryCode)

	return input
}

// ValidateFormat implements [businessid.FormatValidator].
func (Provider) ValidateFormat(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindNationalRegistrationNumber,
		Level:          businessid.ValidationLevelFormat,
		InputValue:     input.Value,
		CanonicalValue: input.Value,
		CountryCode:    input.CountryCode,
	}

	if input.CountryCode == "" {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonMissingCountryCode
		res.Message = "country code is required"

		return res, nil
	}

	if input.Value == "" {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonEmpty
		res.Message = "empty value"

		return res, nil
	}

	if len(input.Value) > maxLen {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "national registration number too long"

		return res, nil
	}

	if !isRegistrationCharset(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "national registration number contains invalid characters"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

func isRegistrationCharset(s string) bool {
	for i := range len(s) {
		c := s[i]

		switch {
		case c >= '0' && c <= '9',
			c >= 'A' && c <= 'Z',
			c == '.', c == '/', c == '-', c == ' ':
			continue
		default:
			return false
		}
	}

	return true
}
