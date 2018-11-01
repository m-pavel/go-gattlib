package gattlib

import (
	"context"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

func Scan(cbk ble.AdvHandler, timeout int) {
	d, err := dev.NewDevice("default")
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second))
	err = ble.Scan(ctx, false, cbk, nil)
	if err != nil {
		log.Fatal(err)
	}
}
