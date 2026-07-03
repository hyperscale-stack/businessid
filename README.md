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
| EUID                          | `providers/euid`                | ✓      | Native EU-27 registers    | —        |
| EORI                          | `providers/eori`                | ✓      | —                     | —        |
| VAT                           | `providers/vat`                 | ✓      | 30 codes (EU-27 + XI, GB, NO, IS kennitala) | —        |
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
| Iceland              | IS   | 5, 6, or 10 (kennitala) digits      |
| Liechtenstein        | LI   | 5 digits                            |

Checksum validation (`ValidateChecksum`) is implemented for **30 codes**:
EU-27 + `XI`, `GB`, `NO`, and `IS` (kennitala 10-digit form only). Each
algorithm is sourced from the national tax authority documentation and
cross-verified against `python-stdnum` test vectors. `LI` remains
format-only (no published algorithm). See
[providers/vat/checksums.go](providers/vat/checksums.go) for per-country
source references. Coverage per country includes multi-variant support:
`BG` (BULSTAT 9d, EGN 10d, foreigner LNCh 10d), `CZ` (legal 8d + 9d + rodné
číslo 10d), `LV` (legal entity + natural person personal code).

**Aliases**: `GR` → `EL` and `UK` → `GB` are canonicalized. `IE` legacy
VAT numbers containing `+` or `*` in position 2 are accepted with the
`vat.WithLegacy()` option. `BE` pre-2005 9-digit VAT numbers are
automatically canonicalized to their post-2005 10-digit form.

**Reserved prefixes**: `CY12…` is rejected as it is reserved for legacy
Cypriot TINs, not VAT.

### EUID (meta-format, native validators for EU-27)

`ValidateFormat` enforces the BRIS layout
`<CC><REGISTER>.<REGISTRATION>` (Regulation (EU) 2015/884). The register
segment is 1–20 upper-case alphanumeric characters; the registration
segment is 1–64 characters from `[A-Z0-9./\- +]`.

**Native register validators for all EU-27**: no external wiring
required. `euid.New()` alone validates the REGISTRATION segment against
the national register's format (and checksum where documented):

| Country | Register | Format | Checksum |
|---------|----------|--------|:--------:|
| AT | Firmenbuchnummer (FN) | 1-6 digits + 1 letter | — |
| BE | KBO/BCE | 10 digits | ✓ (mod-97) |
| BG | EIK | 9 or 13 digits | ✓ (BULSTAT) |
| HR | MBS | 8 digits | — |
| CY | HE number | 6 digits | — |
| CZ | IČO | 8 digits | ✓ (mod-11) |
| DE | Handelsregister | free-form (court/type/number) | — |
| DK | CVR | 8 digits | ✓ (mod-11) |
| EE | Registrikood | 8 digits | ✓ (mod-11) |
| EL | GEMI | 12 digits | — |
| ES | NIF/CIF | alnum + 7 digits + alnum | ✓ (DNI/NIE/CIF) |
| FI | Y-tunnus | 8 digits | ✓ (mod-11) |
| FR | SIREN | 9 digits | ✓ (Luhn) |
| HU | Cégjegyzékszám | 10 digits (with optional dashes) | — |
| IE | CRO number | 5-7 digits | — |
| IT | Codice Fiscale entità | 11 digits | ✓ (Luhn) |
| LT | Juridinio asmens kodas | 9 digits | ✓ (mod-11) |
| LU | RCSL | B + 4-6 digits | — |
| LV | Reģistrācijas numurs | 11 digits | ✓ (mod-11) |
| MT | Company number | C + 4-6 digits | — |
| NL | KVK | 8 digits | — |
| PL | KRS | 10 digits | — |
| PT | NIPC | 9 digits | ✓ (mod-11) |
| RO | CUI | 2-10 digits | ✓ (mod-11) |
| SE | Organisationsnummer | 10 digits | ✓ (Luhn) |
| SI | Matična številka | 7 digits | — |
| SK | IČO | 8 digits | ✓ (mod-11) |

Non-BRIS prefixes (XI, GB, NO, IS, LI) are accepted for meta-format
shape only; use `euid.WithCountryValidator(cc, validator)` to inject a
custom validator for extension.

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
v := vat.New(vat.WithLegacy())                             // accept IE +/* legacy chars
s := siret.New(siret.WithDerogation("999999999", myRule))  // custom rule
e := euid.New(euid.WithCountryValidator("XI", myValidator)) // custom register validator
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
