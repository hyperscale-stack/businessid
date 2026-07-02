// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

// ValidationLevel enumerates the checks a validator can perform.
type ValidationLevel string

// Supported validation levels.
const (
	ValidationLevelFormat   ValidationLevel = "format"
	ValidationLevelChecksum ValidationLevel = "checksum"
	ValidationLevelRegistry ValidationLevel = "registry"
)

// ValidationStatus is the outcome of a single validation.
type ValidationStatus string

// Supported validation outcomes.
const (
	ValidationStatusValid       ValidationStatus = "valid"
	ValidationStatusInvalid     ValidationStatus = "invalid"
	ValidationStatusUnsupported ValidationStatus = "unsupported"
	ValidationStatusUnknown     ValidationStatus = "unknown"
)

// ValidationResult is the output of a single validation step.
type ValidationResult struct {
	Kind           IdentifierKind
	Level          ValidationLevel
	Status         ValidationStatus
	InputValue     string
	CanonicalValue string
	CountryCode    string
	ReasonCode     string
	Message        string
}
