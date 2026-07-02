// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

// IdentifierKind names a supported family of business identifier.
type IdentifierKind string

// Supported identifier kinds.
const (
	IdentifierKindEUID                       IdentifierKind = "euid"
	IdentifierKindVAT                        IdentifierKind = "vat"
	IdentifierKindSIREN                      IdentifierKind = "siren"
	IdentifierKindSIRET                      IdentifierKind = "siret"
	IdentifierKindDUNS                       IdentifierKind = "duns"
	IdentifierKindLEI                        IdentifierKind = "lei"
	IdentifierKindEIN                        IdentifierKind = "ein"
	IdentifierKindCompanyNumber              IdentifierKind = "company_number"
	IdentifierKindNationalRegistrationNumber IdentifierKind = "national_registration_number"
	IdentifierKindEORI                       IdentifierKind = "eori"
)

// IdentifierInput is the caller-supplied value to validate.
//
// CountryCode is required by some kinds (VAT without embedded prefix,
// national registration number). When set it must be a 2-letter ISO code;
// case and surrounding whitespace are normalized by providers.
type IdentifierInput struct {
	Kind        IdentifierKind
	Value       string
	CountryCode string
}
