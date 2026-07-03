// Copyright 2026 Hyperscale. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package siret_test

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/siret"
	"github.com/stretchr/testify/require"
)

// TestValidateRealWorld runs every SIRET in testdata/siret/*.txt
// through the provider and expects both Format and Checksum to be Valid.
func TestValidateRealWorld(t *testing.T) {
	t.Parallel()

	dir := filepath.Join("..", "..", "testdata", "siret")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	p := siret.New()

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "-valide.txt") {
			continue
		}

		lines, err := readCorpus(filepath.Join(dir, e.Name()))
		require.NoError(t, err)

		if len(lines) == 0 {
			t.Errorf("%s: corpus is empty", e.Name())

			continue
		}

		for _, line := range lines {
			t.Run(line, func(t *testing.T) {
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
				require.Equal(t,
					businessid.ValidationStatusValid, checksumRes.Status,
					"Checksum: %s reason=%s message=%s", line, checksumRes.ReasonCode, checksumRes.Message,
				)
			})
		}
	}
}

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
