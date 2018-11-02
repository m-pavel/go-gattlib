package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/m-pavel/go-gattlib/tion"

	"fmt"

	"github.com/go-errors/errors"
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

	var prepare = flag.Bool("prepare", false, "Prepare database")

	var list = flag.Bool("list", false, "list")
	var del = flag.Int("del", -1, "del")
	var on = flag.Bool("on", false, "On")
	var off = flag.Bool("off", false, "Off")
	var schedule = flag.String("schedule", "", "Add schedule")
	var temp = flag.Int("temp", -1, "Temperature target")
	var heater = flag.String("heater", "", "Heater")
	var sound = flag.String("sound", "", "Sound")
	var gate = flag.String("gate", "", "indoor|mixed|outdoor")
	var speed = flag.Int("speed", -1, "speed")

	flag.Parse()
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	dao, err := New(*db)
	if err != nil {
		log.Println(err)
		stop <- struct{}{}
		return
	}
	defer dao.Close()

	if *prepare {
		err = dao.Prepare()
		if err != nil {
			log.Println(err)
		}
		return
	}

	if *list {
		s, err := dao.GetSchedules()
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("ID | SCHEDULE | ENABLED | HEATER | SOUND | TEMP | SPEED | GATE\n")

		fi := func(v *int) string {
			if v == nil {
				return "n/a"
			}
			return fmt.Sprintf("%d", *v)
		}
		fg := func(v *int) string {
			if v == nil {
				return "n/a"
			}
			return tion.GateStatus(byte(*v))
		}
		for _, sch := range s {
			fmt.Printf("%d | %s | %s | %s | %s | %v | %s | %s |\n", sch.Id, sch.Value, fb(sch.Enabled), fb(sch.Heater), fb(sch.Sound), fi(sch.Temp), fi(sch.Speed), fg(sch.Gate))
		}
		return
	}
	if *del != -1 {
		err = dao.Delete(*del)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if *schedule != "" {
		var enb, htr, snd *bool
		var true_, false_ bool
		true_ = true
		false_ = false
		if *heater == "on" {
			htr = &true_
		}
		if *heater == "off" {
			htr = &false_
		}
		if *sound == "on" {
			snd = &true_
		}
		if *sound == "off" {
			snd = &false_
		}
		if *on {
			enb = &true_
		}
		if *off {
			enb = &false_
		}
		var gt *int
		if *gate != "" {
			s := tion.Status{}
			s.SetGateStatus(*gate)
			iv := int(s.Gate)
			gt = &iv
		}
		if *temp == -1 {
			temp = nil
		}
		if *speed == -1 {
			speed = nil
		}

		err = dao.Add(*schedule, enb, htr, snd, gt, speed, temp)
		if err != nil {
			log.Println(err)
		}
		return
	}

	log.Println("Running daemon")
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

func fb(v *bool) string {
	if v == nil {
		return "n/a"
	}
	if *v {
		return "on"
	}
	return "off"
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
				mins := expr.Sub(time.Now()) / time.Minute
				log.Printf("Next time for %d (%s) is %s in %d minute(s).\n", s.Id, fb(s.Enabled), expr.Format("Mon Jan _2 15:04:05 2006"), mins)
				time.Sleep(expr.Sub(time.Now()))
				log.Printf("Executing %d\n", s.Id)
				err := execute(s, device)
				if err != nil {
					log.Println(err)
				}
			}
		}(sch[i])
	}
	done <- struct{}{}
}

func execute(s Schedule, device string) error {
	t := tion.New(device)
	err := t.Connect()
	if err != nil {
		return err
	}
	defer t.Disconnect()
	ts := t.Status()
	if ts == nil {
		return errors.New("Status is nil")
	}

	if s.Enabled != nil {
		ts.Enabled = *s.Enabled
	}
	if s.Gate != nil {
		ts.Gate = byte(*s.Gate)
	}
	if s.Temp != nil {
		ts.TempTarget = byte(*s.Temp)
	}
	if s.Speed != nil {
		ts.Speed = byte(*s.Speed)
	}
	if s.Heater != nil {
		ts.HeaterEnabled = *s.Heater
	}
	if s.Sound != nil {
		ts.SoundEnabled = *s.Sound
	}
	log.Printf("Device request %v\n", ts)
	err = t.Update(ts)
	if err != nil {
		return err
	}
	return nil
}

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	if sig == syscall.SIGQUIT {
		<-done
	}
	return daemon.ErrStop
}
