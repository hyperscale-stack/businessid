# testdata/ — real-world business-identifier corpus

This directory holds a corpus of real national register numbers, VAT
numbers, SIRETs and SIRENs collected from public sources (national
registries, Wikipedia infoboxes, corporate legal notices, VIES). The
files are consumed by the `TestValidateRealWorld` tests under each
provider package and by
`TestNativeRegisterAlgorithmsRealWorld` under `providers/euid/`.

## Structure

```
testdata/
├── registers/<cc>-valide.txt   raw national register numbers (source of truth per country)
├── euid/<cc>-valide.txt        derived: "<CC><REGISTER>." + registers/<cc>
├── vat/<cc>-valide.txt         VAT numbers (hand-collected, not derived)
├── siret/fr-valide.txt         FR SIRETs (source of truth for FR)
└── siren/fr-valide.txt         derived: siret[:9], deduplicated
```

The derived files (`euid/*.txt`, `siren/fr-valide.txt`,
`registers/fr-valide.txt`) carry a `# GENERATED …` header and are
rewritten by `go run ./hack/derive-corpus`. Do not hand-edit them.

## Country coverage (initial pass, 2026-07-03)

| Country | registers | vat | siret | Notes |
|---------|:---------:|:---:|:-----:|-------|
| BE      | 42        | 25  | —     | BEL 20 + several mid-caps. Below the 50 target. |
| DE      | 100       | 25  | —     | DAX 40 + MDAX + Mittelstand. Umlauts expanded (Ü→UE, Ö→OE, Ä→AE, ß→SS) to fit the BRIS `[A-Z0-9./\- +]` charset. |
| ES      | 26        | 25  | —     | IBEX 35. Below the 50 target — collection halted early. |
| FR      | 113 SIREN | 30  | 113   | CAC 40 + SBF 120 + Next 40 + La Poste (SIREN 356000000, mod-5 dérogation). All from annuaire-entreprises.data.gouv.fr. |
| IT      | 100       | 25  | —     | FTSE MIB + FTSE Italia + private (Ferrero, Barilla, Lavazza, Ferragamo…). |
| NL      | 81        | 18  | —     | AEX + AMX + startups. VAT re-verified via btw-zoeken.nl after Wikipedia-sourced values were found to be fabricated (see "Data quality" below). |

Countries **not yet covered** (21): AT, BG, CY, CZ, DK, EE, EL, FI,
HR, HU, IE, LT, LU, LV, MT, PL, PT, RO, SE, SI, SK. To be added in a
later collection pass.

## Sources (per country)

- **BE** — kbopub.economie.fgov.be, staatsbladmonitor.be,
  companyweb.be, jaarrekening.be, corporate legal notices.
- **DE** — unternehmensregister.de, handelsregister.de, Impressum
  pages of DAX 40 / MDAX / SDAX / Mittelstand companies (mandatory
  under §5 TMG).
- **ES** — cnmv.es (regulator), labolsavirtual.com, iberinform.es,
  openmercantil.es, corporate Aviso Legal pages.
- **FR** — annuaire-entreprises.data.gouv.fr (INSEE base). Every FR
  entry cross-referenceable at
  `https://annuaire-entreprises.data.gouv.fr/entreprise/<SIREN>`.
- **IT** — ufficiocamerale.it, visura.pro, companyreports.it,
  Gazzetta Ufficiale, corporate Note Legali.
- **NL** — Wikipedia infobox (KVK), btw-zoeken.nl and vat-search.com
  (VAT), corporate colophon.

## Data quality

- Every ID was published by a source we deem reliable; nothing is
  fabricated by an LLM. Where a source could not be found for a
  candidate company, that company was dropped rather than invented.
- **We do NOT filter with the lib before commit.** The rule (from the
  project owner) is: collect from reliable sources → run the tests →
  investigate failures ex-post. A failing test is either a bug in the
  lib (fix it) or a bad transcription (remove the entry) — decided
  case by case.
- Two library issues were surfaced during the initial run and fixed
  in the same commit:
  1. **NL VAT (Wikipedia-sourced)** — 23 of 25 numbers failed the
     mod-11 check because the sub-agent had emitted Wikipedia
     infobox VAT values that turned out to be inconsistent with the
     Belastingdienst-published numbers. Replaced with values
     from btw-zoeken.nl / vat-search.com (17 entries). **No lib bug
     — data was wrong.**
  2. **ES CIF N-prefix** — `providers/vat/checksums.go` and
     `providers/euid/nationals.go` were treating N-prefix CIFs as
     "digit-check" entities. But N is reserved for foreign entities
     (ex: ArcelorMittal `N0181056C`) which the Spanish tax authority
     issues with a **letter** check digit. Fix: the "either" group
     (`C,D,F,G,J,L,M,N,U,V`) now accepts **both** digit-check and
     letter-check forms. Regression tests added in
     `providers/vat/vat_test.go` (`es-valid-cif-n-foreign`) and
     `providers/euid/euid_test.go` (`es-valid-cif-n-foreign`).

## How to reproduce

```
cd businessid
go run ./hack/derive-corpus   # regenerate euid/ + siren/ + registers/fr from sources
go test -run TestValidateRealWorld ./...
go test -run TestNativeRegisterAlgorithmsRealWorld ./providers/euid/...
make test                     # everything (existing + real-world)
```

## Adding a new country

1. Collect real IDs from national registries and legal notices.
2. Write `testdata/registers/<cc>-valide.txt` (canonical form, one
   per line, `#` comments allowed).
3. Write `testdata/vat/<cc>-valide.txt` (canonical VAT with prefix).
4. Run `go run ./hack/derive-corpus` to (re)generate the EUID file.
5. Run `go test ./...`. Investigate any failure per the rule above.
6. Document in this MANIFEST (country row + sources).
