package tion

import (
	"testing"
)

func Test01(t *testing.T) {
	// 2 - b3 10 22 12 0b 00 12 12 08 60 01 13 14 00 2d 08 00 32 00 5a
	// 1 - b3 10 21 12 0b 00 10 10 08 60 01 13 16 00 1e 08 00 32 00 5a
	bts := []byte{0xb3, 0x10, 0x21, 0x12, 0x0b, 0x00, 0x10, 0x10, 0x08, 0x60, 0x01, 0x13, 0x16, 0x00, 0x1e, 0x08, 0x00, 0x32, 0x00, 0x5a}

	resp, err := FromBytes(bts)
	if err != nil {
		t.Fatal(err)
	}

	if resp.FiltersRemains != 352 {
		t.Fatal(352)
	}
}
