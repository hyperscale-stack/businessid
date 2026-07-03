// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package siret validates the French SIRET establishment identifier.
//
// A SIRET is exactly 14 digits (SIREN + 5-digit NIC) and normally validates
// against the Luhn (mod-10) checksum. The NIC 00000 is reserved by INSEE
// and never attributed, so it is rejected at the format level.
//
// Establishments of La Poste (SIREN 356000000) may instead satisfy INSEE's
// dérogatoire rule where the plain sum of the 14 digits is divisible by 5;
// this validator accepts a La Poste SIRET when either check passes.
// Additional non-Luhn rules for other historical SIRENs can be registered
// via [WithDerogation].
package siret

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const (
	length      = 14
	sirenLength = 9
	reservedNIC = "00000"
	msgEmpty    = "empty value"
)

// DerogationRule reports whether a canonical (digit-only) SIRET satisfies a
// non-Luhn national rule. It is only consulted when Luhn itself fails.
type DerogationRule func(canonical string) bool

// Option configures a Provider at construction time.
type Option func(*Provider)

// WithDerogation registers a non-Luhn rule for SIRETs whose 9-digit SIREN
// prefix matches the given value. When Luhn fails, the registered rule is
// consulted. Multiple calls compose; the last one wins for a given prefix.
func WithDerogation(siren string, rule DerogationRule) Option {
	return func(p *Provider) {
		if p.derogations == nil {
			p.derogations = map[string]DerogationRule{}
		}

		p.derogations[siren] = rule
	}
}

// Provider validates SIRET numbers.
type Provider struct {
	derogations map[string]DerogationRule
}

// New returns a new SIRET provider. See [Option] for available options.
func New(opts ...Option) *Provider {
	p := &Provider{derogations: knownSIRETDerogations()}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

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
		CanonicalValue: input.Value,
		CountryCode:    input.CountryCode,
	}

	if input.Value == "" {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonEmpty
		res.Message = msgEmpty

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

	if input.Value[sirenLength:] == reservedNIC {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidFormat
		res.Message = "SIRET NIC 00000 is reserved by INSEE"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// ValidateChecksum implements [businessid.ChecksumValidator].
func (p Provider) ValidateChecksum(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindSIRET,
		Level:          businessid.ValidationLevelChecksum,
		CanonicalValue: input.Value,
		CountryCode:    input.CountryCode,
	}

	if input.Value == "" {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonEmpty
		res.Message = msgEmpty

		return res, nil
	}

	if !p.isValidChecksum(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidChecksum
		res.Message = "SIRET checksum failed"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// isValidChecksum accepts a SIRET whose 14-digit body satisfies Luhn, or a
// registered dérogation rule for the matching SIREN prefix. Callers must
// have already verified length and digit-only shape.
func (p Provider) isValidChecksum(s string) bool {
	if businessid.Luhn(s) {
		return true
	}

	if len(s) < sirenLength {
		return false
	}

	if rule, ok := p.derogations[s[:sirenLength]]; ok {
		return rule(s)
	}

	return false
}

// digitSum returns the arithmetic sum of the ASCII digits in s. The caller
// must ensure every byte is '0'..'9'.
func digitSum(s string) int {
	n := 0

	for i := range len(s) {
		n += int(s[i] - '0')
	}

	return n
}
