<!--
SPDX-FileCopyrightText: 2025 Mass Labs

SPDX-License-Identifier: MIT
-->

# go-pgmmr

Postgres storage implementation for a Merkle Mountain Range (MMR).

Specifically, this is for the [datatrails-merklelog](https://github.com/datatrails/go-datatrails-merklelog) system, which follows [this](https://github.com/robinbryce/draft-bryce-cose-merkle-mountain-range-proofs) IETF internet draft.

We use [github.com/jackc/pgx/v5](https://github.com/jackc/pgx) as the postgres storage driver.

# Background / Concepts

Please refer to the [cheatsheets](https://github.com/datatrails/go-datatrails-merklelog/blob/main/mmr-math-cheatsheet.md) in the datatrails repo for a quick overview of the concepts. Especially the different [type of indexes](https://github.com/datatrails/go-datatrails-merklelog/blob/main/term-cheatsheet.md).

# Rationale

This is primarily used to serve proofs of inclusion for abitrary data.
We therefore store the leaf data as well as their hashes, such that we can retreive what is needed directly without re-computation.

The datatrails implementaion gives us all the utilities to implement the higher order tree, so this just supplies the storage plumbing and a bit of utility code.

# TODOs

**Caching**

We will build this as a read-through cache. Meaning, leafs will only be fetched once from the database and stored in memory until the tree is closed.

It remains to be seen if we should retreive `[0, n]` nodes from the database when accessing leaf `n` of the tree.
This might be highly advantages performance-wise but as a first/dumb iteration we are not over-optimizing this just yet.

It should also be noted that the drive for this implementation is to store a lot of small trees, not one big one.

**Table layout**

* Each tree uses an ID
* We store tree indicies as the append-only log

**TODOs**

* check out what this means: `// tests and KAT data corresponding to the MMRIVER draft`