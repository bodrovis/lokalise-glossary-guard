# Lokalise glossary guard

![GitHub Release](https://img.shields.io/github/v/release/bodrovis/lokalise-glossary-guard)
![CI](https://github.com/bodrovis/lokalise-glossary-guard/actions/workflows/ci.yml/badge.svg)
[![Code Coverage](https://qlty.sh/gh/bodrovis/projects/lokalise-glossary-guard/coverage.svg)](https://qlty.sh/gh/bodrovis/projects/lokalise-glossary-guard)
[![Maintainability](https://qlty.sh/gh/bodrovis/projects/lokalise-glossary-guard/maintainability.svg)](https://qlty.sh/gh/bodrovis/projects/lokalise-glossary-guard)

**Lokalise glossary guard** (LGG) is a lightweight command-line tool designed to validate and fix CSV files before uploading them to Lokalise as glossaries.

It helps catch common formatting issues early (wrong separators, missing headers, or encoding problems) ensuring your glossary uploads go smoothly.

> Find web version at https://lokalise-glossary-guard.web.app/.

## Usage

```
# Validate a single file
lokalise-glossary-guard validate --file glossary.csv

# Validate multiple files and attempt fixes
lokalise-glossary-guard validate -f glossary1.csv -f glossary2.csv --fix

# (recommended) Validate with explicit language codes, attempt fixes, revalidate after fix
lokalise-glossary-guard validate -f samples/*.csv -l en -l de_DE -l fr --fix --rerun-after-fix
```

Example output:

```
────────────────────────────────────────────────────────────────────────
Validating: samples\glossary.csv
────────────────────────────────────────────────────────────────────────

Mode: FixMode=2, RerunAfterFix=true, HardFailOnErr=false

→ [CRIT] ensure-valid-extension ... PASS
   file extension OK: .csv
→ [CRIT] ensure-utf8-encoding ... PASS
   file encoding is valid UTF-8
→ [NORM] ensure-no-empty-lines ... PASS [changed]
   empty lines removed | note: removed 1 empty line
→ [CRIT] ensure-not-empty ... PASS
   file has content
→ [CRIT] ensure-at-least-two-lines ... PASS
   file has at least two lines (header + data)
→ [CRIT] ensure-semicolon-separators ... PASS
   file uses semicolons as separators
→ [CRIT] no-spaces-in-header ... PASS
   header columns are trimmed (no leading/trailing spaces)
→ [CRIT] ensure-lowercase-header ... PASS
   header service columns are already lowercase
→ [CRIT] ensure-term-description-header ... PASS
   header starts with term;description
→ [NORM] ensure-allowed-columns-header ... PASS [changed]
   header columns normalized (unknown columns removed, missing language columns added) | note: removed unknown columns and ensured declared languages are present
→ [NORM] warn-duplicate-header-cells ... PASS
   no duplicate header columns
→ [CRIT] no-empty-term-values ... PASS
   all rows have non-empty term
→ [NORM] warn-duplicate-term-values ... PASS [changed]
   removed duplicate term rows | note: removed duplicate term rows for: "session" (rows 118); "VAT" (rows 245); "card" (rows 13)
→ [NORM] warn-orphan-locale-descriptions ... PASS
   no orphan *_description columns
→ [CRIT] no-invalid-flags ... PASS
   all flag columns contain only yes/no

Summary for samples\glossary_complex.csv: 15 passed, 0 warning(s), 0 failed, 0 errors
Info wrote fixed file: samples\glossary_complex_fixed.csv (bytes=54876)
Result: PASSED
────────────────────────────────────────────────────────────────────────
```

## Available checks

Each glossary CSV file is validated sequentially through the following checks:

| № | Check Name | Purpose |
|--:|-------------|----------|
| 1 | **`ensure-valid-extension`** | Ensures the file has the `.csv` extension; renames automatically if needed. |
| 2 | **`ensure-valid-encoding`** | Verifies that the file is valid UTF-8. |
| 3 | **`ensure-no-empty-lines`** | Checks that there are no completely empty lines in the file. |
| 4 | **`ensure-non-empty-file`** | Confirms the file isn't empty. |
| 5 | **`ensure-at-least-two-lines`** | Requires at least one header line and one data line. |
| 6 | **`ensure-semicolon-separators`** | Validates that columns are separated by semicolons (`;`), not commas or tabs. |
| 7 | **`ensure-no-header-spaces`** | Checks that known header cell names don't contain spaces. |
| 8 | **`ensure-lowercase-header`** | Ensures all known header names are lowercase (except locale-related ones). |
| 9 | **`ensure-term-description-header`** | Validates that the header includes the required `term` and `description` columns. |
| 10 | **`ensure-allowed-columns-header`** | Allows only known headers. |
| 11 | **`warn-duplicate-header-cells`** | Detects duplicate header names. |
| 12 | **`no-empty-term-values`** | Ensures that every `term` cell contains a non-empty value. |
| 13 | **`warn-duplicate-term-values`** | Checks that `term` values are unique (case-sensitive). |
| 14 | **`warn-orphan-locale-descriptions`** | Prevents `_description` columns without corresponding language columns. |
| 15 | **`no-invalid-flags`** | Validates flag columns (`casesensitive`, `translatable`, `forbidden`) contain only `yes`/`no` values. |
| 16 | **`no-forbidden-non-translatable-terms`** | Checks there are no terms that are both non-translatable and forbidden. |

## Guidelines for creating glossary CSV files

[As the official Lokalise documentation explains](https://docs.lokalise.com/en/articles/1400629-glossary#h_569a1424cc), when preparing a glossary CSV file for upload, you should follow these rules to avoid import errors.

### General formatting rules

- **Separators** — Always use **semicolons (`;`)** as column separators. Other separators (like commas or tabs) are **not supported** and will cause the upload to fail.
  + *This is the single most common issue when working with glossary CSVs.*
- **Header row** — The file **must include a header row** describing the columns.
- **Encoding** — The file **must be encoded in UTF-8**. Using other encodings (e.g., Windows-1251, ISO-8859-1) will result in corrupted text or validation errors.

### Column structure

The recommended column order and meaning are:

| Column | Description |
|---------|--------------|
| **term** | The glossary term you want to add. |
| **description** | A general explanation of the term. |
| **casesensitive** | Either `yes` or `no`. Marks whether the term is case-sensitive. |
| **translatable** | Either `yes` or `no`. Indicates whether the term should be translated. |
| **forbidden** | Either `yes` or `no`. Marks terms that should not be used. |
| **tags** | A comma-separated list of tags (optional). |
| **Language ISO code columns** | Each column should be named after the **language ISO code** used in your Lokalise project (e.g., `en`, `de_DE`, `fr`). These contain translations or can be left empty. |
| **Language description columns** | Columns named like `<lang>_description` (e.g., `fr_description`, `de_description`). These contain language-specific term descriptions or can be left empty. |

**Example header:**

```csv
term;description;casesensitive;translatable;forbidden;tags;en;en_description;de_DE;de_DE_description
```

### Notes

- Columns after `term;description` are optional, but if present, they must follow the naming conventions above.
- Each column name should be unique.
- Empty rows are not recommended.
- Validation is case-insensitive for headers, but term values are checked case-sensitively for duplicates.

## Testing

Run:

```
go test -shuffle=on -count=1 -race ./...
```

## Release verification

Each release includes additional verification artifacts:

* `*_checksums.txt` — SHA-256 checksums for all release files
* `*_checksums.txt.sigstore.json` — keyless Cosign signature bundle for the checksum file
* `*.sbom.json` — SBOM files generated for the release archives

GitHub Artifact Attestations are also published for release files and can be viewed on the [repository attestations page](https://github.com/bodrovis/lokalise-glossary-guard/attestations).

To verify a release, install [Cosign](https://github.com/sigstore/cosign) and the [GitHub CLI](https://cli.github.com/).

Set the release tag you want to verify:

```bash
TAG="v1.0.0"
VERSION="${TAG#v}"
REPO="bodrovis/lokalise-glossary-guard"
```

Download all release assets into a new directory:

```bash
mkdir -p "lokalise-glossary-guard-${TAG}"
cd "lokalise-glossary-guard-${TAG}"

gh release download "$TAG" \
  --repo "$REPO"
```

Locate the checksum manifest and its Cosign bundle:

```bash
CHECKSUMS_FILE="lokalise-glossary-guard_${VERSION}_checksums.txt"
SIGNATURE_FILE="${CHECKSUMS_FILE}.sigstore.json"
```

First, verify that the checksum manifest was signed by the expected GitHub Actions workflow:

```bash
cosign verify-blob \
  --bundle "$SIGNATURE_FILE" \
  --certificate-identity "https://github.com/${REPO}/.github/workflows/goreleaser.yml@refs/tags/${TAG}" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "$CHECKSUMS_FILE"
```

Then verify that the downloaded release files match the signed checksums:

```bash
sha256sum --check "$CHECKSUMS_FILE"
```

Finally, verify the GitHub Artifact Attestation for a release archive. For example, for the Linux x86-64 archive:

```bash
gh attestation verify \
  "lokalise-glossary-guard_${VERSION}_Linux_x86_64.tar.gz" \
  --repo "$REPO" \
  --signer-workflow "$REPO/.github/workflows/goreleaser.yml"
```

All three commands must complete successfully:

1. `cosign verify-blob` confirms that the checksum manifest was signed by the expected tagged GitHub Actions workflow.
2. `sha256sum` confirms that the downloaded release files have not been modified.
3. `gh attestation verify` confirms the provenance and integrity of the selected release artifact.

## License

(c) [Elijah S. Krukowski](https://bodrovis.tech), BSD-3-Clause license