/*
 * Copyright (c) 2023 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package uuid_test

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.arpabet.com/uuid"
	"math/rand"
	"testing"
	"time"
)

func TestSuit(t *testing.T) {

	println("Empty=", uuid.Empty.String())

	id := uuid.New(uuid.DCESecurityVer2)
	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.DCESecurityVer2, id.Version())

	// check Equal

	assert.False(t, uuid.Equal(&id, nil))
	assert.False(t, uuid.Equal(nil, &id))
	assert.True(t, uuid.Equal(nil, nil))
	assert.True(t, uuid.Equal(&id, &id))

	// check Versions

	testTimebasedUUID(t)

	testRandomlyGeneratedUUID(t)
	testNamebasedUUID(t)

	testTimebasedNamedUUID(t)

	testParser(t)

}

func testParser(t *testing.T) {

	id := uuid.New(uuid.TimebasedVer1)
	id.SetTime(time.Now())
	id.SetCounter(rand.Int63())

	comp, err := uuid.Parse(id.String())
	if err != nil {
		t.Fatal("parse failed ", id.String(), err)
	}

	assert.True(t, id.Equal(comp))

	str := id.String()

	// valid wrappers must be accepted
	for _, valid := range []string{
		str,
		"{" + str + "}",
		"\"" + str + "\"",
		"urn:uuid:" + str,
		"534b44a19bf13d20b71ecc4eb77c572f",
	} {
		if _, err := uuid.Parse(valid); err != nil {
			t.Fatalf("expected %q to parse, got %v", valid, err)
		}
	}

	// invalid input must be rejected with an error, not silently zeroed
	for _, invalid := range []string{
		"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",        // 32 chars, not hex
		"X534b44a1-9bf1-3d20-b71e-cc4eb77c572fX",  // wrong wrapper chars
		"{534b44a1-9bf1-3d20-b71e-cc4eb77c572f\"", // mismatched wrapper
		"534b44a1+9bf1-3d20-b71e-cc4eb77c572f",    // wrong separator
		"",                                        // empty
	} {
		if _, err := uuid.Parse(invalid); err == nil {
			t.Fatalf("expected %q to fail parsing, got no error", invalid)
		}
	}

}

func testTimebasedNamedUUID(t *testing.T) {

	id, err := uuid.NameUUIDFromBytes([]byte("content"), uuid.NamebasedVer5)
	if err != nil {
		t.Fatal("fail to create name id ", err)
	}

	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.NamebasedVer5, id.Version())
	assert.Equal(t, uint64(0x40f06fd77405247), id.MostSigBits)
	assert.Equal(t, uint64(0x8d450774f5ba30c5), id.LeastSigBits)

	id.SetUnixTimeMillis(0)
	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.TimebasedVer1, id.Version())
	assert.Equal(t, int64(0), id.UnixTimeMillis())
	assert.Equal(t, uint64(0x138140001dd211b2), id.MostSigBits)
	assert.Equal(t, uint64(0x8d450774f5ba30c5), id.LeastSigBits)

	assertMarshalText(t, id)
	assertMarshalJson(t, id)
	assertMarshalBinary(t, id)
	assertMarshalSortableBinary(t, id)

}

func testTimebasedUUID(t *testing.T) {

	id := uuid.New(uuid.TimebasedVer1)
	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.TimebasedVer1, id.Version())

	assert.Equal(t, int64(0), id.Time100Nanos())
	assert.Equal(t, 0, id.ClockSequence())
	assert.Equal(t, int64(0), id.Node())

	// test nodeMask
	id.SetNode(int64(0x0000FFFFFFFFFFFF))
	assert.Equal(t, int64(0x0000FFFFFFFFFFFF), id.Node())
	assert.Equal(t, uuid.IETF, id.Variant())

	// test clear
	id.SetNode(0)
	assert.Equal(t, int64(0), id.Node())

	// test OverflowNode
	id.SetNode(int64(0x0001FFFFFFFFFFFF))
	assert.Equal(t, int64(0x0000FFFFFFFFFFFF), id.Node())
	assert.Equal(t, uuid.IETF, id.Variant())

	// test clear Node
	id.SetClockSequence(int(0x3FFF))
	id.SetNode(0)
	assert.Equal(t, int64(0), id.Node())
	assert.Equal(t, uuid.IETF, id.Variant())
	id.SetClockSequence(int(0))

	// test OverflowClockSequence
	id.SetClockSequence(int(0x13FFF))
	assert.Equal(t, int(0x3FFF), id.ClockSequence())
	assert.Equal(t, uuid.IETF, id.Variant())
	id.SetClockSequence(0)

	// testMaxClockSequence
	id.SetClockSequence(int(0x3FFF))
	assert.Equal(t, int(0x3FFF), id.ClockSequence())
	assert.Equal(t, uuid.IETF, id.Variant())

	// test clear ClockSequence
	id.SetNode(int64(0x0000FFFFFFFFFFFF))
	id.SetClockSequence(int(0))
	assert.Equal(t, int64(0x0000FFFFFFFFFFFF), id.Node())
	assert.Equal(t, uuid.IETF, id.Variant())
	id.SetNode(int64(0))

	// test maxTimeBits
	id.SetTime100Nanos(int64(0x0FFFFFFFFFFFFFFF))
	assert.Equal(t, int64(0x0FFFFFFFFFFFFFFF), id.Time100Nanos())
	assert.Equal(t, uuid.TimebasedVer1, id.Version())

	// test clear maxTimeBits
	id.SetTime100Nanos(0)
	assert.Equal(t, int64(0), id.Time100Nanos())
	assert.Equal(t, uuid.TimebasedVer1, id.Version())

	// test Milliseconds
	id.SetUnixTimeMillis(1)
	assert.Equal(t, int64(1), id.UnixTimeMillis())

	// test Negative Milliseconds
	id.SetUnixTimeMillis(-1)
	assert.Equal(t, int64(-1), id.UnixTimeMillis())

	// clear
	id.SetUnixTimeMillis(0)
	assert.Equal(t, int64(0), id.UnixTimeMillis())

	// test Counter

	id = uuid.New(uuid.TimebasedVer1)

	id.SetMinCounter()
	fmt.Print("min=", id.String(), "\n")
	fmt.Printf("counter=%x\n", id.Counter())
	binMin, _ := id.MarshalSortableBinary()

	id.SetMaxCounter()
	fmt.Print("max=", id.String(), "\n")
	fmt.Printf("counter=%x\n", id.Counter())
	binMax, _ := id.MarshalSortableBinary()

	for i := 1; i != 100; i = i + 1 {

		anyNumber := int64(i)
		id.SetCounter(anyNumber)

		binLesser, _ := id.MarshalSortableBinary()
		id.SetCounter(anyNumber + 1)

		binGreater, _ := id.MarshalSortableBinary()

		assert.True(t, bytes.Compare(binMin, binLesser) < 0, "min failed")
		assert.True(t, bytes.Compare(binLesser, binGreater) < 0, "seq failed")
		assert.True(t, bytes.Compare(binGreater, binMax) < 0, "max failed")
	}

	id = uuid.New(uuid.TimebasedVer1)

	current := time.Now()

	id.SetTime(current)
	cnt := id.SetCounter(rand.Int63())

	assert.Equal(t, current.UnixNano()/100, id.Time().UnixNano()/100)
	assert.Equal(t, cnt, id.Counter())

	assertMarshalText(t, id)
	assertMarshalJson(t, id)
	assertMarshalBinary(t, id)
	assertMarshalSortableBinary(t, id)

}

func testRandomlyGeneratedUUID(t *testing.T) {

	id := uuid.New(uuid.RandomlyGeneratedVer4)
	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.RandomlyGeneratedVer4, id.Version())

	id, err := uuid.RandomUUID()

	if err != nil {
		t.Fatal("fail to create random id ", err)
	}

	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.RandomlyGeneratedVer4, id.Version())

	assertMarshalText(t, id)
	assertMarshalJson(t, id)
	assertMarshalBinary(t, id)

}

func testNamebasedUUID(t *testing.T) {

	id := uuid.New(uuid.NamebasedVer5)
	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.NamebasedVer5, id.Version())

	id = uuid.New(uuid.NamebasedVer3)
	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.NamebasedVer3, id.Version())

	id, err := uuid.NameUUIDFromBytes([]byte("alex"), uuid.NamebasedVer3)

	if err != nil {
		t.Fatal("fail to create random id ", err)
	}

	assert.Equal(t, uuid.IETF, id.Variant())
	assert.Equal(t, uuid.NamebasedVer3, id.Version())
	assert.Equal(t, uint64(0x534b44a19bf13d20), id.MostSigBits)
	assert.Equal(t, uint64(0xb71ecc4eb77c572f), id.LeastSigBits)

	assert.Equal(t, "534b44a1-9bf1-3d20-b71e-cc4eb77c572f", id.String())

	assertMarshalText(t, id)
	assertMarshalJson(t, id)
	assertMarshalBinary(t, id)

}

func assertMarshalText(t *testing.T, id uuid.UUID) {

	var actual uuid.UUID
	data, err := id.MarshalText()

	if err != nil {
		t.Fatal("fail to MarshalText ", err)
	}

	err = actual.UnmarshalText(data)

	if err != nil {
		t.Fatal("fail to MarshalText ", err)
	}

	assert.Equal(t, id.MostSigBits, actual.MostSigBits)
	assert.Equal(t, id.LeastSigBits, actual.LeastSigBits)

}

func assertMarshalJson(t *testing.T, id uuid.UUID) {

	var actual uuid.UUID
	data, err := id.MarshalJSON()

	if err != nil {
		t.Fatal("fail to MarshalJson ", err)
	}

	err = actual.UnmarshalJSON(data)

	if err != nil {
		t.Fatal("fail to UnmarshalJson ", err)
	}

	assert.Equal(t, id.MostSigBits, actual.MostSigBits)
	assert.Equal(t, id.LeastSigBits, actual.LeastSigBits)

}
func assertMarshalBinary(t *testing.T, id uuid.UUID) {

	var actual uuid.UUID
	data, err := id.MarshalBinary()

	if err != nil {
		t.Fatal("fail to MarshalBinary ", err)
	}

	err = actual.UnmarshalBinary(data)

	if err != nil {
		t.Fatal("fail to UnmarshalBinary ", err)
	}

	assert.Equal(t, id.MostSigBits, actual.MostSigBits)
	assert.Equal(t, id.LeastSigBits, actual.LeastSigBits)

}

func assertMarshalSortableBinary(t *testing.T, id uuid.UUID) {

	var actual uuid.UUID
	data, err := id.MarshalSortableBinary()

	if err != nil {
		t.Fatal("fail to MarshalSortableBinary ", err)
	}

	err = actual.UnmarshalSortableBinary(data)

	if err != nil {
		t.Fatal("fail to UnmarshalSortableBinary ", err)
	}

	assert.Equal(t, id.MostSigBits, actual.MostSigBits)
	assert.Equal(t, id.LeastSigBits, actual.LeastSigBits)

}
