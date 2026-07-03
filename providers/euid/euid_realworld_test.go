// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid_test

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/euid"
	"github.com/stretchr/testify/require"
)

// checksumUnsupportedCountries lists the EU-27 countries whose national
// register has no publicly documented checksum algorithm. For these,
// ValidateChecksum is expected to return ValidationStatusUnsupported —
// not Invalid.
var checksumUnsupportedCountries = map[string]struct{}{
	"HR": {}, "CY": {}, "DE": {}, "EL": {},
	"HU": {}, "IE": {}, "LU": {}, "MT": {},
	"NL": {}, "PL": {}, "SI": {},
}

// TestValidateRealWorld runs every EUID in testdata/euid/*.txt through
// the provider (Format + Checksum) and expects Valid on Format for all,
// Valid on Checksum where the register has a published algorithm, and
// Unsupported (not Invalid) for the registers listed in
// checksumUnsupportedCountries.
func TestValidateRealWorld(t *testing.T) {
	t.Parallel()

	dir := repoTestdataDir(t, "euid")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	p := euid.New()

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "-valide.txt") {
			continue
		}

		cc := strings.ToUpper(strings.TrimSuffix(e.Name(), "-valide.txt"))

		lines, err := readCorpus(filepath.Join(dir, e.Name()))
		require.NoError(t, err)

		if len(lines) == 0 {
			t.Errorf("%s: corpus is empty", e.Name())

			continue
		}

		for _, line := range lines {
			t.Run(cc+"/"+line, func(t *testing.T) {
				t.Parallel()

				canonical := p.Canonicalize(businessid.IdentifierInput{Value: line})

				formatRes, err := p.ValidateFormat(context.Background(), canonical)
				require.NoError(t, err)
				require.Equal(t,
					businessid.ValidationStatusValid, formatRes.Status,
					"Format: %s reason=%s message=%s", line, formatRes.ReasonCode, formatRes.Message,
				)

				checksumRes, err := p.ValidateChecksum(context.Background(), canonical)
				require.NoError(t, err)

				if _, unsupported := checksumUnsupportedCountries[cc]; unsupported {
					require.Equal(t,
						businessid.ValidationStatusUnsupported, checksumRes.Status,
						"Checksum for %s should be Unsupported (no published algo), got %s (%s)",
						cc, checksumRes.Status, checksumRes.ReasonCode,
					)
				} else {
					require.Equal(t,
						businessid.ValidationStatusValid, checksumRes.Status,
						"Checksum: %s reason=%s message=%s", line, checksumRes.ReasonCode, checksumRes.Message,
					)
				}
			})
		}
	}
}

// repoTestdataDir returns the absolute path to <repo>/testdata/<kind>.
// It resolves relative to the test file's package directory so tests
// remain runnable from anywhere (`go test ./...`, IDE, etc.).
func repoTestdataDir(t *testing.T, kind string) string {
	t.Helper()

	// The test binary runs with cwd = package directory
	// (providers/euid). Walk up two levels to reach the module root.
	return filepath.Join("..", "..", "testdata", kind)
}

// readCorpus returns non-empty, non-comment lines from a corpus file.
func readCorpus(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []string

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		out = append(out, line)
	}

	return out, sc.Err()
}
