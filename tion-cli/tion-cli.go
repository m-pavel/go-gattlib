package main

import (
	"flag"
	"log"
	"time"

	"github.com/go-ble/ble"

	"github.com/m-pavel/go-gattlib/pkg"

	"github.com/m-pavel/go-gattlib/tion"
)

func main() {
	var device = flag.String("device", "", "bt addr")
	var action = flag.String("action", "daemon", "scan|on|off|status")
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	if *device == "" && *action != "scan" {
		log.Fatal("Device address is mandatory")
	}
	switch *action {
	case "on":
		err := deviceCall(*device, func(t *tion.Tion) { t.On() })
		if err != nil {
			log.Println(err)
		}
		log.Println("Turned on")
		break
	case "off":
		err := deviceCall(*device, func(t *tion.Tion) {
			err := t.Off()
			if err != nil {
				log.Println(err)
			}
		})
		if err != nil {
			log.Println(err)
		}
		log.Println("Turned off")
		break
	case "":
	case "status":
		state, err := tion.New(*device).ReadState()
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Status: %v, Heater: %v, Sound: %v\n", state.Enabled, state.HeaterEnabled, state.SoundEnabled)
		log.Printf("Target: %v, In: %v, Out: %v\n", state.TempTarget, state.TempIn, state.TempOut)
		log.Printf("Gate: %s, Error: %d\n", state.GateStatus(), state.ErrorCode)
		break
	case "scan":
		scan()
		return
	}
}

func deviceCall(addr string, cb func(*tion.Tion)) error {
	t := tion.New(addr)
	err := t.Connect()
	if err != nil {
		return err
	}
	defer t.Disconnect()
	cb(t)
	if err != nil {
		return err
	}
	return nil
}

func scan() {
	gattlib.Scan(func(ad ble.Advertisement) {
		log.Printf("%s %s", ad.Addr(), ad.LocalName())
	}, 5)
	time.Sleep(10 * time.Second)
}
