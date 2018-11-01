package main

import (
	"flag"
	"log"
	"time"

	"github.com/go-ble/ble"

	"github.com/m-pavel/go-gattlib/pkg"

	"strconv"

	"fmt"

	"github.com/m-pavel/go-gattlib/tion"
)

func main() {
	var device = flag.String("device", "", "bt addr")
	var status = flag.Bool("status", true, "Request status")
	var scanp = flag.Bool("scan", false, "Perform scan")
	var on = flag.Bool("on", false, "Turn on")
	var off = flag.Bool("off", false, "Turn off")
	var temp = flag.Bool("temp", false, "Set target temperature")
	var gate = flag.Bool("gate", false, "Set gate position(indoor|outdoor|mixed)")
	var value = flag.String("val", "", "Value")
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
	if *temp {
		v, err := strconv.Atoi(*value)
		if err != nil {
			log.Println(err)
			return
		}
		deviceCall(*device, func(t *tion.Tion) {
			s := t.Status()
			s.TempTarget = byte(v)
			err := t.Update(s)
			if err != nil {
				log.Println(err)
			}
		}, fmt.Sprintf("Target temperature updated to %d", v))
		return
	}

	if *gate {
		deviceCall(*device, func(t *tion.Tion) {
			s := t.Status()
			s.SetGateStatus(*value)
			err := t.Update(s)
			if err != nil {
				log.Println(err)
			}
		}, fmt.Sprintf("TGate set to %s", *value))
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
		log.Printf("Status: %v, Heater: %v, Sound: %v\n", state.Enabled, state.HeaterEnabled, state.SoundEnabled)
		log.Printf("Target: %v, In: %v, Out: %v\n", state.TempTarget, state.TempIn, state.TempOut)
		log.Printf("Gate: %s, Error: %d\n", state.GateStatus(), state.ErrorCode)
	}
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
