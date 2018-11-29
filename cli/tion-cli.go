package main

import (
	"flag"
	"log"
	"time"

	"github.com/go-ble/ble"

	"github.com/m-pavel/go-gattlib/pkg"

	"fmt"

	"github.com/m-pavel/go-gattlib/tion"
)

func main() {
	var device = flag.String("device", "", "bt addr")
	var status = flag.Bool("status", true, "Request status")
	var scanp = flag.Bool("scan", false, "Perform scan")
	var debug = flag.Bool("debug", false, "Debug")
	var on = flag.Bool("on", false, "Turn on")
	var off = flag.Bool("off", false, "Turn off")
	var temp = flag.Int("temp", 0, "Set target temperature")
	var speed = flag.Int("speed", 0, "Set speed")
	var sound = flag.String("sound", "", "Sound on|off")
	var heater = flag.String("heater", "", "Heater on|off")
	var gate = flag.String("gate", "", "Set gate position(indoor|outdoor|mixed)")
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	if *device == "" && !*scanp {
		log.Fatal("Device address is mandatory")
	}

	if *on {
		deviceCall(*device, *debug, func(t *tion.Tion) error {
			return t.On()
		}, "Turned on")

		return
	}
	if *off {
		deviceCall(*device, *debug, func(t *tion.Tion) error {
			return t.Off()
		}, "Turned off")
		return
	}
	if *temp != 0 {
		deviceCall(*device, *debug, func(t *tion.Tion) error {
			s := t.Status()
			s.TempTarget = byte(*temp)
			return t.Update(s)
		}, fmt.Sprintf("Target temperature updated to %d", *temp))
		return
	}

	if *speed != 0 {
		if *speed <= 0 || *speed > 6 {
			log.Println("Speed range 1..6")
			return
		}
		deviceCall(*device, *debug, func(t *tion.Tion) error {
			s := t.Status()
			s.Speed = byte(*speed)
			return t.Update(s)
		}, fmt.Sprintf("Speed updated to %d", *speed))
		return
	}

	if *gate != "" {
		deviceCall(*device, *debug, func(t *tion.Tion) error {
			s := t.Status()
			s.SetGateStatus(*gate)
			return t.Update(s)
		}, fmt.Sprintf("Gate set to %s", *gate))
		return
	}

	if *sound != "" {
		deviceCall(*device, *debug, func(t *tion.Tion) error {
			s := t.Status()
			if *sound == "on" {
				s.SoundEnabled = true
			} else {
				s.SoundEnabled = false
			}
			return t.Update(s)
		}, fmt.Sprintf("Sound set to %s", *sound))
		return
	}

	if *heater != "" {
		deviceCall(*device, *debug, func(t *tion.Tion) error {
			s := t.Status()
			if *heater == "on" {
				s.HeaterEnabled = true
			} else {
				s.HeaterEnabled = false
			}
			return t.Update(s)
		}, fmt.Sprintf("Heater set to %s", *heater))
		return
	}

	if *scanp {
		scan()
		return
	}

	if *status {
		state, err := tion.New(*device).ReadState()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Status: %s, Heater: %s, Sound: %s\n", sts(state.Enabled), sts(state.HeaterEnabled), sts(state.SoundEnabled))
		log.Printf("Target: %d \u2103, In: %d \u2103, Out: %d \u2103\n", state.TempTarget, state.TempIn, state.TempOut)
		log.Printf("Speed %d, Gate: %s, Error: %d\n", state.Speed, state.GateStatus(), state.ErrorCode)
	}
}

func sts(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func deviceCall(addr string, dbg bool, cb func(*tion.Tion) error, succ string) error {
	t := tion.New(addr)
	t.Debug(dbg)
	err := t.Connect()
	if err != nil {
		return err
	}
	defer t.Disconnect()
	err = cb(t)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(succ)
	}
	return err
}

func scan() {
	gattlib.Scan(func(ad ble.Advertisement) {
		log.Printf("%s %s", ad.Addr(), ad.LocalName())
	}, 5)
	time.Sleep(10 * time.Second)
}
