package controllers

import (
	"cron-server/job"
	"cron-server/repo"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
)

func RegisterJob(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	j := job.FromJson(body)

	if err != nil {
		panic(err)
	}

	if len(j.ServiceName) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Service name is required"))
		return
	}

	if len(j.CronSpec) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cron is required"))
		return
	}

	if j.StartDate.IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cron is required"))
		return
	}

	jd := j.ToDomain()

	schedule, err := cron.ParseStandard(jd.CronSpec)
	if err != nil {
		panic(err)
	}

	jd.State = job.InActiveJob
	jd.NextTime = schedule.Next(jd.StartDate)
	jd.TotalExecs = -1
	jd.MinsBetweenExecs = jd.NextTime.Sub(jd.StartDate).Minutes()

	newJD, err := repo.CreateOne(jd)
	if err != nil {
		panic(err)
	}

	log.Println("Registered ", jd)

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Location", "jobs/"+newJD.ID)
	w.WriteHeader(http.StatusCreated)
}
