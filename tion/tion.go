package tion

import (
	"log"
	"time"

	"github.com/go-errors/errors"
	"github.com/m-pavel/go-gattlib/pkg"
)

const (
	wchar = "6e400002-b5a3-f393-e0a9-e50e24dcca9e"
	rchar = "6e400003-b5a3-f393-e0a9-e50e24dcca9e"
)

type StatusHandler func(*Status)

type Tion struct {
	g        *gattlib.Gatt
	Addr     string
	sc       chan int
	ls       *Status
	sh       StatusHandler
	interval int
}

func New(addr string, interval ...int) *Tion {
	t := Tion{Addr: addr, g: &gattlib.Gatt{}, interval: 10}
	if len(interval) == 1 && interval[0] > 4 {
		t.interval = interval[0]
	}
	return &t
}

func (t *Tion) Connect() error {
	if t.Connected() {
		return errors.New("Already connected")
	}
	err := t.g.Connect(t.Addr)
	if err != nil {
		return err
	}
	t.startStatusLoop()
	return nil
}

func (t *Tion) Disconnect() error {
	t.stopStatusLoop()
	if t.g != nil {
		return t.g.Disconnect()
		t.g = nil
	}
	return nil
}

// ReadState witout keeping connection open
// Must be not connected before execution
func (t *Tion) ReadState(timeout ...int) (*Status, error) {
	if t.g.Connected() {
		return nil, errors.New("Already connected")
	}
	if t.sc != nil {
		return nil, errors.New("Loop already running")
	}
	err := t.g.Connect(t.Addr)
	if err != nil {
		return nil, err
	}
	defer t.g.Disconnect()
	return t.rw()
}

func (t *Tion) RegisterHandler(h StatusHandler) {
	t.sh = h
}

func (t *Tion) startStatusLoop() {
	if t.sc != nil {
		log.Println("Already running")
		return
	}
	ticker := time.NewTicker(time.Duration(t.interval) * time.Second)
	t.sc = make(chan int)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Tick")
				s, err := t.rw()
				if err == nil {
					t.ls = s
					if t.sh != nil {
						t.sh(s)
					}
				} else {
					log.Println(err)
				}
			case <-t.sc:
				ticker.Stop()
				return
			}
		}
	}()
}

func (t *Tion) rw() (*Status, error) {
	if !t.Connected() {
		return nil, errors.New("Not connected")
	}
	err := t.g.Write(wchar, statusRequest)
	time.Sleep(time.Second)
	resp, _, err := t.g.Read(rchar)
	if err != nil {
		return nil, err
	}
	return FromBytes(resp)
}

func (t *Tion) stopStatusLoop() {
	if t.sc == nil {
		log.Println("Not running")
		return
	}
	t.sc <- 1
	close(t.sc)
	t.sc = nil
}

func (t *Tion) Status() Status {
	return *t.ls
}

func (t Tion) Connected() bool {
	return t.g != nil && t.g.Connected()
}

func (t *Tion) On() error {
	if t.ls == nil {
		return errors.New("Current state not retrieved yet")
	}
	rq := BuildRequest(true, t.ls.SoundEnabled, t.ls.HeaterEnabled, t.ls.Speed, t.ls.Gate, t.ls.TempTarget)
	return t.g.Write(wchar, rq)
}

func (t *Tion) Off() error {
	if t.ls == nil {
		return errors.New("Current state not retrieved yet")
	}
	rq := BuildRequest(false, t.ls.SoundEnabled, t.ls.HeaterEnabled, t.ls.Speed, t.ls.Gate, t.ls.TempTarget)
	return t.g.Write(wchar, rq)

}
