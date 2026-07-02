// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

import (
	"context"
	"fmt"
	"reflect"
)

const msgNoProviderRegistered = "no provider registered for kind"

// Option configures a Validator at construction time.
type Option func(*Validator)

// WithProvider registers a single provider.
//
// If two providers share the same Kind the last one wins. Typed-nil
// providers (e.g. `(*siren.Provider)(nil)` boxed in a Provider interface)
// are treated the same as an untyped nil and silently skipped.
func WithProvider(p Provider) Option {
	return func(v *Validator) {
		if isNilProvider(p) {
			return
		}

		v.providers[p.Kind()] = p
	}
}

// WithProviders registers several providers.
func WithProviders(ps ...Provider) Option {
	return func(v *Validator) {
		for _, p := range ps {
			if isNilProvider(p) {
				continue
			}

			v.providers[p.Kind()] = p
		}
	}
}

// isNilProvider returns true when p is either an untyped nil interface or a
// typed nil (e.g. `(*siren.Provider)(nil)` boxed in a Provider interface).
// The plain `p == nil` check only catches the first form.
func isNilProvider(p Provider) bool {
	if p == nil {
		return true
	}

	rv := reflect.ValueOf(p)

	switch rv.Kind() { //nolint:exhaustive // only reference-like kinds can be nil.
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
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

	canonical := p.Canonicalize(input)

	return v.validateFormatWithCanonical(ctx, input, canonical, p)
}

// validateFormatWithCanonical runs the format check on an already-canonicalized
// input. originalInput carries the caller's raw value for the InputValue
// surface field.
func (v *Validator) validateFormatWithCanonical(ctx context.Context, originalInput, canonical IdentifierInput, p Provider) (*ValidationResult, error) {
	fv, ok := p.(FormatValidator)
	if !ok {
		return unsupportedResult(originalInput, canonical, ValidationLevelFormat, ReasonUnsupportedFormat, "format validation not supported"), nil
	}

	res, err := fv.ValidateFormat(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("format validate %q: %w", originalInput.Kind, err)
	}

	if res == nil {
		return nil, fmt.Errorf("format validate %q: provider returned nil result", originalInput.Kind)
	}

	res.InputValue = originalInput.Value

	return res, nil
}

// ValidateChecksum runs the checksum check for input.Kind.
//
// Format is checked first; if it fails, the returned result carries the
// format failure at level [ValidationLevelChecksum] with
// [ValidationStatusInvalid]. If the provider does not implement checksum,
// [ValidationStatusUnsupported] is returned. If the provider implements
// checksum but not format, checksum is also reported as unsupported — the
// format gate is a hard requirement.
func (v *Validator) ValidateChecksum(ctx context.Context, input IdentifierInput) (*ValidationResult, error) {
	p, ok := v.providers[input.Kind]
	if !ok {
		return unsupportedKindResult(input, ValidationLevelChecksum), nil
	}

	canonical := p.Canonicalize(input)

	return v.validateChecksumWithCanonical(ctx, input, canonical, p)
}

// validateChecksumWithCanonical runs the checksum check on an already-
// canonicalized input. It enforces the format-first invariant and injects
// the raw caller value into the returned InputValue.
func (v *Validator) validateChecksumWithCanonical(ctx context.Context, originalInput, canonical IdentifierInput, p Provider) (*ValidationResult, error) {
	fv, hasFormat := p.(FormatValidator)
	if !hasFormat {
		return unsupportedResult(originalInput, canonical, ValidationLevelChecksum, ReasonUnsupportedFormat, "format validation not supported, cannot gate checksum"), nil
	}

	formatRes, err := fv.ValidateFormat(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("format validate %q: %w", originalInput.Kind, err)
	}

	if formatRes == nil {
		return nil, fmt.Errorf("format validate %q: provider returned nil result", originalInput.Kind)
	}

	if formatRes.Status != ValidationStatusValid {
		out := *formatRes
		out.Level = ValidationLevelChecksum
		out.InputValue = originalInput.Value

		return &out, nil
	}

	cv, ok := p.(ChecksumValidator)
	if !ok {
		return unsupportedResult(originalInput, canonical, ValidationLevelChecksum, ReasonUnsupportedChecksum, "checksum validation not supported"), nil
	}

	res, err := cv.ValidateChecksum(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("checksum validate %q: %w", originalInput.Kind, err)
	}

	if res == nil {
		return nil, fmt.Errorf("checksum validate %q: provider returned nil result", originalInput.Kind)
	}

	res.InputValue = originalInput.Value

	return res, nil
}

// Validate runs format then, if format succeeded, checksum.
//
// The returned slice has one or two entries. It is never nil and never empty.
// Canonicalization is performed exactly once per call and the format result
// is reused — the checksum step never re-runs format.
func (v *Validator) Validate(ctx context.Context, input IdentifierInput) ([]ValidationResult, error) {
	results := make([]ValidationResult, 0, 2)

	p, ok := v.providers[input.Kind]
	if !ok {
		results = append(results, *unsupportedKindResult(input, ValidationLevelFormat))

		return results, nil
	}

	canonical := p.Canonicalize(input)

	formatRes, err := v.validateFormatWithCanonical(ctx, input, canonical, p)
	if err != nil {
		return nil, err
	}

	results = append(results, *formatRes)

	if formatRes.Status != ValidationStatusValid {
		return results, nil
	}

	cv, ok := p.(ChecksumValidator)
	if !ok {
		results = append(results, *unsupportedResult(input, canonical, ValidationLevelChecksum, ReasonUnsupportedChecksum, "checksum validation not supported"))

		return results, nil
	}

	checksumRes, err := cv.ValidateChecksum(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("checksum validate %q: %w", input.Kind, err)
	}

	if checksumRes == nil {
		return nil, fmt.Errorf("checksum validate %q: provider returned nil result", input.Kind)
	}

	checksumRes.InputValue = input.Value
	results = append(results, *checksumRes)

	return results, nil
}

// LookupRegistry dispatches to the provider's registry lookup if implemented.
func (v *Validator) LookupRegistry(ctx context.Context, input IdentifierInput) (*RegistryResult, error) {
	p, ok := v.providers[input.Kind]
	if !ok {
		return &RegistryResult{
			Kind:           input.Kind,
			Status:         ValidationStatusUnsupported,
			InputValue:     input.Value,
			CanonicalValue: input.Value,
			ReasonCode:     ReasonUnsupportedKind,
			Message:        msgNoProviderRegistered,
		}, nil
	}

	canonical := p.Canonicalize(input)

	rl, ok := p.(RegistryLookup)
	if !ok {
		return &RegistryResult{
			Kind:           input.Kind,
			Status:         ValidationStatusUnsupported,
			InputValue:     input.Value,
			CanonicalValue: canonical.Value,
			ReasonCode:     ReasonUnsupportedRegistry,
			Message:        "registry lookup not supported",
		}, nil
	}

	res, err := rl.LookupRegistry(ctx, canonical)
	if err != nil {
		return nil, fmt.Errorf("registry lookup %q: %w", input.Kind, err)
	}

	if res == nil {
		return nil, fmt.Errorf("registry lookup %q: provider returned nil result", input.Kind)
	}

	res.Kind = input.Kind
	res.InputValue = input.Value

	return res, nil
}

// unsupportedKindResult builds a result for a kind with no registered
// provider. CanonicalValue is the raw input.Value because no provider is
// available to canonicalize.
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

// unsupportedResult builds a result for a kind whose registered provider
// does not support the requested level. Caller must pass the already-
// canonicalized input alongside the raw one.
func unsupportedResult(originalInput, canonical IdentifierInput, level ValidationLevel, reason, msg string) *ValidationResult {
	return &ValidationResult{
		Kind:           originalInput.Kind,
		Level:          level,
		Status:         ValidationStatusUnsupported,
		InputValue:     originalInput.Value,
		CanonicalValue: canonical.Value,
		CountryCode:    canonical.CountryCode,
		ReasonCode:     reason,
		Message:        msg,
	}
}
