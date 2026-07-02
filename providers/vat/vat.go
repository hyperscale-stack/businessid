// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package vat validates EU VAT numbers, with a country-specific
// implementation for France.
//
// Supported layouts:
//
//   - FR: FR + 2 alphanumeric key characters + 9-digit SIREN
//   - Generic EU: 2-letter prefix + 2..13 upper-case alphanumeric characters
package vat

import (
	"context"

	"github.com/hyperscale-stack/businessid"
)

const (
	frLen        = 13
	genericMin   = 4
	genericMax   = 15
	prefixLength = 2
)

const (
	msgMissingCountryPrefix = "VAT number must start with a 2-letter country code"
	msgFRVATLayout          = "FR VAT must be FR + 2 key chars + 9-digit SIREN"
	msgCountryMismatch      = "VAT prefix does not match provided country code"
	msgEmpty                = "empty value"
)

// Provider validates VAT numbers.
type Provider struct{}

// New returns a new VAT provider.
func New() *Provider { return &Provider{} }

// Kind implements [businessid.Provider].
func (Provider) Kind() businessid.IdentifierKind { return businessid.IdentifierKindVAT }

// Capabilities implements [businessid.Provider].
func (Provider) Capabilities() businessid.Capabilities {
	return businessid.Capabilities{Format: true, Checksum: true, Registry: false}
}

// Canonicalize trims, upper-cases, and strips whitespace, dots and dashes.
// When the value is non-empty and does not already begin with a two-letter
// prefix, the caller-supplied country code (normalized) is prepended. An
// empty value is passed through unchanged so downstream callers can surface
// ReasonEmpty rather than a length error.
func (Provider) Canonicalize(input businessid.IdentifierInput) businessid.IdentifierInput {
	value := businessid.StripSeparators(
		businessid.StripAllSpaces(businessid.TrimUpper(input.Value)),
		".", "-",
	)

	cc := businessid.NormalizeCountryCode(input.CountryCode)

	if value != "" && !businessid.IsASCIICountryPrefix(value) && cc != "" {
		value = cc + value
	}

	input.Value = value
	input.CountryCode = cc

	return input
}

// ValidateFormat implements [businessid.FormatValidator].
func (Provider) ValidateFormat(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindVAT,
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

	if !businessid.IsASCIICountryPrefix(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonMissingCountryCode
		res.Message = msgMissingCountryPrefix

		return res, nil
	}

	prefix := input.Value[:prefixLength]

	if input.CountryCode != "" && input.CountryCode != prefix {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonCountryMismatch
		res.Message = msgCountryMismatch
		res.CountryCode = prefix

		return res, nil
	}

	res.CountryCode = prefix

	body := input.Value[prefixLength:]

	if prefix == "FR" {
		validateFRFormat(res, body)

		return res, nil
	}

	if len(body) < genericMin-prefixLength || len(body) > genericMax-prefixLength {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = "generic EU VAT body length out of range"

		return res, nil
	}

	if !businessid.IsAlnumUpper(body) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "VAT body must be alphanumeric"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// validateFRFormat validates the FR VAT layout on the body (everything after
// the FR prefix). It mutates res in place.
func validateFRFormat(res *businessid.ValidationResult, body string) {
	if len(body) != frLen-prefixLength {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = msgFRVATLayout

		return
	}

	key := body[:2]
	siren := body[2:]

	if !businessid.IsAlnumUpper(key) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "FR VAT key must be alphanumeric"

		return
	}

	if !businessid.IsAllDigits(siren) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidCharacters
		res.Message = "FR VAT SIREN portion must be 9 digits"

		return
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK
}

// ValidateChecksum implements [businessid.ChecksumValidator].
func (Provider) ValidateChecksum(_ context.Context, input businessid.IdentifierInput) (*businessid.ValidationResult, error) {
	res := &businessid.ValidationResult{
		Kind:           businessid.IdentifierKindVAT,
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

	if !businessid.IsASCIICountryPrefix(input.Value) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonMissingCountryCode
		res.Message = msgMissingCountryPrefix

		return res, nil
	}

	prefix := input.Value[:prefixLength]

	if input.CountryCode != "" && input.CountryCode != prefix {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonCountryMismatch
		res.Message = msgCountryMismatch
		res.CountryCode = prefix

		return res, nil
	}

	res.CountryCode = prefix

	body := input.Value[prefixLength:]

	if prefix != "FR" {
		res.Status = businessid.ValidationStatusUnsupported
		res.ReasonCode = businessid.ReasonUnsupportedChecksum
		res.Message = "VAT checksum not implemented for this country"

		return res, nil
	}

	if len(body) != frLen-prefixLength {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidLength
		res.Message = msgFRVATLayout

		return res, nil
	}

	key := body[:2]
	siren := body[2:]

	if !businessid.IsAllDigits(key) {
		res.Status = businessid.ValidationStatusUnsupported
		res.ReasonCode = businessid.ReasonUnsupportedChecksum
		res.Message = "FR VAT alphanumeric key checksum not supported"

		return res, nil
	}

	if !businessid.IsAllDigits(siren) || !businessid.Luhn(siren) {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidChecksum
		res.Message = "FR VAT embedded SIREN checksum failed"

		return res, nil
	}

	sirenInt := parseDigits(siren)
	keyInt := parseDigits(key)
	expected := (12 + 3*(sirenInt%97)) % 97

	if keyInt != expected {
		res.Status = businessid.ValidationStatusInvalid
		res.ReasonCode = businessid.ReasonInvalidChecksum
		res.Message = "FR VAT numeric key does not match SIREN"

		return res, nil
	}

	res.Status = businessid.ValidationStatusValid
	res.ReasonCode = businessid.ReasonOK

	return res, nil
}

// parseDigits converts a digit-only string to an int. The caller must ensure
// every rune is '0'..'9' and that the value fits in a machine word.
func parseDigits(s string) int {
	n := 0

	for i := range len(s) {
		n = n*10 + int(s[i]-'0')
	}

	return n
}
