// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

// Reason codes. These are stable string constants and safe to switch on.
const (
	ReasonOK                  = "ok"
	ReasonEmpty               = "empty"
	ReasonUnsupportedKind     = "unsupported_kind"
	ReasonMissingCountryCode  = "missing_country_code"
	ReasonInvalidLength       = "invalid_length"
	ReasonInvalidCharacters   = "invalid_characters"
	ReasonInvalidFormat       = "invalid_format"
	ReasonInvalidChecksum     = "invalid_checksum"
	ReasonUnsupportedChecksum = "unsupported_checksum"
	ReasonUnsupportedRegistry = "unsupported_registry"
	ReasonCountryMismatch     = "country_mismatch"
)
