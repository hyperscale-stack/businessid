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
// a sub-validator (SIREN for FR, etc.) which validates the REGISTRATION
// segment against its own national rules. Sub-validators are injected
// via [WithSubValidator], typically wired by the defaults package.
package euid

import (
	"context"
	"fmt"
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

// WithSubValidator registers a national sub-validator. The provider's
// [businessid.IdentifierKind] is used as the lookup key; the EUID validator
// consults the country → kind table in [euidCountryConfigs] to route a
// given EUID to the right sub-validator. Multiple calls compose; the last
// one wins for a given kind.
func WithSubValidator(sub businessid.Provider) Option {
	return func(p *Provider) {
		if p.subValidators == nil {
			p.subValidators = map[businessid.IdentifierKind]businessid.Provider{}
		}

		p.subValidators[sub.Kind()] = sub
	}
}

// Provider validates EUID codes.
type Provider struct {
	subValidators map[businessid.IdentifierKind]businessid.Provider
}

// New returns a new EUID provider. See [Option] for available options.
func New(opts ...Option) *Provider {
	p := &Provider{
		subValidators: map[businessid.IdentifierKind]businessid.Provider{},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindEUID }

// Capabilities implements [businessid.Provider]. Checksum reflects the
// ability to delegate to a national sub-validator; it is true even when
// no sub-validator is registered because the interface method exists.
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: true, Registry: false}
}

// Canonicalize trims and upper-cases. It preserves inner spaces, dots and
// dashes because they can appear inside the registration segment.
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

// parse validates the basic BRIS layout and returns the segments. It
// returns a non-nil error result if the value is malformed at the meta-
// format level (no dot, bad country code, bad register/registration
// length or charset). The reason code is set on res.
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

// ValidateFormat implements [businessid.FormatValidator].
//
// After the meta-format layout is checked, if a sub-validator is
// registered for the country's configured kind, the REGISTRATION segment
// is delegated to it. This produces stronger format validation than the
// generic charset check (e.g. FR requires the segment to be a valid
// SIREN — 9 digits — instead of any alnum string).
func (p Provider) ValidateFormat(ctx context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
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

	cfg, hasCfg := euidCountryConfigs[parsed.countryCode]
	if hasCfg && cfg.subValidatorKind != "" {
		if sub, ok := p.subValidators[cfg.subValidatorKind]; ok {
			return p.delegateFormat(ctx, res, sub, parsed)
		}
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// delegateFormat runs the sub-provider's canonicalization and format check
// on the REGISTRATION segment and merges the outcome into res. Format
// failures propagate their reason code so callers can distinguish
// "invalid SIREN length" from "invalid EUID layout".
func (p Provider) delegateFormat(ctx context.Context, res *businessid.ValidationResult, sub businessid.Provider, parsed parsedEUID) (*businessid.ValidationResult, error) {
	fv, ok := sub.(businessid.FormatValidator)
	if !ok {
		res.Status = businessid.ValidationStatusValid
		res.ReasonCode = businessid.ReasonOK

		return res, nil
	}

	subInput := sub.Canonicalize(businessid.IdentifierInput{
		Kind:        sub.Kind(),
		Value:       parsed.registration,
		CountryCode: parsed.countryCode,
	})

	subRes, err := fv.ValidateFormat(ctx, subInput)
	if err != nil {
		return nil, fmt.Errorf("euid sub-validator %s format: %w", sub.Kind(), err)
	}

	if subRes.Status != businessid.ValidationStatusValid {
		res.Status = subRes.Status
		res.ReasonCode = subRes.ReasonCode
		res.Message = "EUID registration failed " + string(sub.Kind()) + " format: " + subRes.Message

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// ValidateChecksum implements [businessid.ChecksumValidator]. It delegates
// to the country-configured sub-validator. Countries without a registered
// sub-validator report [businessid.ValidationStatusUnsupported].
func (p Provider) ValidateChecksum(ctx context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
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

	cfg, hasCfg := euidCountryConfigs[parsed.countryCode]
	if !hasCfg || cfg.subValidatorKind == "" {
		res.Status = businessid.ValidationStatusUnsupported
		res.ReasonCode = businessid.ReasonUnsupportedChecksum
		res.Message = "EUID checksum not implemented for this country"

		return res, nil
	}

	sub, ok := p.subValidators[cfg.subValidatorKind]
	if !ok {
		res.Status = businessid.ValidationStatusUnsupported
		res.ReasonCode = businessid.ReasonUnsupportedChecksum
		res.Message = "EUID sub-validator not registered for kind " + string(cfg.subValidatorKind)

		return res, nil
	}

	cv, ok := sub.(businessid.ChecksumValidator)
	if !ok {
		res.Status = businessid.ValidationStatusUnsupported
		res.ReasonCode = businessid.ReasonUnsupportedChecksum
		res.Message = "EUID sub-validator does not support checksum"

		return res, nil
	}

	subInput := sub.Canonicalize(businessid.IdentifierInput{
		Kind:        sub.Kind(),
		Value:       parsed.registration,
		CountryCode: parsed.countryCode,
	})

	subRes, err := cv.ValidateChecksum(ctx, subInput)
	if err != nil {
		return nil, fmt.Errorf("euid sub-validator %s checksum: %w", sub.Kind(), err)
	}

	res.Status = subRes.Status
	res.ReasonCode = subRes.ReasonCode

	if subRes.Message != "" {
		res.Message = "EUID registration " + string(sub.Kind()) + " checksum: " + subRes.Message
	}

	return res, nil
}
