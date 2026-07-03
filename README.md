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
| EUID                          | `providers/euid`                | ✓      | Delegated to sub-validator | — |
| EORI                          | `providers/eori`                | ✓      | —                     | —        |
| VAT                           | `providers/vat`                 | ✓      | 28 countries (EU-27 + XI, GB) | —        |
| National Registration Number  | `providers/nationalregistration`| ✓      | —                     | —        |
| Defaults wiring               | `defaults`                      | —      | —                     | —        |

`RegistryLookup` is an interface with unsupported defaults today so real
HTTP integrations (INSEE, GLEIF, Companies House…) can be plugged in later
without breaking callers.

## Country coverage

### VAT (format validation)

Every layout below is enforced strictly — length AND character positions
must match, otherwise the result is `ReasonInvalidLength` or
`ReasonInvalidCharacters`. Prefixes not listed fall back to the generic
`2..13 alphanumeric` rule.

| Country              | Code | Layout                              |
| -------------------- | :--: | ----------------------------------- |
| Austria              | AT   | `U` + 8 digits                      |
| Belgium              | BE   | 10 digits (first is 0 or 1)         |
| Bulgaria             | BG   | 9 or 10 digits                      |
| Croatia              | HR   | 11 digits                           |
| Cyprus               | CY   | 8 digits + 1 letter                 |
| Czech Republic       | CZ   | 8, 9 or 10 digits                   |
| Germany              | DE   | 9 digits                            |
| Denmark              | DK   | 8 digits                            |
| Estonia              | EE   | 9 digits                            |
| Greece               | EL   | 9 digits (note: `EL`, not `GR`)     |
| Spain                | ES   | alnum + 7 digits + alnum            |
| Finland              | FI   | 8 digits                            |
| France               | FR   | 2 alnum key + 9-digit SIREN (Luhn + mod-97 checksum) |
| Hungary              | HU   | 8 digits                            |
| Ireland              | IE   | 7 digits + 1-2 letters, or 1 digit + 1 letter + 5 digits + 1 letter |
| Italy                | IT   | 11 digits                           |
| Lithuania            | LT   | 9 or 12 digits                      |
| Luxembourg           | LU   | 8 digits                            |
| Latvia               | LV   | 11 digits                           |
| Malta                | MT   | 8 digits                            |
| Netherlands          | NL   | 9 digits + `B` + 2 digits           |
| Poland               | PL   | 10 digits                           |
| Portugal             | PT   | 9 digits                            |
| Romania              | RO   | 2 to 10 digits                      |
| Sweden               | SE   | 12 digits                           |
| Slovenia             | SI   | 8 digits                            |
| Slovakia             | SK   | 10 digits                           |
| United Kingdom       | GB   | 9 or 12 digits                      |
| Northern Ireland     | XI   | Same as GB (post-Brexit protocol)   |
| Norway               | NO   | 9 digits (org. number)              |
| Iceland              | IS   | 5 or 6 digits                       |
| Liechtenstein        | LI   | 5 digits                            |

Checksum validation (`ValidateChecksum`) is implemented for **28 codes**:
EU-27 + `XI` and `GB` (post-Brexit). Each algorithm is sourced from the
national tax authority documentation and cross-verified against
`python-stdnum` test vectors. `NO`, `IS`, `LI` remain format-only (no
published algorithm we could verify). See
[providers/vat/checksums.go](providers/vat/checksums.go) for per-country
source references.

**Aliases**: `GR` (ISO Greek code) is canonicalized to `EL` (official VAT
prefix). `IE` legacy VAT numbers containing `+` or `*` in position 2 are
accepted with the `vat.WithLegacy()` option.

### EUID (meta-format)

`ValidateFormat` enforces the BRIS layout
`<CC><REGISTER>.<REGISTRATION>` (Regulation (EU) 2015/884) for the EU-27.
The register segment is 1–20 upper-case alphanumeric characters; the
registration segment is 1–64 characters from `[A-Z0-9./\- +]`.

**Meta-format delegation**: the country code selects a national
sub-validator (injected via `euid.WithSubValidator`) which validates the
REGISTRATION segment against its own rules. Today `FR` delegates to the
SIREN provider (registration must be a 9-digit Luhn-valid SIREN); other
countries can be wired the same way as national providers get added.
`defaults.New()` wires the SIREN sub-validator automatically.

`ValidateChecksum` follows the same delegation pattern.

### SIREN / SIRET

Both validators enforce all-digit Luhn (mod-10). SIREN is 9 digits, SIRET
is 14 digits (9-digit SIREN + 5-digit NIC). The NIC `00000` is reserved by
INSEE and rejected at the format level.

**Dérogations**: La Poste (SIREN `356000000`) is fully supported — a La
Poste SIRET is accepted when either Luhn passes or the plain digit sum
is divisible by 5 (INSEE dérogatoire rule). Callers can register
additional non-Luhn rules for other historical SIRENs via
`siren.WithDerogation` / `siret.WithDerogation`.

### Extensibility (option pattern)

All four providers (`vat`, `siren`, `siret`, `euid`) accept functional
options at construction:

```go
v := vat.New(vat.WithLegacy())                  // accept IE +/* legacy chars
s := siret.New(siret.WithDerogation("999999999", myRule))  // custom rule
e := euid.New(euid.WithSubValidator(siren.New()))          // wire SIREN sub-validator
```

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
