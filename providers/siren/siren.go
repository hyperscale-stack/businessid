// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package siren validates the French SIREN business identifier.
//
// A SIREN is exactly 9 digits and validates against the Luhn (mod-10)
// checksum. Callers can register additional non-Luhn dérogations (for
// historical entities that pre-date Luhn adoption) via [WithDerogation].
package siren

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const (
	length   = 9
	msgEmpty = "empty value"
)

// DerogationRule reports whether a canonical (digit-only) SIREN satisfies a
// non-Luhn national rule. It is only consulted when Luhn itself fails.
type DerogationRule func(canonical string) bool

// Option configures a Provider at construction time.
type Option func(*Provider)

// WithDerogation registers a non-Luhn rule for the given SIREN prefix (the
// full 9-digit SIREN). When Luhn fails on a value whose SIREN matches, the
// registered rule is consulted. Multiple calls compose; the last one wins
// for a given prefix.
func WithDerogation(siren string, rule DerogationRule) Option {
	return func(p *Provider) {
		if p.derogations == nil {
			p.derogations = map[string]DerogationRule{}
		}

		p.derogations[siren] = rule
	}
}

// Provider validates SIREN numbers.
type Provider struct {
	derogations map[string]DerogationRule
}

// New returns a new SIREN provider. See [Option] for available options.
func New(opts ...Option) *Provider {
	p := &Provider{derogations: knownSIRENDerogations()}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindSIREN }

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{
		Format:   true,
		Checksum: true,
		Registry: false,
	}
}

// Canonicalize strips whitespace, dots and dashes.
func (Provider) Canonicalize(input businessid.IdentifierInput) businessid.IdentifierInput {
	input.Value = businessid.StripSeparators(businessid.StripAllSpaces(input.Value), ".", "-")
	input.CountryCode = businessid.NormalizeCountryCode(input.CountryCode)

	return input
}

// ValidateFormat implements [businessid.FormatValidator].
func (p Provider) ValidateFormat(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindSIREN,
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
		res.Message = "SIREN must be exactly 9 digits"

		return res, nil
	}

	if !businessid.IsAllDigits(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "SIREN must contain only digits"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// ValidateChecksum implements [businessid.ChecksumValidator].
func (p Provider) ValidateChecksum(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindSIREN,
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
		res.Message = "SIREN checksum failed"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// isValidChecksum accepts a SIREN whose 9-digit body satisfies Luhn, or a
// registered dérogation rule for the matching SIREN prefix. Callers must
// have already verified length and digit-only shape.
func (p Provider) isValidChecksum(s string) bool {
	if businessid.Luhn(s) {
		return true
	}

	if rule, ok := p.derogations[s]; ok {
		return rule(s)
	}

	return false
}
