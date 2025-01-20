# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

.phony: all lint reuse

lint:
	go vet ./...
	revive -formatter friendly
	reuse lint

LIC := MIT
CPY := "Mass Labs"

reuse:
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights Makefile README.md *.go test_schema.sql go.mod .gitignore
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights --force-dot-license go.sum flake.lock
