package gattlib

import "C"

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lgattlib
// #include <stdlib.h>
// #include <gattlib.h>
import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"unsafe"
)

type Gatt struct {
	conn *C.gatt_connection_t
}

type UUID struct {
	uuid C.uuid_t
}

func (g *Gatt) Connect(addr string) error {
	str := C.CString(addr)
	g.conn = C.gattlib_connect(nil, str, C.BDADDR_LE_PUBLIC, C.BT_SEC_LOW, 0, 0)
	if g.conn == nil {
		return errors.New("Unable to connect")
	}
	return nil
}

func (g *Gatt) Disconnect() error {
	if g.conn != nil {
		res := C.gattlib_disconnect(g.conn)
		g.conn = nil
		if res != 0 {
			return errors.New(fmt.Sprintf("Error %d", res))
		}
	}
	return nil
}

func (g *Gatt) Read(uuid string) ([]byte, int, error) {
	buffer := make([]byte, 100)
	var n C.size_t
	n = 100
	uuidS, err := g.uUID(uuid)
	if err != nil {
		return nil, 0, err
	}

	res := C.gattlib_read_char_by_uuid(g.conn, &uuidS.uuid, unsafe.Pointer(&buffer[0]), &n)
	if res != 0 {
		return nil, 0, errors.New(fmt.Sprintf("Error %d", res))
	}
	return buffer, int(n), nil
}

func (g *Gatt) Write(uuid string, bf []byte) error {
	uuidS, err := g.uUID(uuid)
	if err != nil {
		return err
	}

	res := C.gattlib_write_char_by_uuid(g.conn, &uuidS.uuid, unsafe.Pointer(&bf[0]), C.size_t(len(bf)))
	if res != 0 {
		return errors.New(fmt.Sprintf("Error %d", res))
	}
	return nil
}

func (g *Gatt) uUID(uuid string) (*UUID, error) {
	var res UUID
	ci := C.gattlib_string_to_uuid(C.CString(uuid), C.size_t(len(uuid)+1), &res.uuid)
	if ci != 0 {
		return nil, errors.New(strconv.Itoa(int(ci)))
	} else {
		return &res, nil
	}

}
