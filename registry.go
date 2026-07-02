// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

import "time"

// RegistryAddress is the postal address returned by a registry lookup.
type RegistryAddress struct {
	Line1       string
	Line2       string
	PostalCode  string
	City        string
	Region      string
	CountryCode string
}

// RegistryCompanyProfile describes an entity as reported by an external registry.
//
// Raw carries the untouched registry payload for callers that need to inspect
// provider-specific fields.
type RegistryCompanyProfile struct {
	IdentifierKind  IdentifierKind
	IdentifierValue string
	CountryCode     string

	LegalName      string
	TradingName    string
	RegistrationID string
	VATID          string
	LEI            string
	DUNS           string

	Status    string
	LegalForm string
	Address   RegistryAddress

	RegistryName  string
	RegistryURL   string
	LastUpdatedAt *time.Time

	Raw map[string]any
}

// RegistryResult is the outcome of a registry lookup.
type RegistryResult struct {
	Status         ValidationStatus
	Profile        *RegistryCompanyProfile
	CanonicalValue string
	ReasonCode     string
	Message        string
}
