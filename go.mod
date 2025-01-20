// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

module github.com/cryptix/go-pgmmr

go 1.23.4

require (
	github.com/datatrails/go-datatrails-merklelog/mmr v0.1.1
	github.com/jackc/pgx/v5 v5.7.2
	github.com/peterldowns/testy v0.0.3
)

require (
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/exp v0.0.0-20230519143937-03e91628a987 // indirect
	golang.org/x/text v0.21.0 // indirect
)

// replace github.com/datatrails/go-datatrails-merklelog/mmr => ../go-datatrails-merklelog/mmr
