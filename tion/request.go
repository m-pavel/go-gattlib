package tion

import (
	"bytes"
)

var (
	statusRequest = []byte{0x3d, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5a}
)

func BuildRequest(enabled, sound, heater bool, speed, gate, temp byte) []byte {
	bf := bytes.NewBufferString("")
	bf.WriteByte(0x3d)
	bf.WriteByte(0x02)
	bf.WriteByte(speed)
	bf.WriteByte(temp)
	bf.WriteByte(gate)
	flags := byte(0)
	if heater {
		flags |= 1
	}
	if enabled {
		flags |= 2
	}
	if sound {
		flags |= 8
	}

	bf.WriteByte(flags)

	if heater {
		bf.WriteByte(0x01)
	} else {
		bf.WriteByte(0x00)
	}

	bf.Write([]byte{0x00, 0x00, 0x00, 0x00})
	bf.Write([]byte{0x00, 0x00})
	bf.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	bf.WriteByte(0x5a)
	return bf.Bytes()
}
