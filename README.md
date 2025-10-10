# Lokalise glossary guard

![GitHub Release](https://img.shields.io/github/v/release/bodrovis/lokalise-glossary-guard)
![CI](https://github.com/bodrovis/lokalise-glossary-guard/actions/workflows/ci.yml/badge.svg)


**Lokalise Glossary Guard** (LGG) is a lightweight command-line tool designed to validate glossary CSV files before uploading them to Lokalise as glossaries.

It helps catch common formatting issues early (wrong separators, missing headers, or encoding problems) ensuring your glossary uploads go smoothly.

## Usage

```
glossary-guard validate --file path/to/glossary.csv
```