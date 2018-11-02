package tion

import (
	bytes2 "bytes"

	"github.com/go-errors/errors"
)

type Status struct {
	Enabled         bool
	HeaterEnabled   bool
	SoundEnabled    bool
	TimerEnabled    bool
	Speed           byte
	Gate            byte // 0 - indoor, 1 - mixed, 2 - outdoor; when 0 than heater off; when 1 speed 1,2 unavailiable
	TempTarget      byte
	TempOut         byte // Outcoming from device - inside
	TempIn          byte // Incoming to device - outside
	FiltersRemains  int
	Hours           byte
	Minutes         byte
	ErrorCode       byte
	Productivity    byte //m3pH
	RunDays         int
	FirmwareVersion int
	Todo            byte
}

func GateStatus(v byte) string {
	switch v {
	case 0:
		return "indoor"
	case 1:
		return "mixed"
	case 2:
		return "outdoor"
	default:
		return "unknown"
	}
}
func (s Status) GateStatus() string {
	return GateStatus(s.Gate)
}

func (s *Status) SetGateStatus(str string) {
	switch str {
	case "indoor":
		s.Gate = 0
	case "mixed":
		s.Gate = 1
	case "outdoor":
		s.Gate = 2
	}
}

func FromBytes(bytes []byte) (*Status, error) {
	if len(bytes) < 20 {
		return nil, errors.New("Expecting 20 bytes array")
	}
	buffer := bytes2.NewBuffer(bytes[2:])

	bt := rb(buffer)
	tr := Status{}
	tr.Speed = byte(int(bt) & 0xF)
	tr.Gate = bt >> 4
	tr.TempTarget, _ = buffer.ReadByte()

	bt = rb(buffer)
	if bt&1 != 0 {
		tr.HeaterEnabled = true
	}
	if bt&2 != 0 {
		tr.Enabled = true
	}
	if bt&4 != 0 {
		tr.TimerEnabled = true
	}
	if bt&8 != 0 {
		tr.SoundEnabled = true
	}
	tr.Todo = rb(buffer)
	//log.Println(tr.Todo)
	tr.TempOut = (rb(buffer) + rb(buffer)) / 2
	tr.TempIn = rb(buffer)
	tr.FiltersRemains = ri(buffer)
	tr.Hours = rb(buffer)
	tr.Minutes = rb(buffer)
	tr.ErrorCode = rb(buffer)
	tr.Productivity = rb(buffer)
	tr.RunDays = ri(buffer)
	tr.FirmwareVersion = ri(buffer)
	return &tr, nil
}

func rb(b *bytes2.Buffer) byte {
	bt, _ := b.ReadByte()
	return bt & 0xFF
}

func ri(b *bytes2.Buffer) int {
	return int(rb(b)) + int(rb(b))<<8
}
