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
// [A-Z0-9./\- +].
//
// The provider treats the EUID as a meta-format: the country code selects
// a native national-register validator (see [nationals.go]) which
// validates the REGISTRATION segment against its own rules (format +
// checksum where documented). All 27 EU-27 registers are covered
// natively — no external wiring is required.
//
// For countries outside BRIS (XI, GB, NO, IS, LI) or for a custom
// override, [WithCountryValidator] injects a caller-provided validator.
package euid

import (
	"context"
	"strings"

	"github.com/hyperscale-stack/businessid"
)

const (
	maxRegisterLen     = 20
	maxRegistrationLen = 64
	msgEmpty           = "empty value"
)

// Option configures a Provider at construction time.
type Option func(*Provider)

// WithCountryValidator registers a custom register-validator for a
// 2-letter country code. This overrides the native validator (if any)
// for that country, and is the extension point for non-EU codes.
// Multiple calls compose; the last one wins for a given code.
func WithCountryValidator(cc string, rv registerValidator) Option {
	return func(p *Provider) {
		if p.overrides == nil {
			p.overrides = map[string]registerValidator{}
		}

		p.overrides[strings.ToUpper(cc)] = rv
	}
}

// Provider validates EUID codes.
type Provider struct {
	overrides map[string]registerValidator
}

// New returns a new EUID provider. See [Option] for available options.
func New(opts ...Option) *Provider {
	p := &Provider{}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindEUID }

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: true, Registry: false}
}

// Canonicalize trims and upper-cases. It preserves inner spaces, dots
// and dashes because they can appear inside the registration segment.
func (Provider) Canonicalize(input businessid.IdentifierInput) businessid.IdentifierInput {
	input.Value = businessid.TrimUpper(input.Value)
	input.CountryCode = businessid.NormalizeCountryCode(input.CountryCode)

	return input
}

// parsedEUID captures the three segments of a canonicalized EUID.
type parsedEUID struct {
	countryCode  string
	register     string // segment between CC and the '.'
	registration string // segment after the '.'
}

// parse validates the basic BRIS layout and returns the segments. On
// failure it sets the reason on res and returns ok = false.
func parse(value string, res *businessid.ValidationResult) (parsedEUID, bool) {
	dot := strings.IndexByte(value, '.')
	if dot < 3 {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidFormat
		res.Message = "EUID must contain a dot after the register segment"

		return parsedEUID{}, false
	}

	prefix := value[:dot]
	registration := value[dot+1:]

	if len(prefix) < 3 || len(prefix) > 2+maxRegisterLen {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "EUID register segment length out of range"

		return parsedEUID{}, false
	}

	if !businessid.IsASCIICountryPrefix(prefix) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidFormat
		res.Message = "EUID must begin with a 2-letter country code"

		return parsedEUID{}, false
	}

	register := prefix[2:]

	if !businessid.IsAlnumUpper(register) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "EUID register segment must be alphanumeric"

		return parsedEUID{}, false
	}

	if len(registration) == 0 || len(registration) > maxRegistrationLen {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "EUID registration segment length out of range"

		return parsedEUID{}, false
	}

	if !businessid.IsRegistrationCharset(registration) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "EUID registration segment contains invalid characters"

		return parsedEUID{}, false
	}

	return parsedEUID{
		countryCode:  prefix[:2],
		register:     register,
		registration: registration,
	}, true
}

// lookup resolves the register-validator for a country code, giving
// caller-supplied overrides priority over the native table.
func (p Provider) lookup(cc string) (registerValidator, bool) {
	if rv, ok := p.overrides[cc]; ok {
		return rv, true
	}

	rv, ok := euidRegisterValidators[cc]

	return rv, ok
}

// ValidateFormat implements [businessid.FormatValidator].
//
// After the BRIS layout is checked, the country's native validator
// (or an override registered via [WithCountryValidator]) validates the
// REGISTRATION segment. If no validator exists for the country, only
// the generic BRIS charset check applies.
func (p Provider) ValidateFormat(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindEUID,
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

	parsed, ok := parse(input.Value, res)
	if !ok {
		return res, nil
	}

	res.CountryCode = parsed.countryCode

	rv, hasRV := p.lookup(parsed.countryCode)
	if !hasRV || rv.validateFormat == nil {
		res.Status = businessid.ValidationStatusValid
		res.ReasonCode = businessid.ReasonOK

		return res, nil
	}

	registration := parsed.registration
	if rv.canonicalize != nil {
		registration = rv.canonicalize(registration)
	}

	valid, reason, msg := rv.validateFormat(registration)
	if !valid {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = reason
		res.Message = "EUID registration invalid: " + msg

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// ValidateChecksum implements [businessid.ChecksumValidator]. It runs
// the country's native checksum on the REGISTRATION segment. Countries
// whose national register has no publicly documented checksum report
// [businessid.ValidationStatusUnsupported].
func (p Provider) ValidateChecksum(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindEUID,
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

	parsed, ok := parse(input.Value, res)
	if !ok {
		return res, nil
	}

	res.CountryCode = parsed.countryCode

	rv, hasRV := p.lookup(parsed.countryCode)
	if !hasRV || rv.validateChecksum == nil {
		res.Status = businessid.ValidationStatusUnsupported
		res.ReasonCode = businessid.ReasonUnsupportedChecksum
		res.Message = "EUID checksum not implemented for this country's register"

		return res, nil
	}

	registration := parsed.registration
	if rv.canonicalize != nil {
		registration = rv.canonicalize(registration)
	}

	// A checksum is only meaningful when the format was already OK — we
	// re-run the format check to avoid handing malformed input to the
	// checksum function (which trusts positional invariants).
	if rv.validateFormat != nil {
		if valid, reason, msg := rv.validateFormat(registration); !valid {
			res.Status = businessid.ValidationStatusInvalid
			res.ReasonCode = reason
			res.Message = "EUID registration invalid: " + msg

			return res, nil
		}
	}

	if !rv.validateChecksum(registration) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidChecksum
		res.Message = "EUID registration checksum failed"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}
