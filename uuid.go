/*
 * Copyright (c) 2023 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package uuid

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

/**
UUID represented as two 64-bit unsigned longs in the similar way like in Java
*/

type UUID struct {
	MostSigBits  uint64
	LeastSigBits uint64
}

/**
Zero version of the UUID
*/

var Empty = UUID{0, 0}

type Variant int

// Constants returned by Variant.
const (
	NCSReserved       = Variant(iota)
	IETF              // The IETF variant specified in RFC4122
	MicrosoftReserved // Reserved, Microsoft Corporation backward compatibility.
	FutureReserved    // Reserved for future definition.
	UnknownVariant
)

const (
	variantIETFBits = uint64(0x80) << 56

	one100NanosInSecond       = int64(time.Second) / 100
	one100NanosInMillis       = int64(time.Millisecond) / 100
	num100NanosSinceUUIDEpoch = int64(0x01b21dd213814000)

	versionMask          = uint64(0x000000000000F000)
	timebasedVersionBits = uint64(0x0000000000001000)
	maxTimeBits          = uint64(0xFFFFFFFFFFFF0FFF)

	nodeMask      = int64(0x0000FFFFFFFFFFFF)
	nodeClearMask = uint64(0xFFFF000000000000)

	clockSequenceBits      = int(0x3FFF)
	clockSequenceClearMask = uint64(0xC000FFFFFFFFFFFF)

	flipSignedBits = uint64(0x0080808080808080)

	counterMask    = uint64(0x3FFFFFFFFFFFFFFF)
	minCounterBits = uint64(0x0080808080808080)
	maxCounterBits = uint64(0x7f7f7f7f7f7f7f7f)
)

var (
	ErrorWrongLen              = errors.New("wrong len")
	ErrorRequiredTimebasedUUID = errors.New("required timebased UUID")
)

type Version int

// Constants returned by Version.
const (
	BadVersion = Version(iota)
	TimebasedVer1
	DCESecurityVer2
	NamebasedVer3
	RandomlyGeneratedVer4
	NamebasedVer5
	ReorderedTimeVer6 // RFC 9562 reordered Gregorian time, sortable
	UnixTimeVer7      // RFC 9562 Unix Epoch time, sortable
	UnknownVersion
)

/**
Compare two required values of UUID
*/

func (u UUID) Equal(other UUID) bool {
	return u.MostSigBits == other.MostSigBits && u.LeastSigBits == other.LeastSigBits
}

/**
	Compare two optional values of UUID

    return true if both are nil or equal
*/

func Equal(left *UUID, right *UUID) bool {
	if left != nil {
		if right != nil {
			return left.Equal(*right)
		} else {
			return false
		}
	} else {
		return right == nil
	}
}

/**
Creates new UUID for the specific version
*/

func New(version Version) (uuid UUID) {
	uuid.MostSigBits = uint64(version) << 12
	uuid.LeastSigBits = variantIETFBits
	return uuid
}

/**
Creates UUID from the specific most and least sig bits
*/

func Create(MostSigBits, LeastSigBits int64) (uuid UUID) {
	uuid.MostSigBits = uint64(MostSigBits)
	uuid.LeastSigBits = uint64(LeastSigBits)
	return uuid
}

/**
Gets most significant bits as long
*/

func (u UUID) MostSignificantBits() int64 {
	return int64(u.MostSigBits)
}

/**
Sets most significant bits from long
*/

func (u *UUID) SetMostSignificantBits(MostSigBits int64) {
	u.MostSigBits = uint64(MostSigBits)
}

/**
Gets least significant bits as long
*/

func (u UUID) LeastSignificantBits() int64 {
	return int64(u.LeastSigBits)
}

/**
Sets least significant bits from long
*/

func (u *UUID) SetLeastSignificantBits(LeastSigBits int64) {
	u.LeastSigBits = uint64(LeastSigBits)
}

/**
  Stores UUID in to 16 bytes

  MarshalBinary implements the encoding.BinaryMarshaler interface.
*/

func (u UUID) MarshalBinary() (dst []byte, err error) {
	dst = make([]byte, 16)
	err = u.MarshalBinaryTo(dst)
	return dst, err

}

/**
  Stores UUID in to slice
*/

func (u UUID) MarshalBinaryTo(dst []byte) error {

	if len(dst) < 16 {
		return ErrorWrongLen
	}

	binary.BigEndian.PutUint64(dst, u.MostSigBits)
	binary.BigEndian.PutUint64(dst[8:], u.LeastSigBits)

	return nil
}

/**
  Convert serialized 16 bytes to UUID

  UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
*/

func (u *UUID) UnmarshalBinary(data []byte) error {

	if len(data) < 16 {
		return ErrorWrongLen
	}

	u.MostSigBits = binary.BigEndian.Uint64(data)
	u.LeastSigBits = binary.BigEndian.Uint64(data[8:])

	return nil
}

/**
  Stores UUID in to 16 bytes by flipping timestamp parts to make byte array sortable

  Used only for Time-based UUID
*/

func (u UUID) MarshalSortableBinary() ([]byte, error) {
	dst := make([]byte, 16)
	err := u.MarshalSortableBinaryTo(dst)
	return dst, err
}

/**
  Stores UUID in to the slice by flipping timestamp parts to make byte array sortable and converts signed bytes to unsigned

  Used only for Time-based UUID

  Result:

  msb: 4-bit version + 60-bit timestamp in 100 nanos
  lsb: 2-bit variant + 62-bit counter (clockSequence and Node) converted to unsigned bytes

*/

func (u UUID) MarshalSortableBinaryTo(dst []byte) error {

	if len(dst) < 16 {
		return ErrorWrongLen
	}

	versionAndTimeHigh := uint16(u.MostSigBits)

	if versionAndTimeHigh&0xF000 != 0x1000 {
		return ErrorRequiredTimebasedUUID
	}

	timeMid := uint16(u.MostSigBits >> 16)
	timeLow := uint32(u.MostSigBits >> 32)

	binary.BigEndian.PutUint16(dst, versionAndTimeHigh)
	binary.BigEndian.PutUint16(dst[2:], timeMid)
	binary.BigEndian.PutUint32(dst[4:], timeLow)
	binary.BigEndian.PutUint64(dst[8:], u.LeastSigBits^flipSignedBits)

	return nil
}

/**
  Convert sortable representation of serialized 16 bytes to UUID

  Sortable representation flips timestamp blocks to make TimeUUID sortable as byte array and converts signed bytes to unsigned

  Used only for Time-based UUID

  Data:

  msb: 4-bit version + 60-bit timestamp in 100 nanos
  lsb: 2-bit variant + 62-bit counter (clockSequence and Node) converted to unsigned bytes

*/

func (u *UUID) UnmarshalSortableBinary(data []byte) error {

	if len(data) < 16 {
		return ErrorWrongLen
	}

	versionAndTimeHigh := uint64(binary.BigEndian.Uint16(data))

	if versionAndTimeHigh&0xF000 != 0x1000 {
		return ErrorRequiredTimebasedUUID
	}

	timeMid := uint64(binary.BigEndian.Uint16(data[2:]))
	timeLow := uint64(binary.BigEndian.Uint32(data[4:]))

	u.MostSigBits = (timeLow << 32) | (timeMid << 16) | versionAndTimeHigh
	u.LeastSigBits = binary.BigEndian.Uint64(data[8:]) ^ flipSignedBits

	return nil
}

/**
  Generates random UUID by using pseudo-random cryptographic generator
*/

func RandomUUID() (uuid UUID, err error) {

	var randomBytes = make([]byte, 16)
	if _, err = rand.Read(randomBytes); err != nil {
		return Empty, err
	}

	randomBytes[6] &= 0x0f /* clear version        */
	randomBytes[6] |= 0x40 /* set to version 4     */
	randomBytes[8] &= 0x3f /* clear variant        */
	randomBytes[8] |= 0x80 /* set to IETF variant  */

	err = uuid.UnmarshalBinary(randomBytes)
	return uuid, err

}

/**
	Creates UUID based on digest of incoming byte array
    Used for authentication purposes
*/

func NameUUIDFromBytes(name []byte, version Version) (uuid UUID, err error) {
	err = uuid.SetName(name, version)
	return uuid, err
}

/**
	Sets name digest of incoming byte array
    Used for authentication purposes
*/

func (u *UUID) SetName(name []byte, version Version) error {

	switch version {

	case NamebasedVer3:

		digest := md5.Sum(name)

		digest[6] &= 0x0f /* clear version        */
		digest[6] |= 0x30 /* set to version 3     */
		digest[8] &= 0x3f /* clear variant        */
		digest[8] |= 0x80 /* set to IETF variant  */

		return u.UnmarshalBinary(digest[:])

	case NamebasedVer5:

		digest := sha1.Sum(name)

		digest[6] &= 0x0f /* clear version        */
		digest[6] |= 0x50 /* set to version 5     */
		digest[8] &= 0x3f /* clear variant        */
		digest[8] |= 0x80 /* set to IETF variant  */

		return u.UnmarshalBinary(digest[:])

	default:
		return errors.Errorf("unknown namebased version: %q", version)
	}

}

/**
  Gets version of the UUID
*/

func (u UUID) Version() Version {

	version := int((u.MostSigBits & versionMask) >> 12)

	if version >= int(UnknownVersion) {
		return UnknownVersion
	}

	return Version(version)
}

/**
Gets variant of the UUID
*/

func (u UUID) Variant() Variant {

	variant := int((u.LeastSigBits >> 56) & 0xFF)

	// This field is composed of a varying number of bits.
	// 0    x    x   x   Reserved for NCS backward compatibility
	// 1    0    x   x   The IETF aka Leach-Salz variant (used by this class)
	// 1    1    0   x   Reserved, Microsoft backward compatibility
	// 1    1    1   x   Reserved for future definition.

	switch {
	case variant&0x80 == 0:
		return NCSReserved
	case variant&0xC0 == 0x80:
		return IETF
	case variant&0xE0 == 0xC0:
		return MicrosoftReserved
	case variant&0xE0 == 0xE0:
		return FutureReserved
	default:
		return UnknownVariant
	}
}

/**
  Gets timestamp as 60bit int64 from Time-based UUID

  It is measured in 100-nanosecond units since midnight, October 15, 1582 UTC.

  valid only for version 1 or 2
*/

func (u UUID) Time100Nanos() int64 {
	return int64(u.Time100NanosUnsigned())
}

/**
  Gets timestamp as 60bit uint64 from Time-based UUID

  It is measured in 100-nanosecond units since midnight, October 15, 1582 UTC.

  valid only for version 1 or 2
*/

func (u UUID) Time100NanosUnsigned() uint64 {

	timeHigh := u.MostSigBits & 0x0FFF
	timeMid := (u.MostSigBits >> 16) & 0xFFFF
	timeLow := (u.MostSigBits >> 32) & 0xFFFFFFFF

	return (timeHigh << 48) | (timeMid << 32) | timeLow
}

/**
Sets 60-bit time in 100 nanoseconds since midnight, October 15, 1582 UTC.
*/

func (u *UUID) SetTime100Nanos(time100Nanos int64) {
	u.SetTime100NanosUnsigned(uint64(time100Nanos))
}

/**
Sets 60-bit time in 100 nanoseconds since midnight, October 15, 1582 UTC.
*/

func (u *UUID) SetTime100NanosUnsigned(time100Nanos uint64) {

	bits := timebasedVersionBits

	// timeLow
	bits |= (time100Nanos & 0xFFFFFFFF) << 32

	// timeMid
	bits |= (time100Nanos & 0xFFFF00000000) >> 16

	// timeHigh
	bits |= (time100Nanos & 0xFFF000000000000) >> 48

	u.MostSigBits = bits

}

/**
Sets minimum possible 60-bit time value
*/

func (u *UUID) SetMinTime() {
	u.MostSigBits = timebasedVersionBits
}

/**
Sets maximum possible 60-bit time value
*/

func (u *UUID) SetMaxTime() {
	u.MostSigBits = timebasedVersionBits | maxTimeBits
}

/**
Gets timestamp in milliseconds from Time-based UUID

It is measured in millisecond units in unix time since 1 Jan 1970
*/

func (u UUID) UnixTimeMillis() int64 {
	return (u.Time100Nanos() - num100NanosSinceUUIDEpoch) / one100NanosInMillis
}

/**
	Sets timestamp in milliseconds to Time-based UUID

    It is measured in millisecond units in unix time since 1 Jan 1970
*/

func (u *UUID) SetUnixTimeMillis(unixTimeMillis int64) {
	time100Nanos := (unixTimeMillis * one100NanosInMillis) + num100NanosSinceUUIDEpoch
	u.SetTime100Nanos(time100Nanos)
}

/**
Gets timestamp in 100 nanoseconds from Time-based UUID

It is measured in millisecond units in unix time since 1 Jan 1970
*/

func (u UUID) UnixTime100Nanos() int64 {
	return u.Time100Nanos() - num100NanosSinceUUIDEpoch
}

/**
	Sets timestamp in 100 nanoseconds to Time-based UUID

    It is measured in millisecond units in unix time since 1 Jan 1970
*/

func (u *UUID) SetUnixTime100Nanos(unixTime100Nanos int64) {
	u.SetTime100Nanos(unixTime100Nanos + num100NanosSinceUUIDEpoch)
}

/**
Gets Time from Time-based UUID
*/

func (u UUID) Time() time.Time {
	unixTime100Nanos := u.UnixTime100Nanos()
	return time.Unix(unixTime100Nanos/one100NanosInSecond, (unixTime100Nanos%one100NanosInSecond)*100)
}

/**
Sets Time to Time-based UUID
*/

func (u *UUID) SetTime(t time.Time) {
	sec := t.Unix()
	nsec := int64(t.Nanosecond())
	one100Nanos := (nsec / 100) % one100NanosInSecond
	u.SetUnixTime100Nanos(sec*one100NanosInSecond + one100Nanos)
}

/**
  Gets raw 14 bit clock sequence value from Time-based UUID

  unsigned in range [0, 0x3FFF]

  Does not convert signed to unsigned
*/

func (u UUID) ClockSequence() int {
	variantAndSequence := u.LeastSigBits >> 48
	return int(variantAndSequence) & clockSequenceBits
}

/**
	Sets raw 14 bit clock sequence value to Time-based UUID

    unsigned in range [0, 0x3FFF]

    Does not convert signed to unsigned
*/

func (u *UUID) SetClockSequence(clockSequence int) {
	sanitizedSequence := uint64(clockSequence & clockSequenceBits)
	u.LeastSigBits = (u.LeastSigBits & clockSequenceClearMask) | (sanitizedSequence << 48)
}

/**
  Gets raw node value associated with Time-based UUID

  48 bit node is intended to hold the IEEE 802 address of the machine that generated this UUID to guarantee spatial uniqueness.

  unsigned in range [0, 0xFFFFFFFFFFFF]

  Does not convert signed to unsigned
*/

func (u UUID) Node() int64 {
	return int64(u.LeastSigBits) & nodeMask
}

/**
	Stores raw 48 bit value to the node

    unsigned in range [0, 0xFFFFFFFFFFFF]

    Does not convert signed to unsigned
*/

func (u *UUID) SetNode(node int64) {
	sanitizedNode := uint64(node & nodeMask)
	u.LeastSigBits = (u.LeastSigBits & nodeClearMask) | sanitizedNode
}

/**
	Gets counter in range [0 to 3fffffffffffffff] sequence_and_variant

    Counter is the composition of ClockSequenceAndNode

    Converts from signed values automatically
*/

func (u UUID) Counter() int64 {
	return int64(u.CounterUnsigned())
}

/**
	Gets counter in range [0 to 3fffffffffffffff] sequence_and_variant

    Counter is the composition of ClockSequenceAndNode

    Converts from signed values automatically
*/

func (u UUID) CounterUnsigned() uint64 {
	return (u.LeastSigBits ^ flipSignedBits) & counterMask
}

/**
	Sets counter in range [0 to 3fffffffffffffff] sequence_and_variant

    Counter is the composition of ClockSequenceAndNode

    Converts to signed values automatically

    return sanitized value stored in UUID
*/

func (u *UUID) SetCounter(counter int64) int64 {
	return int64(u.SetCounterUnsigned(uint64(counter)))
}

/**
	Sets counter in range [0 to 3fffffffffffffff] sequence_and_variant

    Counter is the composition of ClockSequenceAndNode

    Converts to signed values automatically

    return sanitized value stored in UUID
*/

func (u *UUID) SetCounterUnsigned(counter uint64) uint64 {
	sanitizedCounter := counter & counterMask
	u.LeastSigBits = (sanitizedCounter | variantIETFBits) ^ flipSignedBits
	return sanitizedCounter
}

/**
  Sets min counter (sequence_and_variant)

  Guarantees that in sortable binary block will be first after sorting
*/

func (u *UUID) SetMinCounter() {
	u.LeastSigBits = minCounterBits | variantIETFBits
}

/**
  Sets max counter (sequence_and_variant)

  Guarantees that in sortable binary block will be last after sorting
*/

func (u *UUID) SetMaxCounter() {
	u.LeastSigBits = maxCounterBits | variantIETFBits
}

/**
Parses string representation of UUID
*/

func Parse(s string) (UUID, error) {
	// the longest accepted form is the 45 character urn:uuid: representation;
	// copy into a stack buffer to avoid heap allocating the string conversion.
	if len(s) > 45 {
		return Empty, fmt.Errorf("invalid UUID length: %q", s)
	}
	var buf [45]byte
	n := copy(buf[:], s)
	return ParseBytes(buf[:n])
}

/**
  Parses bytes are a string representation of UUID
*/

func ParseBytes(src []byte) (UUID, error) {

	switch len(src) {

	// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case 36:
		if src[8] != '-' || src[13] != '-' || src[18] != '-' || src[23] != '-' {
			return Empty, fmt.Errorf("invalid UUID format: %q", string(src))
		}
		var data [16]byte
		if _, e1 := hex.Decode(data[0:4], src[0:8]); e1 != nil {
			return Empty, fmt.Errorf("invalid UUID format: %q", string(src))
		}
		if _, e2 := hex.Decode(data[4:6], src[9:13]); e2 != nil {
			return Empty, fmt.Errorf("invalid UUID format: %q", string(src))
		}
		if _, e3 := hex.Decode(data[6:8], src[14:18]); e3 != nil {
			return Empty, fmt.Errorf("invalid UUID format: %q", string(src))
		}
		if _, e4 := hex.Decode(data[8:10], src[19:23]); e4 != nil {
			return Empty, fmt.Errorf("invalid UUID format: %q", string(src))
		}
		if _, e5 := hex.Decode(data[10:16], src[24:36]); e5 != nil {
			return Empty, fmt.Errorf("invalid UUID format: %q", string(src))
		}
		return fromBytes(data), nil

	// urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case 36 + 9:
		if !bytes.EqualFold(src[:9], []byte("urn:uuid:")) {
			return Empty, fmt.Errorf("invalid urn prefix in %q", string(src))
		}
		return ParseBytes(src[9:])

	// {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx} or "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	case 36 + 2:
		switch {
		case src[0] == '{' && src[37] == '}':
		case src[0] == '"' && src[37] == '"':
		default:
			return Empty, fmt.Errorf("invalid UUID wrapper in %q", string(src))
		}
		return ParseBytes(src[1:37])

	// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	case 32:
		var data [16]byte
		if _, err := hex.Decode(data[:], src); err != nil {
			return Empty, fmt.Errorf("invalid UUID format: %q: %w", string(src), err)
		}
		return fromBytes(data), nil

	default:
		return Empty, fmt.Errorf("invalid UUID length: %q", string(src))
	}
}

// fromBytes builds a UUID from a decoded 16 byte big-endian representation.
func fromBytes(data [16]byte) UUID {
	return UUID{
		MostSigBits:  binary.BigEndian.Uint64(data[:]),
		LeastSigBits: binary.BigEndian.Uint64(data[8:]),
	}
}

/**
UnmarshalText implements the encoding.TextUnmarshaler interface.
*/

func (u *UUID) UnmarshalText(data []byte) error {
	var err error
	*u, err = ParseBytes(data)
	return err
}

/**
  MarshalText implements the encoding.TextMarshaler interface.
*/

func (u UUID) MarshalText() ([]byte, error) {
	dst := make([]byte, 36)
	err := u.MarshalTextTo(dst)
	return dst, err
}

/**
Marshal text to preallocated slice
*/

func (u UUID) MarshalTextTo(dst []byte) error {

	if len(dst) < 36 {
		return ErrorWrongLen
	}

	var data [16]byte
	binary.BigEndian.PutUint64(data[:], u.MostSigBits)
	binary.BigEndian.PutUint64(data[8:], u.LeastSigBits)

	hex.Encode(dst, data[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], data[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], data[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], data[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], data[10:])
	return nil
}

/**
UnmarshalJSON implements the json.Unmarshaler interface.
*/

func (u *UUID) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	*u, err = ParseBytes(data)
	return err
}

/**
MarshalJSON implements the json.Marshaler interface.
*/

func (u UUID) MarshalJSON() ([]byte, error) {

	jsonVal := make([]byte, 36+2)
	jsonVal[0] = '"'
	jsonVal[37] = '"'
	err := u.MarshalTextTo(jsonVal[1:37])

	return jsonVal, err
}

/**
	Converts UUID in to string

    For Time-based UUID:

	<time_low> "-" <time_mid> "-" <time_high_and_version> "-" <variant_and_sequence> "-" <node>

	time_low               = 4*<hexOctet>
    time_mid               = 2*<hexOctet>
    version_and_time_high  = 2*<hexOctet>
    sequence_and_variant   = 2*<hexOctet>
    node                   = 6*<hexOctet>

*/

func (u UUID) String() string {
	var buf [36]byte
	u.MarshalTextTo(buf[:])
	return string(buf[:])
}

/**
Gets URN name of the UUID
*/

func (u UUID) URN() string {
	var buf [45]byte
	copy(buf[:9], "urn:uuid:")
	u.MarshalTextTo(buf[9:])
	return string(buf[:])
}

/**
Gets version name
*/

func (v Version) String() string {
	switch v {
	case TimebasedVer1:
		return "TimebasedVer1"
	case DCESecurityVer2:
		return "DCESecurityVer2"
	case NamebasedVer3:
		return "NamebasedVer3"
	case RandomlyGeneratedVer4:
		return "RandomlyGeneratedVer4"
	case NamebasedVer5:
		return "NamebasedVer5"
	case ReorderedTimeVer6:
		return "ReorderedTimeVer6"
	case UnixTimeVer7:
		return "UnixTimeVer7"
	}
	return fmt.Sprintf("BadVersion%d", int(v))
}

/**
Gets variant name
*/

func (v Variant) String() string {
	switch v {
	case IETF:
		return "IETF"
	case NCSReserved:
		return "NCSReserved"
	case MicrosoftReserved:
		return "MicrosoftReserved"
	case FutureReserved:
		return "FutureReserved"
	}
	return fmt.Sprintf("BadVariant%d", int(v))
}

/**
Checks if varian is valid and supported by this module
*/

func (v Variant) Valid() bool {
	return v == IETF
}
