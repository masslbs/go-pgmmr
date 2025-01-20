// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package pgmmr

import (
	"context"
	"errors"
	"fmt"
	"hash"

	"github.com/datatrails/go-datatrails-merklelog/mmr"
	pgx "github.com/jackc/pgx/v5"
)

// VerifierTree is a tree that can be used to verify the inclusion of values in the tree.
type VerifierTree interface {
	// Add adds a value to the tree and returns the index of the value.
	Add(value []byte) (uint64, error)
	// GetValue returns the value at the given index.
	GetValue(i uint64) ([]byte, error)
	// GetNode returns the node at the given index.
	GetNode(i uint64) ([]byte, error)

	// Root returns the root of the tree.
	Root() ([]byte, error)
	// MakeProof returns a proof that the value at the given index is in the tree.
	MakeProof(i uint64) (*Proof, error)
	// VerifyProof verifies a proof that the value at the given index is in the tree.
	VerifyProof(proof Proof) error
}

type Proof struct {
	Root      []byte
	NodeIndex uint64
	Path      [][]byte
}

func NewPostgresVerifierTree(db *pgx.Conn, hasher hash.Hash, id uint64) (VerifierTree, error) {
	nodes, err := NewPostgresNodeStore(db, id)
	if err != nil {
		return nil, err
	}
	return &PostgresVerifierTree{
		db:     db,
		hasher: hasher,
		treeId: id,
		nodes:  nodes,
	}, nil
}

type PostgresVerifierTree struct {
	db     *pgx.Conn
	hasher hash.Hash
	treeId uint64
	nodes  *PostgresNodeStore
}

func (t *PostgresVerifierTree) Add(value []byte) (uint64, error) {
	hasher := t.hasher
	hasher.Reset()
	hasher.Write(value)
	data := hasher.Sum(nil)

	newSize, err := mmr.AddHashedLeaf(t.nodes, t.hasher, data)
	if err != nil {
		return 0, err
	}
	// AddHashedLeaf returns the new size of the tree
	// which is equal to the last node _position_ in the tree
	leafIdx := mmr.LeafIndex(newSize - 1)

	const insertValueQry = "INSERT INTO pgmmr_values (tree_id, leaf_idx, data) VALUES ($1, $2, $3)"
	_, err = t.db.Exec(context.Background(), insertValueQry, t.treeId, leafIdx, value)
	if err != nil {
		return 0, err
	}

	return leafIdx, nil
}

func (t *PostgresVerifierTree) GetNode(i uint64) ([]byte, error) {
	// turn leaf index into node index
	nodeIdx := mmr.MMRIndex(i)
	return t.nodes.Get(nodeIdx)
}

func (t *PostgresVerifierTree) GetValue(i uint64) ([]byte, error) {
	var value []byte
	err := t.db.QueryRow(context.Background(), "SELECT data FROM pgmmr_values WHERE tree_id = $1 AND leaf_idx = $2", t.treeId, i).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to get value %d: %w", i, mmr.ErrNotFound)
		}
		return nil, err
	}
	return value, nil
}

func (t *PostgresVerifierTree) Root() ([]byte, error) {
	count, err := t.nodeCount()
	if err != nil {
		return nil, err
	}
	root, err := mmr.GetRoot(count, t.nodes, t.hasher)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func (t *PostgresVerifierTree) MakeProof(i uint64) (*Proof, error) {
	count, err := t.nodeCount()
	if err != nil {
		return nil, err
	}
	mmrIndex := mmr.MMRIndex(i)
	proof, err := mmr.InclusionProofBagged(count, t.nodes, t.hasher, mmrIndex)
	if err != nil {
		return nil, err
	}
	root, err := t.Root()
	if err != nil {
		return nil, err
	}
	return &Proof{
		Root:      root,
		NodeIndex: mmrIndex,
		Path:      proof,
	}, nil
}

func (t *PostgresVerifierTree) VerifyProof(proof Proof) error {
	count, err := t.nodeCount()
	if err != nil {
		return err
	}
	node, err := t.nodes.Get(proof.NodeIndex)
	if err != nil {
		return err
	}
	if !mmr.VerifyInclusionBagged(count, t.hasher, node, proof.NodeIndex, proof.Path, proof.Root) {
		return fmt.Errorf("proof verification for %d failed: %w", proof.NodeIndex, mmr.ErrVerifyInclusionFailed)
	}
	return nil
}

func (t *PostgresVerifierTree) nodeCount() (uint64, error) {
	var count uint64
	const qry = "SELECT count(*) FROM pgmmr_nodes WHERE tree_id = $1"
	err := t.db.QueryRow(context.Background(), qry, t.treeId).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
