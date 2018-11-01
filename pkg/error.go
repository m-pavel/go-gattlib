package gattlib

import "fmt"

type GattErr struct {
	Id int
}

func (g GattErr) Error() string {
	return fmt.Sprintf("GATT Error %d", g.Id)
}

func GattError(id int) *GattErr {
	return &GattErr{Id: id}
}
