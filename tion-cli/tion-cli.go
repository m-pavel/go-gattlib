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
	var on = flag.Bool("on", false, "Turn on")
	var off = flag.Bool("off", false, "Turn off")
	var temp = flag.Int("temp", 0, "Set target temperature")
	var gate = flag.String("gate", "", "Set gate position(indoor|outdoor|mixed)")
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	if *device == "" && !*scanp {
		log.Fatal("Device address is mandatory")
	}

	if *on {
		deviceCall(*device, func(t *tion.Tion) {
			err := t.On()
			if err != nil {
				log.Println(err)
			}
		}, "Turned on")

		return
	}
	if *off {
		deviceCall(*device, func(t *tion.Tion) {
			err := t.Off()
			if err != nil {
				log.Println(err)
			}
		}, "Turned off")
		return
	}
	if *temp != 0 {
		deviceCall(*device, func(t *tion.Tion) {
			s := t.Status()
			s.TempTarget = byte(*temp)
			err := t.Update(s)
			if err != nil {
				log.Println(err)
			}
		}, fmt.Sprintf("Target temperature updated to %d", *temp))
		return
	}

	if *gate != "" {
		deviceCall(*device, func(t *tion.Tion) {
			s := t.Status()
			s.SetGateStatus(*gate)
			err := t.Update(s)
			if err != nil {
				log.Println(err)
			}
		}, fmt.Sprintf("TGate set to %s", *gate))
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
		log.Printf("Gate: %s, Error: %d\n", state.GateStatus(), state.ErrorCode)
	}
}

func sts(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func deviceCall(addr string, cb func(*tion.Tion), succ string) error {
	t := tion.New(addr)
	err := t.Connect()
	if err != nil {
		return err
	}
	defer t.Disconnect()
	cb(t)
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
