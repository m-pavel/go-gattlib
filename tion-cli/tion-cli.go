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
	//var value = flag.String("value", "", "schedule or id")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

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
		err := deviceCall(*device, func(t *tion.Tion) { t.Off() })
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
		log.Printf("Staus: %v Heater: %v Sound: %v\n", state.Enabled, state.HeaterEnabled, state.SoundEnabled)
		log.Printf("Target: %v In: %v Out: %v\n", state.TempTarget, state.TempIn, state.TempOut)
		log.Printf("Gate: %d, Error: %d\n", state.Gate, state.ErrorCode)
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
	log.Println("Scan")
	gattlib.Scan(func(ad ble.Advertisement) {
		log.Printf("%s %s", ad.Addr(), ad.LocalName())
	}, 5)
	time.Sleep(10 * time.Second)
}
