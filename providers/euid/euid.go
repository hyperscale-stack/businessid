// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package euid validates the European Unique Identifier for cross-border
// business registers.
//
// An EUID has the form:
//
//	<CC><REGISTER>.<REGISTRATION>
//
// where CC is a 2-letter country code, REGISTER is 1..20 upper-case
// alphanumeric characters and REGISTRATION is 1..64 characters from
// [A-Z0-9./\- ].
package euid

import (
	"context"
	"strings"

	"github.com/hyperscale-stack/businessid"
)

const (
	maxRegisterLen     = 20
	maxRegistrationLen = 64
)

// Provider validates EUID codes.
type Provider struct{}

// New returns a new EUID provider.
func New() *Provider { return &Provider{} }

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindEUID }

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: false, Registry: false}
}

// Canonicalize trims and upper-cases. It preserves inner spaces, dots and
// dashes because they can appear inside the registration segment.
func (Provider) Canonicalize(input businessid.IdentifierInput) businessid.IdentifierInput {
	input.Value = businessid.TrimUpper(input.Value)
	input.CountryCode = businessid.NormalizeCountryCode(input.CountryCode)

	return input
}

// ValidateFormat implements [businessid.FormatValidator].
func (Provider) ValidateFormat(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindEUID,
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

	dot := strings.IndexByte(input.Value, '.')
	if dot < 3 {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidFormat
		res.Message = "EUID must contain a dot after the register segment"

		return res, nil
	}

	prefix := input.Value[:dot]
	registration := input.Value[dot+1:]

	if len(prefix) < 3 || len(prefix) > 2+maxRegisterLen {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "EUID register segment length out of range"

		return res, nil
	}

	if !isTwoLetterPrefix(prefix[:2]) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidFormat
		res.Message = "EUID must begin with a 2-letter country code"

		return res, nil
	}

	if !businessid.IsAlnumUpper(prefix[2:]) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "EUID register segment must be alphanumeric"

		return res, nil
	}

	if len(registration) == 0 || len(registration) > maxRegistrationLen {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "EUID registration segment length out of range"

		return res, nil
	}

	if !isRegistrationCharset(registration) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "EUID registration segment contains invalid characters"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// isTwoLetterPrefix reports whether the 2-byte slice s is two upper-case
// ASCII letters. The caller must have already ensured len(s) == 2.
func isTwoLetterPrefix(s string) bool {
	return s[0] >= 'A' && s[0] <= 'Z' && s[1] >= 'A' && s[1] <= 'Z'
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
