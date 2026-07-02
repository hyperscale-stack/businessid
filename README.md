Hyperscale businessid [![Last release](https://img.shields.io/github/release/hyperscale-stack/businessid.svg)](https://github.com/hyperscale-stack/businessid/releases/latest) [![Documentation](https://godoc.org/github.com/hyperscale-stack/businessid?status.svg)](https://godoc.org/github.com/hyperscale-stack/businessid)
====================

[![Go Report Card](https://goreportcard.com/badge/github.com/hyperscale-stack/businessid)](https://goreportcard.com/report/github.com/hyperscale-stack/businessid)

| Branch  | Status | Coverage |
|---------|--------|----------|
| main    | [![Build Status](https://github.com/hyperscale-stack/businessid/workflows/Go/badge.svg?branch=main)](https://github.com/hyperscale-stack/businessid/actions?query=workflow%3AGo) | [![Coveralls](https://img.shields.io/coveralls/hyperscale-stack/businessid/main.svg)](https://coveralls.io/github/hyperscale-stack/businessid?branch=main) |

An offline Go library for validating business identifiers — SIREN, SIRET,
LEI, VAT, DUNS, EIN, UK company number, EUID, EORI, and national
registration numbers. Validation is layered in three levels (`format` →
`checksum` → `registry`) and each identifier is implemented as a small
provider so callers only pull in what they use. No network calls, no
protobuf, no product-specific dependencies.

## Supported identifiers

| Kind                          | Provider package                | Format | Checksum              | Registry |
| ----------------------------- | ------------------------------- | :----: | --------------------- | :------: |
| SIREN                         | `providers/siren`               | ✓      | Luhn                  | —        |
| SIRET                         | `providers/siret`               | ✓      | Luhn                  | —        |
| LEI                           | `providers/lei`                 | ✓      | ISO 7064 Mod 97-10    | —        |
| DUNS                          | `providers/duns`                | ✓      | —                     | —        |
| EIN                           | `providers/ein`                 | ✓      | —                     | —        |
| UK Company Number             | `providers/companynumber`       | ✓      | —                     | —        |
| EUID                          | `providers/euid`                | ✓      | —                     | —        |
| EORI                          | `providers/eori`                | ✓      | —                     | —        |
| VAT                           | `providers/vat`                 | ✓      | FR key + SIREN Luhn   | —        |
| National Registration Number  | `providers/nationalregistration`| ✓      | —                     | —        |
| Defaults wiring               | `defaults`                      | —      | —                     | —        |

`RegistryLookup` is an interface with unsupported defaults today so real
HTTP integrations (INSEE, GLEIF, Companies House…) can be plugged in later
without breaking callers.

## Install

```sh
go get github.com/hyperscale-stack/businessid
```

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/defaults"
)

func main() {
	v := defaults.New()

	results, _ := v.Validate(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindSIREN,
		Value: "552 100 554",
	})
	for _, r := range results {
		fmt.Printf("%s: %s (%s)\n", r.Level, r.Status, r.ReasonCode)
	}

	vatRes, _ := v.Validate(context.Background(), businessid.IdentifierInput{
		Kind:  businessid.IdentifierKindVAT,
		Value: "FR96552100554",
	})
	for _, r := range vatRes {
		fmt.Printf("VAT %s: %s\n", r.Level, r.Status)
	}
}
```

`Validate` runs the format check first and, only if it passes, the
checksum check — so the returned slice contains one or two results. Use
`v.ValidateFormat`, `v.ValidateChecksum` or `v.LookupRegistry` directly
when you want a single level. Every result carries a stable
`ReasonCode` (`businessid.ReasonOK`, `ReasonInvalidLength`,
`ReasonInvalidChecksum`, `ReasonUnsupportedChecksum`, …) that is safe to
switch on.

Register only the providers you actually need:

```go
import (
	"github.com/hyperscale-stack/businessid"
	"github.com/hyperscale-stack/businessid/providers/lei"
	"github.com/hyperscale-stack/businessid/providers/siren"
)

v := businessid.NewValidator(
	businessid.WithProvider(siren.New()),
	businessid.WithProvider(lei.New()),
)
```

## Development

```sh
make build    # go build ./...
make test     # race + coverage
make lint     # golangci-lint with the shared config
```

## License

Hyperscale businessid is licensed under [the MIT license](LICENSE.md).
