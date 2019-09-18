package main

import (
	"encoding/json"
	"flag"
	"log"
	_ "net/http"
	_ "net/http/pprof"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/m-pavel/go-gattlib/tion"
	"github.com/m-pavel/go-hassio-mqtt/pkg"
)

type Request struct {
	Gate   string `json:"gate"`
	On     bool   `json:"on"`
	Heater bool   `json:"heater"`
	Sound  bool   `json:"sound"`
	Out    int8   `json:"temp_out"`
	In     int8   `json:"temp_in"`
	Target int8   `json:"temp_target"`
	Speed  int8   `json:"speed"`
}

type TionService struct {
	t     *tion.Tion
	bt    *string
	debug bool
	ss    ghm.SendState
}

func (ts *TionService) PrepareCommandLineParams() {
	ts.bt = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")
}
func (ts TionService) Name() string { return "tion" }

func (ts *TionService) Init(client MQTT.Client, topic, topicc, topica string, debug bool, ss ghm.SendState) error {
	ts.t = tion.New(*ts.bt)
	if token := client.Subscribe(topicc, 0, ts.control); token.Error() != nil {
		return token.Error()
	}
	ts.debug = debug
	ts.ss = ss
	return nil
}

func (ts TionService) Do(client MQTT.Client) (interface{}, error) {
	s, err := ts.t.ReadState(7)
	if err != nil {
		return nil, err
	}

	return &Request{
		Gate:   s.GateStatus(),
		On:     s.Enabled,
		Heater: s.HeaterEnabled,
		Out:    s.TempOut,
		In:     s.TempIn,
		Target: s.TempTarget,
		Speed:  s.Speed,
		Sound:  s.SoundEnabled,
	}, nil
}

func (ts *TionService) control(cli MQTT.Client, msg MQTT.Message) {
	req := Request{}
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println(err)
		return
	}
	if ts.debug {
		log.Println(req)
	}

	cs, err := ts.t.ReadState(7)
	if err != nil {
		log.Println(err)
		return
	}
	if cs.Enabled && !req.On {
		cs.Enabled = false
		err = ts.t.Update(cs)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Turned off by MQTT request")
		}
	}
	if !cs.Enabled && req.On {
		cs.Enabled = true
		err = ts.t.Update(cs)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Turned on  by MQTT request")
		}
	}

	ts.ss()
}

func (ts TionService) Close() error {
	return ts.t.Disconnect()
}

func main() {
	ghm.NewStub(&TionService{}).Main()
}
