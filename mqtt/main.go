package main

import (
	"encoding/json"
	"flag"
	"log"
	_ "net/http"
	_ "net/http/pprof"
	"os"
	"syscall"
	"time"

	"net/http"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/m-pavel/go-gattlib/tion"
	"github.com/sevlyar/go-daemon"
)

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func main() {
	var logf = flag.String("log", "mqtttion.log", "log")
	var pid = flag.String("pid", "mqtttion.pid", "pid")
	var notdaemonize = flag.Bool("n", false, "Do not do to background.")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var mqtt = flag.String("mqtt", "tcp://localhost:1883", "MQTT endpoint")
	var topic = flag.String("t", "nn/tion", "MQTT topic")
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

	daemonf(*mqtt, *topic, *user, *pass, *device, *interval)

}

func daemonf(mqtt, topic string, u, p string, device string, interval int) {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	opts := MQTT.NewClientOptions().AddBroker(mqtt)
	opts.SetClientID("temper-go-cli")
	if u != "" {
		opts.Username = u
		opts.Password = p
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	t := tion.New(device)

	erinr := 0
	for {
		select {
		case <-stop:
			log.Println("Exiting")
			break
		case <-time.After(time.Duration(interval) * time.Second):
			s, err := t.ReadState(7)
			if err != nil {
				log.Println(err)
				erinr++
			} else {
				reportMqtt(client, topic, s)
				erinr = 0
			}
			if erinr == 10 {
				return
			}
		}
	}

	done <- struct{}{}
}

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

func reportMqtt(i MQTT.Client, t string, s *tion.Status) {
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
	tkn := i.Publish(t, 0, false, bp)
	if tkn.Error() != nil {
		log.Println(tkn.Error())
	}
}

func termHandler(sig os.Signal) error {
	log.Println("Terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}
