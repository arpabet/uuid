/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package uuid_test

import (
	"github.com/stretchr/testify/assert"
	"go.arpabet.com/uuid"
	"strings"
	"testing"
)

func TestBitsAccessors(t *testing.T) {
	id := uuid.Create(0x0123456789abcdef, 0x7654321001234567)
	assert.Equal(t, int64(0x0123456789abcdef), id.MostSignificantBits())
	assert.Equal(t, int64(0x7654321001234567), id.LeastSignificantBits())

	var n uuid.UUID
	var u64 uint64 = 0x8877665544332211
	lsb := int64(u64)
	n.SetMostSignificantBits(0x1122334455667788)
	n.SetLeastSignificantBits(lsb)
	assert.Equal(t, int64(0x1122334455667788), n.MostSignificantBits())
	assert.Equal(t, lsb, n.LeastSignificantBits())
}

func TestMinMaxTime(t *testing.T) {
	id := uuid.New(uuid.TimebasedVer1)

	id.SetMaxTime()
	assert.Equal(t, uuid.TimebasedVer1, id.Version())
	assert.Equal(t, int64(0x0FFFFFFFFFFFFFFF), id.Time100Nanos())

	id.SetMinTime()
	assert.Equal(t, uuid.TimebasedVer1, id.Version())
	assert.Equal(t, int64(0), id.Time100Nanos())
}

func TestURN(t *testing.T) {
	id, err := uuid.Parse("534b44a1-9bf1-3d20-b71e-cc4eb77c572f")
	assert.NoError(t, err)

	urn := id.URN()
	assert.Equal(t, "urn:uuid:534b44a1-9bf1-3d20-b71e-cc4eb77c572f", urn)

	parsed, err := uuid.Parse(urn)
	assert.NoError(t, err)
	assert.True(t, id.Equal(parsed))
}

func TestVersionStrings(t *testing.T) {
	cases := map[uuid.Version]string{
		uuid.TimebasedVer1:         "TimebasedVer1",
		uuid.DCESecurityVer2:       "DCESecurityVer2",
		uuid.NamebasedVer3:         "NamebasedVer3",
		uuid.RandomlyGeneratedVer4: "RandomlyGeneratedVer4",
		uuid.NamebasedVer5:         "NamebasedVer5",
		uuid.ReorderedTimeVer6:     "ReorderedTimeVer6",
		uuid.UnixTimeVer7:          "UnixTimeVer7",
	}
	for v, name := range cases {
		assert.Equal(t, name, v.String())
	}
	assert.True(t, strings.HasPrefix(uuid.BadVersion.String(), "BadVersion"))
}

func TestVersionUnknown(t *testing.T) {
	// version nibble 9 is above the highest known version
	id := uuid.UUID{MostSigBits: 0x9000}
	assert.Equal(t, uuid.UnknownVersion, id.Version())

	id6 := uuid.UUID{MostSigBits: 0x6000}
	assert.Equal(t, uuid.ReorderedTimeVer6, id6.Version())
	id7 := uuid.UUID{MostSigBits: 0x7000}
	assert.Equal(t, uuid.UnixTimeVer7, id7.Version())
}

func TestVariants(t *testing.T) {
	assert.Equal(t, uuid.NCSReserved, uuid.UUID{LeastSigBits: uint64(0x00) << 56}.Variant())
	assert.Equal(t, uuid.IETF, uuid.UUID{LeastSigBits: uint64(0x80) << 56}.Variant())
	assert.Equal(t, uuid.MicrosoftReserved, uuid.UUID{LeastSigBits: uint64(0xC0) << 56}.Variant())
	assert.Equal(t, uuid.FutureReserved, uuid.UUID{LeastSigBits: uint64(0xE0) << 56}.Variant())

	assert.Equal(t, "IETF", uuid.IETF.String())
	assert.Equal(t, "NCSReserved", uuid.NCSReserved.String())
	assert.Equal(t, "MicrosoftReserved", uuid.MicrosoftReserved.String())
	assert.Equal(t, "FutureReserved", uuid.FutureReserved.String())
	assert.True(t, strings.HasPrefix(uuid.Variant(99).String(), "BadVariant"))

	assert.True(t, uuid.IETF.Valid())
	assert.False(t, uuid.NCSReserved.Valid())
}

func TestBinaryErrors(t *testing.T) {
	id := uuid.New(uuid.TimebasedVer1)

	short := make([]byte, 8)
	assert.Equal(t, uuid.ErrorWrongLen, id.MarshalBinaryTo(short))
	assert.Equal(t, uuid.ErrorWrongLen, id.MarshalTextTo(short))
	assert.Equal(t, uuid.ErrorWrongLen, id.MarshalSortableBinaryTo(short))

	var dst uuid.UUID
	assert.Equal(t, uuid.ErrorWrongLen, dst.UnmarshalBinary(short))
	assert.Equal(t, uuid.ErrorWrongLen, dst.UnmarshalSortableBinary(short))
}

func TestSortableBinaryRequiresTimebased(t *testing.T) {
	// a random (v4) UUID is not time-based
	id, err := uuid.RandomUUID()
	assert.NoError(t, err)
	_, err = id.MarshalSortableBinary()
	assert.Equal(t, uuid.ErrorRequiredTimebasedUUID, err)

	notTimebased := make([]byte, 16) // version nibble 0
	var dst uuid.UUID
	assert.Equal(t, uuid.ErrorRequiredTimebasedUUID, dst.UnmarshalSortableBinary(notTimebased))
}

func TestSetNameUnknownVersion(t *testing.T) {
	var id uuid.UUID
	err := id.SetName([]byte("x"), uuid.RandomlyGeneratedVer4)
	assert.Error(t, err)

	_, err = uuid.NameUUIDFromBytes([]byte("x"), uuid.BadVersion)
	assert.Error(t, err)
}

func TestParseTooLong(t *testing.T) {
	_, err := uuid.Parse(strings.Repeat("a", 100))
	assert.Error(t, err)
}

func TestUnmarshalJSON(t *testing.T) {
	var id uuid.UUID
	// null keeps the zero value without error
	assert.NoError(t, id.UnmarshalJSON([]byte("null")))
	assert.True(t, uuid.Empty.Equal(id))

	// invalid content reports an error
	assert.Error(t, id.UnmarshalJSON([]byte("\"not-a-uuid\"")))
}
