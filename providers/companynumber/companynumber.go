// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package companynumber validates a UK Companies House company number.
//
// A valid company number is either eight digits, or a two-letter prefix
// followed by six digits. The recognized prefixes are:
//
//	SC (Scotland)   NI (Northern Ireland)
//	OC (LLP E&W)    SO (LLP Scotland)
//	NC (LLP NI)     IP (Industrial & Provident society)
//	SP, IC, SI, R0 (specialist registers)
package companynumber

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const digitLength = 8

// Recognized two-letter prefixes followed by six digits.
var validPrefixes = map[string]struct{}{
	"SC": {}, "NI": {}, "OC": {}, "SO": {}, "NC": {},
	"IP": {}, "SP": {}, "IC": {}, "SI": {}, "R0": {},
}

// Provider validates UK company numbers.
type Provider struct{}

// New returns a new UK company number provider.
func New() *Provider { return &Provider{} }

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindCompanyNumber }

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
		Kind:           businessid.IdentifierKindCompanyNumber,
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

	if len(input.Value) != digitLength {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "UK company number must be 8 characters"

		return res, nil
	}

	if businessid.IsAllDigits(input.Value) {
		res.Status = businessid.ValidationStatusValid
		res.ReasonCode = businessid.ReasonOK

		return res, nil
	}

	prefix := input.Value[:2]
	suffix := input.Value[2:]

	if _, ok := validPrefixes[prefix]; !ok {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidFormat
		res.Message = "UK company number prefix not recognized"

		return res, nil
	}

	if !businessid.IsAllDigits(suffix) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "UK company number suffix must be 6 digits"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}
