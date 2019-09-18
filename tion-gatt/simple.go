package tion_gatt

import (
	"log"
	"time"

	"github.com/go-errors/errors"
	"github.com/m-pavel/go-gattlib/pkg"
	tion2 "github.com/m-pavel/go-gattlib/tion"
)

const (
	wchar = "6e400002-b5a3-f393-e0a9-e50e24dcca9e"
	rchar = "6e400003-b5a3-f393-e0a9-e50e24dcca9e"
)

type tion struct {
	g     *gattlib.Gatt
	Addr  string
	debug bool
}

func New(addr string, debug ...bool) tion2.Tion {
	t := tion{Addr: addr, g: &gattlib.Gatt{}}
	if len(debug) == 1 && debug[0] {
		t.debug = true
	}
	return &t
}

type cRes struct {
	s *tion2.Status
	e error
}

func (t *tion) ReadState(timeout int) (*tion2.Status, error) {
	if t.g.Connected() {
		return nil, errors.New("Already connected")
	}

	c1 := make(chan cRes, 1)

	go func() {
		err := t.g.Connect(t.Addr)
		if err != nil {
			c1 <- cRes{e: err}
			return
		}
		defer t.g.Disconnect()

		r, e := t.rw()
		c1 <- cRes{e: e, s: r}
	}()

	select {
	case res := <-c1:
		return res.s, res.e
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil, errors.New("Read timeout")
	}
}

func (t *tion) rw() (*tion2.Status, error) {
	if !t.g.Connected() {
		return nil, errors.New("Not connected")
	}
	err := t.g.Write(wchar, tion2.StatusRequest)
	time.Sleep(time.Second)
	resp, _, err := t.g.Read(rchar)
	if err != nil {
		return nil, err
	}
	if t.debug {
		log.Printf("RSP: %v\n", resp)
	}
	return tion2.FromBytes(resp)
}

func (t *tion) Update(s *tion2.Status) error {
	return t.g.Write(wchar, tion2.FromStatus(s))
}
