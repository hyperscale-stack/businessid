// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package businessid

// Capabilities advertises what a provider can validate.
//
// Providers that cannot honor a capability should return
// [ValidationStatusUnsupported] instead of panicking.
type Capabilities struct {
	Format   bool
	Checksum bool
	Registry bool
}
