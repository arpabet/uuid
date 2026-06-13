/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"time"
)

/**
State for the monotonic version 1 generator.

The node is randomly generated once with the multicast bit set (RFC 4122 §4.5)
so it can not collide with a real IEEE 802 MAC address. The clock sequence is
seeded randomly and incremented whenever the clock does not advance, keeping
generated values strictly increasing within a single process.
*/

var (
	v1mu          sync.Mutex
	v1Initialized bool
	v1ClockSeq    uint16
	v1Node        int64
	v1LastTime    uint64 // last used timestamp in 100-nanosecond units since the Gregorian epoch
)

func initV1Locked() error {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return err
	}
	v1ClockSeq = binary.BigEndian.Uint16(b[0:2]) & uint16(clockSequenceBits)

	node := int64(0)
	for _, octet := range b[2:8] {
		node = (node << 8) | int64(octet)
	}
	// set the multicast bit of the first octet so the random node can not be
	// mistaken for a globally assigned IEEE 802 address.
	node |= int64(0x010000000000)
	v1Node = node & nodeMask

	v1Initialized = true
	return nil
}

/**
Generates a time-based version 1 UUID using the current time, a process wide
random node and a monotonic clock sequence.

Safe for concurrent use. Values generated within a single process are strictly
increasing even if the wall clock does not advance or moves backwards.
*/

func NewV1() (uuid UUID, err error) {

	v1mu.Lock()
	defer v1mu.Unlock()

	if !v1Initialized {
		if err = initV1Locked(); err != nil {
			return Empty, err
		}
	}

	now := uint64(time.Now().UnixNano()/100) + uint64(num100NanosSinceUUIDEpoch)

	if now <= v1LastTime {
		// clock did not advance: bump the sequence and force monotonic time
		v1ClockSeq = (v1ClockSeq + 1) & uint16(clockSequenceBits)
		now = v1LastTime + 1
	}
	v1LastTime = now

	uuid = New(TimebasedVer1)
	uuid.SetTime100NanosUnsigned(now & 0x0FFFFFFFFFFFFFFF)
	uuid.SetClockSequence(int(v1ClockSeq))
	uuid.SetNode(v1Node)

	return uuid, nil
}

/**
Generates a time-ordered version 7 UUID (RFC 9562).

Layout:

	msb: 48-bit Unix time in milliseconds + 4-bit version + 12-bit random
	lsb: 2-bit variant + 62-bit random

The leading millisecond timestamp makes version 7 UUIDs naturally sortable as
both binary and text without any custom encoding.
*/

func NewV7() (uuid UUID, err error) {

	var b [10]byte
	if _, err = rand.Read(b[:]); err != nil {
		return Empty, err
	}

	ms := uint64(time.Now().UnixMilli()) & 0xFFFFFFFFFFFF // 48 bits

	randA := uint64(binary.BigEndian.Uint16(b[0:2])) & 0x0FFF
	uuid.MostSigBits = (ms << 16) | (uint64(UnixTimeVer7) << 12) | randA

	randB := binary.BigEndian.Uint64(b[2:10])
	uuid.LeastSigBits = (randB & counterMask) | variantIETFBits

	return uuid, nil
}

/**
Gets the Unix timestamp in milliseconds from a version 7 UUID.

Valid only for version 7.
*/

func (this UUID) UnixTimeMillisV7() int64 {
	return int64(this.MostSigBits >> 16)
}

/**
Gets the time encoded in a version 7 UUID.

Valid only for version 7.
*/

func (this UUID) TimeV7() time.Time {
	return time.UnixMilli(this.UnixTimeMillisV7())
}
