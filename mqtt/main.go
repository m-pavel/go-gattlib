package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	_ "net/http"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/m-pavel/go-gattlib/tion"
	"github.com/sevlyar/go-daemon"
)

var (
	stop = make(chan struct{})
	done = make(chan struct{})
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

func main() {
	var logf = flag.String("log", "mqtttion.log", "log")
	var pid = flag.String("pid", "mqtttion.pid", "pid")
	var notdaemonize = flag.Bool("n", false, "Do not do to background.")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var mqtt = flag.String("mqtt", "tcp://localhost:1883", "MQTT endpoint")
	var topic = flag.String("t", "nn/tion", "MQTT topic")
	var topicctrl = flag.String("tc", "nn/tioncontrol", "MQTT control topic")
	var user = flag.String("mqtt-user", "", "MQTT user")
	var pass = flag.String("mqtt-pass", "", "MQTT password")
	var device = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")
	var interval = flag.Int("interval", 30, "Interval secons")
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	cntxt := &daemon.Context{
		PidFileName: *pid,
		PidFilePerm: 0644,
		LogFileName: *logf,
		LogFilePerm: 0640,
		WorkDir:     "/tmp",
		Umask:       027,
		Args:        os.Args,
	}

	if !*notdaemonize && len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %v", err)
		}
		daemon.SendCommands(d)
		return
	}

	if !*notdaemonize {
		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal(err)
		}
		if d != nil {
			return
		}
	}

	daemonf(*mqtt, *topic, *topicctrl, *user, *pass, *device, *interval)

}

type mqttCli struct {
	tion *tion.Tion
	mqtt MQTT.Client

	bt       string
	topic    string
	topicc   string
	user     string
	password string
	mqtturl  string
	interval int
}

func (mc *mqttCli) control(cli MQTT.Client, msg MQTT.Message) {
	req := Request{}
	err := json.Unmarshal(msg.Payload(), &req)
	if err != nil {
		log.Println(err)
		return
	}

	cs := mc.tion.Status()
	if cs.Enabled && !req.On {
		cs.Enabled = false
		err = mc.tion.Update(cs)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Turned off by MQTT request")
		}
	}
	if !cs.Enabled && req.On {
		cs.Enabled = true
		err := mc.tion.Update(cs)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Turned on  by MQTT request")
		}
	}
}

func (mc *mqttCli) Connect() error {
	opts := MQTT.NewClientOptions().AddBroker(mc.mqtturl)
	opts.SetClientID("temper-go-cli")
	if mc.user != "" {
		opts.Username = mc.user
		opts.Password = mc.password
	}

	mc.mqtt = MQTT.NewClient(opts)
	if token := mc.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	mc.mqtt.Subscribe(mc.topicc, 0, mc.control)
	mc.tion = tion.New(mc.bt)
	return nil
}

func (mc *mqttCli) Loop() {
	erinr := 0
	for {
		select {
		case <-stop:
			log.Println("Exiting")
			break
		case <-time.After(time.Duration(mc.interval) * time.Second):
			s, err := mc.tion.ReadState(7)
			if err != nil {
				log.Println(err)
				erinr++
			} else {
				mc.reportMqtt(s)
				erinr = 0
			}
			if erinr == 10 {
				return
			}
		}
	}
}

func (mc *mqttCli) reportMqtt(s *tion.Status) {
	req := Request{
		Gate:   s.GateStatus(),
		On:     s.Enabled,
		Heater: s.HeaterEnabled,
		Out:    s.TempOut,
		In:     s.TempIn,
		Target: s.TempTarget,
		Speed:  s.Speed,
		Sound:  s.SoundEnabled,
	}

	bp, err := json.Marshal(&req)
	if err != nil {
		log.Println(err)
		return
	}
	tkn := mc.mqtt.Publish(mc.topic, 0, false, bp)
	if tkn.Error() != nil {
		log.Println(tkn.Error())
	}
}

func (mc *mqttCli) Close() error {
	mc.mqtt.Disconnect(3000)
	return nil
}

func daemonf(mqtt, topic, topicc string, u, p string, device string, interval int) {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	mq := mqttCli{bt: device, topicc: topicc, topic: topic, user: u, password: p, interval: interval, mqtturl: mqtt}
	err := mq.Connect()
	if err != nil {
		panic(err)
	}
	mq.Loop()
	done <- struct{}{}
}

func termHandler(sig os.Signal) error {
	log.Println("Terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}
