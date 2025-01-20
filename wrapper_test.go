// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package pgmmr_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/cryptix/go-pgmmr"
	mmr "github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/jackc/pgx/v5"
	"github.com/peterldowns/testy/assert"
)

func TestWrapper(t *testing.T) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("DATABASE_URL not set")
	}

	connPool, err := pgx.Connect(context.Background(), dbUrl)
	assert.Nil(t, err)

	hasher := sha256.New()

	const testId = 42 * 123

	t.Logf("testId: %d", testId)

	// delete previous values and tree nodes for clean slate
	_, err = connPool.Exec(context.Background(), "DELETE FROM pgmmr_values WHERE tree_id = $1", testId)
	assert.Nil(t, err)
	_, err = connPool.Exec(context.Background(), "DELETE FROM pgmmr_nodes WHERE tree_id = $1", testId)
	assert.Nil(t, err)

	tree, err := pgmmr.NewPostgresVerifierTree(connPool, hasher, testId)
	assert.Nil(t, err)

	// roll some random values and save their indices
	const mmrSize = 8
	// var val []byte
	type testValue struct {
		idx uint64
		val []byte
	}
	numLeafs := mmr.LeafCount(mmrSize)
	testValues := make([]testValue, numLeafs)
	// for i := 0; i < size; i++ {
	// 	val = make([]byte, 32)
	// 	rand.Read(val)
	// 	idx, err := tree.Add(val)
	// 	assert.Nil(t, err)
	// 	testValues[i] = testValue{idx: idx, val: val}
	// }

	for i := 0; i < int(numLeafs); i++ {
		input := []byte(fmt.Sprintf("hello %02d", i))
		idx, err := tree.Add(input)
		assert.Nil(t, err)
		t.Logf("idx: %02d: %s", idx, input)
		testValues[i] = testValue{idx: idx, val: input}
	}

	root, err := tree.Root()
	assert.Nil(t, err)
	t.Logf("root: %x", root)

	// ensure we can get the values back
	for _, tv := range testValues {
		t.Logf("getting value %d", tv.idx)
		val, err := tree.GetValue(tv.idx)
		assert.Nil(t, err)
		assert.Equal(t, tv.val, val)

		hasher.Reset()
		hasher.Write(tv.val)
		data := hasher.Sum(nil)
		node, err := tree.GetNode(tv.idx)
		assert.Nil(t, err)
		assert.Equal(t, data, node)
	}

	t.Log("values checked. now verifying proofs")

	// verify all the values
	for _, tv := range testValues {
		proof, err := tree.MakeProof(tv.idx)
		assert.Nil(t, err)
		assert.NotEqual(t, nil, proof)
		t.Logf("proof for %d created", tv.idx)
		assert.Nil(t, tree.VerifyProof(*proof))
	}
}
