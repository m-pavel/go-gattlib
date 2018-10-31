package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/m-pavel/go-gattlib/tion"

	"fmt"

	"github.com/gorhill/cronexpr"
	"github.com/sevlyar/go-daemon"
)

var (
	stop = make(chan struct{})
	done = make(chan struct{})
)

func main() {
	var logf = flag.String("log", "schedule.log", "log")
	var pid = flag.String("pid", "schedule.pid", "pid")
	var notdaemonize = flag.Bool("n", false, "Do not do to background.")
	var signal = flag.String("s", "", `send signal to the daemon stop — shutdown`)
	var db = flag.String("db", "schedule.db", "Schedule db")
	var device = flag.String("device", "xx:yy:zz:aa:bb:cc", "Device BT address")
	var action = flag.String("action", "daemon", "daemon|prepare|list|add|del")
	var value = flag.String("value", "", "schedule or id")
	var act = flag.String("act", "", "act")
	//var interval = flag.Int("interval", 10, "Interval secons")
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	dao, err := New(*db)
	if err != nil {
		log.Println(err)
		stop <- struct{}{}
		return
	}
	defer dao.Close()

	if *action == "prepare" {
		err = dao.Prepare()
		if err != nil {
			log.Println(err)
		}
		return
	}
	if *action == "add" {
		if *act == "" {
			log.Println("Act is required")
			return
		}
		_, err := cronexpr.Parse(*value)
		if err != nil {
			log.Printf("Wrong cron expreassion: %v", err)
			return
		}
		err = dao.Add(*value, *act)
		if err != nil {
			log.Println(err)
		}
		return
	}
	if *action == "del" {
		err = dao.Delete(*value)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if *action == "list" {
		s, err := dao.GetSchedules()
		if err != nil {
			log.Println(err)
		}
		for _, sch := range s {
			fmt.Printf("%d | %s | %s\n", sch.Id, sch.Value, sch.Action)
		}
		return
	}

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

	daemonf(*device, dao)

}

func daemonf(device string, dao *Dao) {
	sch, err := dao.GetSchedules()
	if err != nil {
		log.Println(err)
		return
	}
	for i, _ := range sch {
		go func(s Schedule) {
			for {
				expr := cronexpr.MustParse(s.Value).Next(time.Now())
				log.Printf("Next time for %d (%s) %s\n", s.Id, s.Action, expr.Format("Mon Jan _2 15:04:05 2006"))
				time.Sleep(expr.Sub(time.Now()))
				log.Printf("Executing %d\n", s.Id)
				t := tion.New(device)
				err = t.Connect()
				if err != nil {
					log.Println(err)
				} else {
					if s.Action == "on" {
						err := t.On()
						if err != nil {
							log.Println(err)
						}
					}
					if s.Action == "off" {
						t.Off()
						if err != nil {
							log.Println(err)
						}
					}
					t.Disconnect()
				}

			}
		}(sch[i])
	}
	done <- struct{}{}
}

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}
