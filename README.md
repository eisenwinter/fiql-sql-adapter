<a name="readme-top"></a>

# fiql-sql-adapter

[![Go Report Card](https://goreportcard.com/badge/github.com/eisenwinter/fiql-sql-adapter)](https://goreportcard.com/report/github.com/eisenwinter/fiql-sql-adapter) [![Go](https://github.com/eisenwinter/fiql-sql-adapter/actions/workflows/go.yml/badge.svg)](https://github.com/eisenwinter/fiql-sql-adapter/actions/workflows/go.yml) [![GoDoc reference example](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/eisenwinter/fiql-sql-adapter)  [![Project Status: WIP - Initial development is in progress, but there has not yet been a stable, usable release suitable for the public.](https://www.repostatus.org/badges/latest/wip.svg)](https://www.repostatus.org/#wip)

golang fiql2sql adapter - early work in progress

## Getting Started

Do not use this as of now, it's still being actively developed.

If you wanna peek at the project feel free to do so, but expect it to break.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Why

`fiql-sql-adapter` exists to have a common and in `GET` requests useable query possibility. Its main goal is to free
you from writing custom search queries for classic CRUD applications.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


## Featurelist

This list represents the features I would like it to have once its done.

- [x] safe FIQL parsing
- [x] defining a fiql-to-table mapping manually 
- [x] defining a fiql-to-table mapping by struct tags
- [ ] sophisticated type checks (as of now its rather crude with minimal type support)
- [ ] value converters to convert fiql supplied arguments to the corresponding sql parameter
- [ ] join and computed column handling - this is tricky and needs some more thought [^1]
- [ ] [MAYBE] read struct tags from sqlx https://jmoiron.github.io/sqlx/ but I am not sure if that's a good idea


[^1]: Right now this could be archived with a view but that's not what I am after.

## License

Distributed under the BSD-2-Clause license. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>