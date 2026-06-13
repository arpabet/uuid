/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package uuid_test

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"go.arpabet.com/uuid"
	"testing"
	"time"
)

func TestNewV1(t *testing.T) {

	a, err := uuid.NewV1()
	assert.NoError(t, err)
	assert.Equal(t, uuid.TimebasedVer1, a.Version())
	assert.Equal(t, uuid.IETF, a.Variant())

	// node must have the multicast bit set (random node, RFC 4122 §4.5)
	assert.Equal(t, int64(0x010000000000), a.Node()&int64(0x010000000000))

	// generated values must be strictly monotonic even in a tight loop
	prev, err := a.MarshalSortableBinary()
	assert.NoError(t, err)
	for i := 0; i < 1000; i++ {
		next, err := uuid.NewV1()
		assert.NoError(t, err)
		assert.Equal(t, uuid.TimebasedVer1, next.Version())
		bin, err := next.MarshalSortableBinary()
		assert.NoError(t, err)
		assert.True(t, bytes.Compare(prev, bin) < 0, "v1 not monotonic at %d", i)
		prev = bin
	}
}

func TestNewV7(t *testing.T) {

	before := time.Now().UnixMilli()
	id, err := uuid.NewV7()
	after := time.Now().UnixMilli()
	assert.NoError(t, err)

	assert.Equal(t, uuid.UnixTimeVer7, id.Version())
	assert.Equal(t, uuid.IETF, id.Variant())

	ms := id.UnixTimeMillisV7()
	assert.True(t, ms >= before && ms <= after, "v7 timestamp out of range")
	assert.Equal(t, ms, id.TimeV7().UnixMilli())

	// survives a text round trip
	parsed, err := uuid.Parse(id.String())
	assert.NoError(t, err)
	assert.True(t, id.Equal(parsed))
	assert.Equal(t, uuid.UnixTimeVer7, parsed.Version())
}

func TestV7Sortable(t *testing.T) {

	var prev []byte
	for i := 0; i < 50; i++ {
		id, err := uuid.NewV7()
		assert.NoError(t, err)
		bin, err := id.MarshalBinary()
		assert.NoError(t, err)
		if prev != nil {
			// monotonic per millisecond; ensure non-decreasing ordering
			assert.True(t, bytes.Compare(prev, bin) <= 0, "v7 binary not sortable at %d", i)
		}
		prev = bin
		time.Sleep(time.Millisecond)
	}
}

func TestVersionString(t *testing.T) {
	assert.Equal(t, "ReorderedTimeVer6", uuid.ReorderedTimeVer6.String())
	assert.Equal(t, "UnixTimeVer7", uuid.UnixTimeVer7.String())
}

func TestSQLValueScan(t *testing.T) {

	id, err := uuid.NewV7()
	assert.NoError(t, err)

	val, err := id.Value()
	assert.NoError(t, err)
	assert.Equal(t, id.String(), val)

	// scan from string
	var fromString uuid.UUID
	assert.NoError(t, fromString.Scan(id.String()))
	assert.True(t, id.Equal(fromString))

	// scan from textual bytes
	var fromText uuid.UUID
	assert.NoError(t, fromText.Scan([]byte(id.String())))
	assert.True(t, id.Equal(fromText))

	// scan from raw 16 byte binary
	bin, err := id.MarshalBinary()
	assert.NoError(t, err)
	var fromBinary uuid.UUID
	assert.NoError(t, fromBinary.Scan(bin))
	assert.True(t, id.Equal(fromBinary))

	// nil keeps the zero value
	var fromNil uuid.UUID
	assert.NoError(t, fromNil.Scan(nil))
	assert.True(t, uuid.Empty.Equal(fromNil))

	// unsupported type fails
	var bad uuid.UUID
	assert.Error(t, bad.Scan(12345))

	// invalid string fails
	var badStr uuid.UUID
	assert.Error(t, badStr.Scan("not-a-uuid"))
}
