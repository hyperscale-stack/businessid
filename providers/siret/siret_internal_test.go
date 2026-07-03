// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siret

import "testing"

// TestWithDerogationOnNilMap covers the defensive nil-check inside
// WithDerogation. In practice New always seeds the map, so the branch
// is only reachable through direct construction.
func TestWithDerogationOnNilMap(t *testing.T) {
	t.Parallel()

	p := &Provider{}
	opt := WithDerogation("123456789", func(string) bool { return true })
	opt(p)

	if p.derogations == nil {
		t.Fatal("WithDerogation should have allocated the derogations map")
	}

	if _, ok := p.derogations["123456789"]; !ok {
		t.Fatal("WithDerogation did not register the rule")
	}
}
