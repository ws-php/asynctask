package main

import (
	"io/ioutil"
	"strings"
	"time"
)

type Worker struct {
	Id int

	run bool

	task chan *Task
	s    *Scheduler

	LastTime time.Time
}

func (w *Worker) Init(id int, s *Scheduler) *Worker {
	w.Id = id
	w.s = s
	w.task = make(chan *Task)
	return w
}

func (w *Worker) doHttp(t *Task) (status int, msg string) {

	resp, err := w.s.e.Client.Post(
		w.s.e.BaseUrl+"/"+t.job.Name,
		"application/x-www-form-urlencoded",
		strings.NewReader(t.Content),
	)

	if err != nil {
		status = -1
		msg = err.Error()
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	status = resp.StatusCode
	msg = string(body)

	return
}

func (w *Worker) log(t *Task) {
	w.s.e.Info.Printf(
		"%d %s %0.3fs %0.3fs %d %s %s\n",
		t.Id,
		t.job.Name,
		t.StartTime.Sub(t.AddTime).Seconds(),
		t.EndTime.Sub(t.StartTime).Seconds(),
		t.Status,
		t.Content,
		t.Msg,
	)
}

func (w *Worker) Run() {
	for t := range w.task {
		t.StartTime = time.Now()
		t.Status, t.Msg = w.doHttp(t)
		t.EndTime = time.Now()

		w.log(t)

		w.s.complete <- t
	}
}

func (w *Worker) Close() {
	close(w.task)
}
