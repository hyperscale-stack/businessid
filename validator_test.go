// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/defaults"
	"github.com/hyperscale-stack/businessid/providers/siren"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errBoom is a sentinel used by the fake providers below to force error paths.
var errBoom = errors.New("boom")

const testKind businessid.IdentifierKind = "test-kind"

// bareProvider satisfies only [businessid.Provider] — no optional interfaces.
type bareProvider struct{}

func (bareProvider) Kind() businessid.IdentifierKind { return testKind }
func (bareProvider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{}
}

func (bareProvider) Canonicalize(in businessid.IdentifierInput) businessid.IdentifierInput {
	return in
}

// errorFormatProvider returns an error from ValidateFormat.
type errorFormatProvider struct{ bareProvider }

func (errorFormatProvider) ValidateFormat(context.Context, businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	return nil, errBoom
}

// errorChecksumProvider always passes format but errors on checksum.
type errorChecksumProvider struct{ bareProvider }

func (errorChecksumProvider) ValidateFormat(_ context.Context, in businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	return &businessid.ValidationResult{
		Kind:       in.Kind,
		Level:      businessid.ValidationLevelFormat,
		Status:     businessid.ValidationStatusValid,
		ReasonCode: businessid.ReasonOK,
	}, nil
}

func (errorChecksumProvider) ValidateChecksum(context.Context, businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	return nil, errBoom
}

// successRegistryProvider returns a canned profile.
type successRegistryProvider struct{ bareProvider }

func (successRegistryProvider) LookupRegistry(_ context.Context, in businessid.IdentifierInput) (*businessid.RegistryResult, error) {
	return &businessid.RegistryResult{
		Status:         businessid.ValidationStatusValid,
		CanonicalValue: in.Value,
		ReasonCode:     businessid.ReasonOK,
	}, nil
}

// errorRegistryProvider returns an error from LookupRegistry.
type errorRegistryProvider struct{ bareProvider }

func (errorRegistryProvider) LookupRegistry(context.Context, businessid.IdentifierInput) (*businessid.RegistryResult, error) {
	return nil, errBoom
}

func TestValidator_UnsupportedKind(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator()

	res, err := v.ValidateFormat(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "552100554",
	})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusUnsupported, res.Status)
	assert.Equal(t, businessid.ReasonUnsupportedKind, res.ReasonCode)
}

func TestValidator_ValidateFormatValid(t *testing.T) {
	t.Parallel()

	v := defaults.New()

	res, err := v.ValidateFormat(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "552 100 554",
	})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusValid, res.Status)
	assert.Equal(t, "552100554", res.CanonicalValue)
}

func TestValidator_ValidateReturnsFormatOnlyWhenFormatFails(t *testing.T) {
	t.Parallel()

	v := defaults.New()

	results, err := v.Validate(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "not-a-number",
	})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, businessid.ValidationLevelFormat, results[0].Level)
	assert.Equal(t, businessid.ValidationStatusInvalid, results[0].Status)
}

func TestValidator_ValidateReturnsFormatAndChecksumWhenFormatOK(t *testing.T) {
	t.Parallel()

	v := defaults.New()

	results, err := v.Validate(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "552100554",
	})
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, businessid.ValidationLevelFormat, results[0].Level)
	assert.Equal(t, businessid.ValidationStatusValid, results[0].Status)
	assert.Equal(t, businessid.ValidationLevelChecksum, results[1].Level)
	assert.Equal(t, businessid.ValidationStatusValid, results[1].Status)
}

func TestValidator_ChecksumUnsupported(t *testing.T) {
	t.Parallel()

	v := defaults.New()

	res, err := v.ValidateChecksum(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindDUNS,
		Value: "123456789",
	})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusUnsupported, res.Status)
	assert.Equal(t, businessid.ReasonUnsupportedChecksum, res.ReasonCode)
}

func TestValidator_ChecksumSkippedWhenFormatFails(t *testing.T) {
	t.Parallel()

	v := defaults.New()

	res, err := v.ValidateChecksum(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "abc",
	})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusInvalid, res.Status)
	assert.Equal(t, businessid.ValidationLevelChecksum, res.Level)
}

func TestValidator_LookupRegistryUnsupported(t *testing.T) {
	t.Parallel()

	v := defaults.New()

	res, err := v.LookupRegistry(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "552100554",
	})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusUnsupported, res.Status)
	assert.Equal(t, businessid.ReasonUnsupportedRegistry, res.ReasonCode)
}

func TestValidator_LookupRegistryUnknownKind(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator()

	res, err := v.LookupRegistry(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "552100554",
	})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusUnsupported, res.Status)
	assert.Equal(t, businessid.ReasonUnsupportedKind, res.ReasonCode)
}

func TestValidator_WithProvider(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(siren.New()))

	_, ok := v.Provider(businessid.IdentifierKindSIREN)
	assert.True(t, ok)

	_, ok = v.Provider(businessid.IdentifierKindVAT)
	assert.False(t, ok)
}

func TestValidator_WithProviderNil(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(nil), businessid.WithProviders(nil, nil))

	_, ok := v.Provider(businessid.IdentifierKindSIREN)
	assert.False(t, ok)
}

func TestValidator_ChecksumUnknownKind(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator()

	res, err := v.ValidateChecksum(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "552100554",
	})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusUnsupported, res.Status)
	assert.Equal(t, businessid.ReasonUnsupportedKind, res.ReasonCode)
}

func TestValidator_FormatUnsupportedByProvider(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(bareProvider{}))

	res, err := v.ValidateFormat(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "x"})
	require.NoError(t, err)
	assert.Equal(t, businessid.ValidationStatusUnsupported, res.Status)
	assert.Equal(t, businessid.ReasonUnsupportedKind, res.ReasonCode)
}

func TestValidator_ValidateFormatWrapsProviderError(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(errorFormatProvider{}))

	res, err := v.ValidateFormat(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "x"})
	assert.Nil(t, res)
	require.ErrorIs(t, err, errBoom)
}

func TestValidator_ValidateChecksumWrapsFormatError(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(errorFormatProvider{}))

	res, err := v.ValidateChecksum(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "x"})
	assert.Nil(t, res)
	require.ErrorIs(t, err, errBoom)
}

func TestValidator_ValidateChecksumWrapsChecksumError(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(errorChecksumProvider{}))

	res, err := v.ValidateChecksum(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "x"})
	assert.Nil(t, res)
	require.ErrorIs(t, err, errBoom)
}

func TestValidator_ValidateWrapsError(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(errorFormatProvider{}))

	results, err := v.Validate(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "x"})
	assert.Nil(t, results)
	require.ErrorIs(t, err, errBoom)
}

func TestValidator_ValidateWrapsChecksumError(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(errorChecksumProvider{}))

	results, err := v.Validate(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "x"})
	assert.Nil(t, results)
	require.ErrorIs(t, err, errBoom)
}

func TestValidator_LookupRegistrySuccess(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(successRegistryProvider{}))

	res, err := v.LookupRegistry(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "abc"})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, businessid.ValidationStatusValid, res.Status)
	assert.Equal(t, "abc", res.CanonicalValue)
}

func TestValidator_LookupRegistryWrapsError(t *testing.T) {
	t.Parallel()

	v := businessid.NewValidator(businessid.WithProvider(errorRegistryProvider{}))

	res, err := v.LookupRegistry(context.Background(), businessid.IdentifierInput{Kind: testKind, Value: "x"})
	assert.Nil(t, res)
	require.ErrorIs(t, err, errBoom)
}

func TestDefaultsCoversEveryKind(t *testing.T) {
	t.Parallel()

	v := defaults.New()

	kinds := []businessid.IdentifierKind{
		businessid.IdentifierKindSIREN,
		businessid.IdentifierKindSIRET,
		businessid.IdentifierKindLEI,
		businessid.IdentifierKindDUNS,
		businessid.IdentifierKindEIN,
		businessid.IdentifierKindCompanyNumber,
		businessid.IdentifierKindEUID,
		businessid.IdentifierKindEORI,
		businessid.IdentifierKindVAT,
		businessid.IdentifierKindNationalRegistrationNumber,
	}

	for _, k := range kinds {
		_, ok := v.Provider(k)
		assert.Truef(t, ok, "kind %q should have a default provider", k)
	}
}
