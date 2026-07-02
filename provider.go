// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

import "context"

// Provider is the minimal contract implemented by every identifier kind.
//
// Optional behaviors are expressed by additionally implementing
// [FormatValidator], [ChecksumValidator] or [RegistryLookup].
type Provider interface {
	Kind() IdentifierKind
	Capabilities() Capabilities
	Canonicalize(input IdentifierInput) IdentifierInput
}

// FormatValidator is implemented by providers that can check syntactic form.
type FormatValidator interface {
	ValidateFormat(ctx context.Context, input IdentifierInput) (*ValidationResult, error)
}

// ChecksumValidator is implemented by providers that can verify a
// self-consistency check (Luhn, Mod-97-10, country-specific...).
type ChecksumValidator interface {
	ValidateChecksum(ctx context.Context, input IdentifierInput) (*ValidationResult, error)
}

// RegistryLookup is implemented by providers that can query an external
// registry. Implementations may perform network I/O.
type RegistryLookup interface {
	LookupRegistry(ctx context.Context, input IdentifierInput) (*RegistryResult, error)
}
