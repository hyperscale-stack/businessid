// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package euid

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNativeRegisterAlgorithmsRealWorld runs every raw national register
// entry in testdata/registers/<cc>-valide.txt through the country's
// native validator functions (validateFormat + validateChecksum) —
// bypassing the BRIS meta-format layer.
//
// This catches bugs that are specific to a register's algorithm and that
// might otherwise be masked (or unreachable) from the public
// [Provider.ValidateFormat] / [Provider.ValidateChecksum] path.
func TestNativeRegisterAlgorithmsRealWorld(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("..", "..", "testdata", "registers")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "-valide.txt") {
			continue
		}

		cc := strings.ToUpper(strings.TrimSuffix(e.Name(), "-valide.txt"))

		rv, ok := euidRegisterValidators[cc]
		if !ok {
			t.Errorf("no native validator for country %q (file %s)", cc, e.Name())

			continue
		}

		lines, err := readNativeCorpus(filepath.Join(dir, e.Name()))
		require.NoError(t, err)

		if len(lines) == 0 {
			t.Errorf("%s: corpus is empty", e.Name())

			continue
		}

		for _, raw := range lines {
			t.Run(cc+"/"+raw, func(t *testing.T) {
				t.Parallel()

				value := raw
				if rv.canonicalize != nil {
					value = rv.canonicalize(value)
				}

				if rv.validateFormat != nil {
					ok, reason, msg := rv.validateFormat(value)
					require.True(t, ok, "format: %s (canonicalized %q) reason=%s message=%s",
						raw, value, reason, msg)
				}

				if rv.validateChecksum != nil {
					require.True(t,
						rv.validateChecksum(value),
						"checksum failed for %s (canonicalized %q)", raw, value,
					)
				}
			})
		}
	}
}

func readNativeCorpus(path string) ([]string, error) {
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
