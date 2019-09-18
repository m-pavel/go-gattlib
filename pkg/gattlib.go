package gattlib

// #cgo CFLAGS: -g -Wall
// #cgo LDFLAGS: -lgattlib
// #include <gattlib.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

type Gatt struct {
	conn *C.gatt_connection_t
}

type UUID struct {
	uuid C.uuid_t
}

func (g Gatt) Connected() bool {
	return g.conn != nil
}

func (g *Gatt) Connect(addr string) error {
	str := C.CString(strings.ToUpper(addr))
	defer C.free(unsafe.Pointer(str))
	var err error
	g.conn, err = C.gattlib_connect(nil, str, C.GATTLIB_CONNECTION_OPTIONS_LEGACY_BDADDR_LE_PUBLIC|C.GATTLIB_CONNECTION_OPTIONS_LEGACY_BT_SEC_LOW)
	if g.conn == nil {
		return err
	}
	return nil
}

func (g *Gatt) Disconnect() error {
	if g.conn != nil {
		res := C.gattlib_disconnect(g.conn)
		g.conn = nil
		if res != 0 {
			return GattError(int(res))
		}
	}
	return nil
}

func (g *Gatt) Read(uuid string) ([]byte, int, error) {
	var n C.size_t
	uuidS, err := g.uUID(uuid)
	if err != nil {
		return nil, 0, err
	}
	var ptr unsafe.Pointer
	res := C.gattlib_read_char_by_uuid(g.conn, &uuidS.uuid, &ptr, &n)
	if res != 0 {
		return nil, 0, GattError(int(res))
	}

	buffer := C.GoBytes(ptr, C.int(n))
	fmt.Println(n)
	return buffer, int(n), nil
}

func (g *Gatt) Write(uuid string, bf []byte) error {
	uuidS, err := g.uUID(uuid)
	if err != nil {
		return err
	}

	res := C.gattlib_write_char_by_uuid(g.conn, &uuidS.uuid, unsafe.Pointer(&bf[0]), C.size_t(len(bf)))
	if res != 0 {
		return GattError(int(res))
	}
	return nil
}

func (g *Gatt) uUID(uuid string) (*UUID, error) {
	var rUUID UUID
	cuuid := C.CString(uuid)
	defer C.free(unsafe.Pointer(cuuid))
	res := C.gattlib_string_to_uuid(cuuid, C.size_t(len(uuid)+1), &rUUID.uuid)
	if res != 0 {
		return nil, GattError(int(res))
	} else {
		return &rUUID, nil
	}
}
