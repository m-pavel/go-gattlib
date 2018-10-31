package main

import (
	"database/sql"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Dao struct {
	db *sql.DB
}

type Schedule struct {
	Id     int
	Value  string
	Action string
}

func New(db string) (*Dao, error) {
	dao := Dao{}
	var err error
	dao.db, err = sql.Open("sqlite3", db)
	if err != nil {
		return nil, err
	}
	return &dao, nil
}

func (d *Dao) Prepare() error {
	_, err := d.db.Exec("CREATE TABLE SCHEDULES (SCHEDULE text NOT NULL, ACTION test NOT NULL)")
	return err
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
		d.db = nil
	}
}

func (d *Dao) GetSchedules() ([]Schedule, error) {
	stmt, err := d.db.Prepare("SELECT ROWID, SCHEDULE, ACTION FROM SCHEDULES")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sch := make([]Schedule, 0)
	for {
		if !rows.Next() {
			break
		}
		var s Schedule
		rows.Scan(&s.Id, &s.Value, &s.Action)
		if err != nil {
			return nil, err
		}
		sch = append(sch, s)
	}
	return sch, err
}

func (d *Dao) Add(schedule, action string) error {
	_, err := d.db.Exec("INSERT INTO SCHEDULES VALUES (?, ?)", schedule, action)
	return err
}

func (d *Dao) Delete(id string) error {
	iid, _ := strconv.Atoi(id)
	_, err := d.db.Exec("DELETE FROM SCHEDULES WHERE ROWID=?", iid)
	return err
}
