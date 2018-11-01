package main

import (
	"flag"
	"log"

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
		t := tion.New(*device)
		err := t.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer t.Disconnect()
		err = t.On()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Turned on")
		break
	case "off":
		t := tion.New(*device)
		err := t.Connect()
		if err != nil {
			log.Println(err)
			return
		}
		defer t.Disconnect()
		err = t.Off()
		if err != nil {
			log.Println(err)
			return
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
		log.Println(state)
		break
	}
}
