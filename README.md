# Lokalise glossary guard

![GitHub Release](https://img.shields.io/github/v/release/bodrovis/lokalise-glossary-guard)
![CI](https://github.com/bodrovis/lokalise-glossary-guard/actions/workflows/ci.yml/badge.svg)

**Lokalise glossary guard** (LGG) is a lightweight command-line tool designed to validate glossary CSV files before uploading them to Lokalise as glossaries.

It helps catch common formatting issues early (wrong separators, missing headers, or encoding problems) ensuring your glossary uploads go smoothly.

## Usage

```
# Validate a single file
glossary-guard validate --file glossary.csv

# Validate multiple files
glossary-guard validate -f glossary1.csv -f glossary2.csv

# Validate with explicit language codes
glossary-guard validate -f glossary.csv -l en -l de_DE -l fr
```

Example output:

```
────────────────────────────────────────────────────────────────────────
Validating: samples\glossary.csv
────────────────────────────────────────────────────────────────────────

→ [CRIT] ensure-csv-extension ... PASS
   File extension OK: .csv
→ [CRIT] ensure-utf8-encoding ... PASS
   File encoding is valid UTF-8
→ [CRIT] ensure-header-and-rows ... PASS
   Header valid; required columns present; ';' delimiter confirmed; data parsed successfully
→ [NORM] ensure-known-optional-headers-with-langs ... PASS
   Optional header columns are valid for declared languages
→ [NORM] ensure-no-orphan-lang-descriptions ... PASS
   No orphan *_description columns found
→ [NORM] ensure-non-empty-term ... PASS
   All 'term' values are non-empty
→ [NORM] ensure-unique-headers ... PASS
   All header names are unique
→ [NORM] ensure-unique-terms ... PASS
   All terms are unique (case-sensitive)
→ [NORM] ensure-yn-flags ... PASS
   Y/N flag columns valid (yes/no only)

Summary for samples\glossary.csv: 9 passed, 0 failed, 0 errors
Result: PASSED
────────────────────────────────────────────────────────────────────────
```

## Available checks

Each CSV file is validated through a set of **critical** and **normal** checks:

| Category | Check | Purpose |
|-----------|--------|----------|
| **Critical** | `ensure-csv-extension` | File has `.csv` extension |
|  | `ensure-utf8-encoding` | File is valid UTF-8 |
|  | `ensure-header-and-rows` | Header is present, delimited by `;`, includes required columns (`term;description`), and data rows exist |
| **Normal** | `ensure-non-empty-term` | Every `term` cell must be filled |
|  | `ensure-known-optional-headers-with-langs` | Only known headers and declared languages are allowed |
|  | `ensure-no-orphan-lang-descriptions` | No `_description` columns without matching language columns |
|  | `ensure-unique-headers` | Header names must be unique |
|  | `ensure-unique-terms` | `term` values must be unique (case-sensitive) |
|  | `ensure-yn-flags` | Certain flag columns (`casesensitive`, `translatable`, `forbidden`) can only contain `yes`/`no` |

## Guidelines for creating glossary CSV files

[As the official Lokalise documentation explains](https://docs.lokalise.com/en/articles/1400629-glossary#h_569a1424cc), when preparing a glossary CSV file for upload, you should follow these rules to avoid import errors.

### General formatting rules

- **Separators** — Always use **semicolons (`;`)** as column separators. Other separators (like commas or tabs) are **not supported** and will cause the upload to fail.
  + *This is the single most common issue when working with glossary CSVs.*
- **Header Row** — The file **must include a header row** describing the columns.
- **Encoding** — The file **must be encoded in UTF-8**. Using other encodings (e.g., Windows-1251, ISO-8859-1) will result in corrupted text or validation errors.

### Column structure

The recommended column order and meaning are:

| Column | Description |
|---------|--------------|
| **term** | The glossary term you want to add. |
| **description** | A general explanation of the term. |
| **casesensitive** | Either `yes` or `no`. Marks whether the term is case-sensitive. |
| **translatable** | Either `yes` or `no`. Indicates whether the term should be translated. |
| **forbidden** | Either `yes` or `no`. Marks terms that should *not* be used. |
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
- Empty rows are not allowed.
- Validation is case-insensitive for headers, but term values are checked case-sensitively for duplicates.

## License

BSD Clause 3