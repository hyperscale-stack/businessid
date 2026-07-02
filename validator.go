// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

import (
	"context"
	"fmt"
)

const msgNoProviderRegistered = "no provider registered for kind"

// Option configures a Validator at construction time.
type Option func(*Validator)

// WithProvider registers a single provider.
//
// If two providers share the same Kind the last one wins.
func WithProvider(p Provider) Option {
	return func(v *Validator) {
		if p == nil {
			return
		}

		v.providers[p.Kind()] = p
	}
}

// WithProviders registers several providers.
func WithProviders(ps ...Provider) Option {
	return func(v *Validator) {
		for _, p := range ps {
			if p == nil {
				continue
			}

			v.providers[p.Kind()] = p
		}
	}
}

// Validator dispatches validation requests to registered providers.
//
// The zero-value is not usable — construct one with [NewValidator].
type Validator struct {
	providers map[IdentifierKind]Provider
}

// NewValidator creates a Validator configured by the given options.
//
// A Validator with no providers registered will report
// [ValidationStatusUnsupported] for every request.
func NewValidator(options ...Option) *Validator {
	v := &Validator{
		providers: make(map[IdentifierKind]Provider),
	}

	for _, opt := range options {
		opt(v)
	}

	return v
}

// Provider returns the provider registered for kind, if any.
func (v *Validator) Provider(kind IdentifierKind) (Provider, bool) {
	p, ok := v.providers[kind]

	return p, ok
}

// ValidateFormat runs the format check for input.Kind.
//
// It never returns an error together with a nil result: on success it always
// returns a non-nil ValidationResult.
func (v *Validator) ValidateFormat(ctx context.Context, input IdentifierInput) (*ValidationResult, error) {
	p, ok := v.providers[input.Kind]
	if !ok {
		return unsupportedKindResult(input, ValidationLevelFormat), nil
	}

	fv, ok := p.(FormatValidator)
	if !ok {
		return unsupportedResult(input, p, ValidationLevelFormat, ReasonUnsupportedKind, "format validation not supported"), nil
	}

	canonical := p.Canonicalize(input)

	res, err := fv.ValidateFormat(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("format validate %q: %w", input.Kind, err)
	}

	return res, nil
}

// ValidateChecksum runs the checksum check for input.Kind.
//
// Format is checked first; if it fails, the returned result carries the
// format failure at level [ValidationLevelChecksum] with
// [ValidationStatusInvalid]. If the provider does not implement checksum,
// [ValidationStatusUnsupported] is returned.
func (v *Validator) ValidateChecksum(ctx context.Context, input IdentifierInput) (*ValidationResult, error) {
	p, ok := v.providers[input.Kind]
	if !ok {
		return unsupportedKindResult(input, ValidationLevelChecksum), nil
	}

	canonical := p.Canonicalize(input)

	if fv, ok := p.(FormatValidator); ok {
		formatRes, err := fv.ValidateFormat(ctx, canonical)
		if err != nil {
			return nil, fmt.Errorf("format validate %q: %w", input.Kind, err)
		}

		if formatRes.Status != ValidationStatusValid {
			out := *formatRes
			out.Level = ValidationLevelChecksum

			return &out, nil
		}
	}

	cv, ok := p.(ChecksumValidator)
	if !ok {
		return unsupportedResult(input, p, ValidationLevelChecksum, ReasonUnsupportedChecksum, "checksum validation not supported"), nil
	}

	res, err := cv.ValidateChecksum(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("checksum validate %q: %w", input.Kind, err)
	}

	return res, nil
}

// Validate runs format then, if format succeeded, checksum.
//
// The returned slice has one or two entries. It is never nil and never empty.
func (v *Validator) Validate(ctx context.Context, input IdentifierInput) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, 2)

	formatRes, err := v.ValidateFormat(ctx, input)
	if err != nil {
		return nil, err
	}

	results = append(results, *formatRes)

	if formatRes.Status != ValidationStatusValid {
		return results, nil
	}

	checksumRes, err := v.ValidateChecksum(ctx, input)
	if err != nil {
		return nil, err
	}

	results = append(results, *checksumRes)

	return results, nil
}

// LookupRegistry dispatches to the provider's registry lookup if implemented.
func (v *Validator) LookupRegistry(ctx context.Context, input IdentifierInput) (*RegistryResult, error) {
	p, ok := v.providers[input.Kind]
	if !ok {
		return &RegistryResult{
			Status:         ValidationStatusUnsupported,
			CanonicalValue: input.Value,
			ReasonCode:     ReasonUnsupportedKind,
			Message:        msgNoProviderRegistered,
		}, nil
	}

	canonical := p.Canonicalize(input)

	rl, ok := p.(RegistryLookup)
	if !ok {
		return &RegistryResult{
			Status:         ValidationStatusUnsupported,
			CanonicalValue: canonical.Value,
			ReasonCode:     ReasonUnsupportedRegistry,
			Message:        "registry lookup not supported",
		}, nil
	}

	res, err := rl.LookupRegistry(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("registry lookup %q: %w", input.Kind, err)
	}

	return res, nil
}

func unsupportedKindResult(input IdentifierInput, level ValidationLevel) *ValidationResult {
	return &ValidationResult{
		Kind:           input.Kind,
		Level:          level,
		Status:         ValidationStatusUnsupported,
		InputValue:     input.Value,
		CanonicalValue: input.Value,
		CountryCode:    NormalizeCountryCode(input.CountryCode),
		ReasonCode:     ReasonUnsupportedKind,
		Message:        msgNoProviderRegistered,
	}
}

func unsupportedResult(input IdentifierInput, p Provider, level ValidationLevel, reason, msg string) *ValidationResult {
	canonical := p.Canonicalize(input)

	return &ValidationResult{
		Kind:           input.Kind,
		Level:          level,
		Status:         ValidationStatusUnsupported,
		InputValue:     input.Value,
		CanonicalValue: canonical.Value,
		CountryCode:    NormalizeCountryCode(canonical.CountryCode),
		ReasonCode:     reason,
		Message:        msg,
	}
}
