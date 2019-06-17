/*
 *
 * Copyright 2018-present Alex Shvid.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package uuid

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

/**
UUID implementation for Golang Tests

@author Alex Shvid
*/

func TestSuit(t *testing.T) {

	println("Empty=", Empty.String())

	uuid := NewUUID(DCESecurityVer2)
	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, DCESecurityVer2, uuid.Version())

	// check Equal

	assert.False(t, Equal(&uuid, nil))
	assert.False(t, Equal(nil, &uuid))
	assert.True(t, Equal(nil, nil))
	assert.True(t, Equal(&uuid, &uuid))

	// check Versions

	testTimebasedUUID(t)

	testRandomlyGeneratedUUID(t)
	testNamebasedUUID(t)

	testTimebasedNamedUUID(t)

	testParser(t)

}

func testParser(t *testing.T) {

	uuid := NewUUID(TimebasedVer1)
	uuid.SetTime(time.Now())
	uuid.SetCounter(rand.Int63())

	comp, err := Parse(uuid.String())
	if err != nil {
		t.Fatal("parse failed ", uuid.String(), err)
	}

	assert.True(t, uuid.Equal(comp))

}

func testTimebasedNamedUUID(t *testing.T) {

	uuid, err := NameUUIDFromBytes([]byte("content"), NamebasedVer5)
	if err != nil {
		t.Fatal("fail to create name uuid ", err)
	}

	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, NamebasedVer5, uuid.Version())
	assert.Equal(t, uint64(0x40f06fd77405247), uuid.MostSigBits)
	assert.Equal(t, uint64(0x8d450774f5ba30c5), uuid.LeastSigBits)

	uuid.SetUnixTimeMillis(0)
	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, TimebasedVer1, uuid.Version())
	assert.Equal(t, int64(0), uuid.UnixTimeMillis())
	assert.Equal(t, uint64(0x138140001dd211b2), uuid.MostSigBits)
	assert.Equal(t, uint64(0x8d450774f5ba30c5), uuid.LeastSigBits)

	assertMarshalText(t, uuid)
	assertMarshalJson(t, uuid)
	assertMarshalBinary(t, uuid)
	assertMarshalSortableBinary(t, uuid)

}

func testTimebasedUUID(t *testing.T) {

	uuid := NewUUID(TimebasedVer1)
	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, TimebasedVer1, uuid.Version())

	assert.Equal(t, int64(0), uuid.Time100Nanos())
	assert.Equal(t, 0, uuid.ClockSequence())
	assert.Equal(t, int64(0), uuid.Node())

	// test nodeMask
	uuid.SetNode(int64(0x0000FFFFFFFFFFFF))
	assert.Equal(t, int64(0x0000FFFFFFFFFFFF), uuid.Node())
	assert.Equal(t, IETF, uuid.Variant())

	// test clear
	uuid.SetNode(0)
	assert.Equal(t, int64(0), uuid.Node())

	// test OverflowNode
	uuid.SetNode(int64(0x0001FFFFFFFFFFFF))
	assert.Equal(t, int64(0x0000FFFFFFFFFFFF), uuid.Node())
	assert.Equal(t, IETF, uuid.Variant())

	// test clear Node
	uuid.SetClockSequence(int(0x3FFF))
	uuid.SetNode(0)
	assert.Equal(t, int64(0), uuid.Node())
	assert.Equal(t, IETF, uuid.Variant())
	uuid.SetClockSequence(int(0))

	// test OverflowClockSequence
	uuid.SetClockSequence(int(0x13FFF))
	assert.Equal(t, int(0x3FFF), uuid.ClockSequence())
	assert.Equal(t, IETF, uuid.Variant())
	uuid.SetClockSequence(0)

	// testMaxClockSequence
	uuid.SetClockSequence(int(0x3FFF))
	assert.Equal(t, int(0x3FFF), uuid.ClockSequence())
	assert.Equal(t, IETF, uuid.Variant())

	// test clear ClockSequence
	uuid.SetNode(int64(0x0000FFFFFFFFFFFF))
	uuid.SetClockSequence(int(0))
	assert.Equal(t, int64(0x0000FFFFFFFFFFFF), uuid.Node())
	assert.Equal(t, IETF, uuid.Variant())
	uuid.SetNode(int64(0))

	// test maxTimeBits
	uuid.SetTime100Nanos(int64(0x0FFFFFFFFFFFFFFF))
	assert.Equal(t, int64(0x0FFFFFFFFFFFFFFF), uuid.Time100Nanos())
	assert.Equal(t, TimebasedVer1, uuid.Version())

	// test clear maxTimeBits
	uuid.SetTime100Nanos(0)
	assert.Equal(t, int64(0), uuid.Time100Nanos())
	assert.Equal(t, TimebasedVer1, uuid.Version())

	// test Milliseconds
	uuid.SetUnixTimeMillis(1)
	assert.Equal(t, int64(1), uuid.UnixTimeMillis())

	// test Negative Milliseconds
	uuid.SetUnixTimeMillis(-1)
	assert.Equal(t, int64(-1), uuid.UnixTimeMillis())

	// clear
	uuid.SetUnixTimeMillis(0)
	assert.Equal(t, int64(0), uuid.UnixTimeMillis())

	// test Counter

	uuid = NewUUID(TimebasedVer1)

	uuid.SetMinCounter()
	fmt.Print("min=", uuid.String(), "\n")
	fmt.Printf("counter=%x\n", uuid.Counter())
	binMin, _ := uuid.MarshalSortableBinary()

	uuid.SetMaxCounter()
	fmt.Print("max=", uuid.String(), "\n")
	fmt.Printf("counter=%x\n", uuid.Counter())
	binMax, _ := uuid.MarshalSortableBinary()

	for i := 1; i != 100; i = i + 1 {

		anyNumber := int64(i)
		uuid.SetCounter(anyNumber)

		binLesser, _ := uuid.MarshalSortableBinary()
		uuid.SetCounter(anyNumber + 1)

		binGreater, _ := uuid.MarshalSortableBinary()

		assert.True(t, bytes.Compare(binMin, binLesser) < 0, "min failed")
		assert.True(t, bytes.Compare(binLesser, binGreater) < 0, "seq failed")
		assert.True(t, bytes.Compare(binGreater, binMax) < 0, "max failed")
	}

	uuid = NewUUID(TimebasedVer1)

	current := time.Now()

	uuid.SetTime(current)
	cnt := uuid.SetCounter(rand.Int63())

	assert.Equal(t, current.UnixNano()/100, uuid.Time().UnixNano()/100)
	assert.Equal(t, cnt, uuid.Counter())

	assertMarshalText(t, uuid)
	assertMarshalJson(t, uuid)
	assertMarshalBinary(t, uuid)
	assertMarshalSortableBinary(t, uuid)

}

func testRandomlyGeneratedUUID(t *testing.T) {

	uuid := NewUUID(RandomlyGeneratedVer4)
	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, RandomlyGeneratedVer4, uuid.Version())

	uuid, err := RandomUUID()

	if err != nil {
		t.Fatal("fail to create random uuid ", err)
	}

	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, RandomlyGeneratedVer4, uuid.Version())

	assertMarshalText(t, uuid)
	assertMarshalJson(t, uuid)
	assertMarshalBinary(t, uuid)

}

func testNamebasedUUID(t *testing.T) {

	uuid := NewUUID(NamebasedVer5)
	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, NamebasedVer5, uuid.Version())

	uuid = NewUUID(NamebasedVer3)
	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, NamebasedVer3, uuid.Version())

	uuid, err := NameUUIDFromBytes([]byte("alex"), NamebasedVer3)

	if err != nil {
		t.Fatal("fail to create random uuid ", err)
	}

	assert.Equal(t, IETF, uuid.Variant())
	assert.Equal(t, NamebasedVer3, uuid.Version())
	assert.Equal(t, uint64(0x534b44a19bf13d20), uuid.MostSigBits)
	assert.Equal(t, uint64(0xb71ecc4eb77c572f), uuid.LeastSigBits)

	assert.Equal(t, "534b44a1-9bf1-3d20-b71e-cc4eb77c572f", uuid.String())

	assertMarshalText(t, uuid)
	assertMarshalJson(t, uuid)
	assertMarshalBinary(t, uuid)

}

func assertMarshalText(t *testing.T, uuid UUID) {

	var actual UUID
	data, err := uuid.MarshalText()

	if err != nil {
		t.Fatal("fail to MarshalText ", err)
	}

	err = actual.UnmarshalText(data)

	if err != nil {
		t.Fatal("fail to MarshalText ", err)
	}

	assert.Equal(t, uuid.MostSigBits, actual.MostSigBits)
	assert.Equal(t, uuid.LeastSigBits, actual.LeastSigBits)

}

func assertMarshalJson(t *testing.T, uuid UUID) {

	var actual UUID
	data, err := uuid.MarshalJSON()

	if err != nil {
		t.Fatal("fail to MarshalJson ", err)
	}

	err = actual.UnmarshalJSON(data)

	if err != nil {
		t.Fatal("fail to UnmarshalJson ", err)
	}

	assert.Equal(t, uuid.MostSigBits, actual.MostSigBits)
	assert.Equal(t, uuid.LeastSigBits, actual.LeastSigBits)

}
func assertMarshalBinary(t *testing.T, uuid UUID) {

	var actual UUID
	data, err := uuid.MarshalBinary()

	if err != nil {
		t.Fatal("fail to MarshalBinary ", err)
	}

	err = actual.UnmarshalBinary(data)

	if err != nil {
		t.Fatal("fail to UnmarshalBinary ", err)
	}

	assert.Equal(t, uuid.MostSigBits, actual.MostSigBits)
	assert.Equal(t, uuid.LeastSigBits, actual.LeastSigBits)

}

func assertMarshalSortableBinary(t *testing.T, uuid UUID) {

	var actual UUID
	data, err := uuid.MarshalSortableBinary()

	if err != nil {
		t.Fatal("fail to MarshalSortableBinary ", err)
	}

	err = actual.UnmarshalSortableBinary(data)

	if err != nil {
		t.Fatal("fail to UnmarshalSortableBinary ", err)
	}

	assert.Equal(t, uuid.MostSigBits, actual.MostSigBits)
	assert.Equal(t, uuid.LeastSigBits, actual.LeastSigBits)

}
