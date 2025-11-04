# Lokalise glossary guard

![GitHub Release](https://img.shields.io/github/v/release/bodrovis/lokalise-glossary-guard)
![CI](https://github.com/bodrovis/lokalise-glossary-guard/actions/workflows/ci.yml/badge.svg)

**Lokalise glossary guard** (LGG) is a lightweight command-line tool designed to validate glossary CSV files before uploading them to Lokalise as glossaries.

It helps catch common formatting issues early (wrong separators, missing headers, or encoding problems) ensuring your glossary uploads go smoothly.

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
| 11 | **`ensure-no-duplicate-header-cells`** | Detects duplicate header names. |
| 12 | **`ensure-no-empty-term-values`** | Ensures that every `term` cell contains a non-empty value. |
| 13 | **`ensure-no-duplicate-term-values`** | Checks that `term` values are unique (case-sensitive). |
| 14 | **`ensure-no-orphan-locale-descriptions`** | Prevents `_description` columns without corresponding language columns. |
| 15 | **`ensure-no-invalid-flags`** | Validates flag columns (`casesensitive`, `translatable`, `forbidden`) contain only `yes`/`no` values. |

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

## License

BSD Clause 3