/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package uuid_test

import (
	"go.arpabet.com/uuid"
	"testing"
)

var benchID = func() uuid.UUID {
	id, _ := uuid.Parse("534b44a1-9bf1-3d20-b71e-cc4eb77c572f")
	return id
}()

var benchString = benchID.String()

func BenchmarkString(b *testing.B) {
	b.ReportAllocs()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = benchID.String()
	}
	_ = sink
}

func BenchmarkURN(b *testing.B) {
	b.ReportAllocs()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = benchID.URN()
	}
	_ = sink
}

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	var sink uuid.UUID
	for i := 0; i < b.N; i++ {
		sink, _ = uuid.Parse(benchString)
	}
	_ = sink
}

func BenchmarkParseHex32(b *testing.B) {
	b.ReportAllocs()
	const s = "534b44a19bf13d20b71ecc4eb77c572f"
	var sink uuid.UUID
	for i := 0; i < b.N; i++ {
		sink, _ = uuid.Parse(s)
	}
	_ = sink
}

func BenchmarkMarshalText(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = benchID.MarshalText()
	}
}

func BenchmarkMarshalBinary(b *testing.B) {
	b.ReportAllocs()
	var dst [16]byte
	for i := 0; i < b.N; i++ {
		_ = benchID.MarshalBinaryTo(dst[:])
	}
}

func BenchmarkNewV7(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = uuid.NewV7()
	}
}

func BenchmarkNewV1(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = uuid.NewV1()
	}
}
