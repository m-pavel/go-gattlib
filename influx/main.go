package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"strconv"

	"fmt"

	"bitbucket.org/autogrowsystems/go-sdk/util/tell"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/m-pavel/go-gattlib/tion"
	"github.com/sevlyar/go-daemon"
)

func main() {
	var logf = flag.String("log", "influxexport.log", "log")
	var pid = flag.String("pid", "influxexport.pid", "pid")
	var notdaemonize = flag.Bool("n", false, "Do not do to background.")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var iserver = flag.String("influx", "http://localhost:8086", "Influx DB endpoint")
	var nserver = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")
	var interval = flag.Int("interval", 10, "Interval secons")
	flag.Parse()
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGTERM, termHandler)

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

	daemonf(*iserver, *nserver, *interval)

}

func daemonf(iserver, device string, interval int) {
	var err error
	cli, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: iserver,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	tn := tion.New(device)
	err = tn.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer tn.Disconnect()

	tn.RegisterHandler(func(s *tion.Status) {
		point, err := client.NewPoint("tion",
			map[string]string{
				"gate":   strconv.Itoa(int(s.Gate)),
				"on":     fmt.Sprintf("%v", s.Enabled),
				"heater": fmt.Sprintf("%v", s.HeaterEnabled),
			},
			map[string]interface{}{
				"out": s.TempIn,
				"in":  s.TempOut,
				"tgt": s.TempTarget,
				"spd": s.Speed,
			},
			time.Now())
		if err != nil {
			log.Println(err)
			return
		}
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  "tion",
			Precision: "s",
		})
		if err != nil {
			log.Println("Insert data error:")
		}
		bp.AddPoint(point)
		err = cli.Write(bp)
		if err != nil {
			log.Println("Insert data error:")
		}
	})

	done <- struct{}{}
}

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func termHandler(sig os.Signal) error {
	tell.Info("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}
