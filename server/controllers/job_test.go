package controllers

import (
	"context"
	"cron-server/server/domains"
	"cron-server/server/dtos"
	"cron-server/server/migrations"
	"cron-server/server/misc"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

var (
	jobController = JobController{}
	inboundJob    dtos.JobDto
	jobModel      domains.JobDomain
	project       domains.ProjectDomain
)

func TestJobController_CreateOne(t *testing.T) {
	var jobsPool, err = migrations.NewPool(migrations.CreateConnection, 1)
	misc.CheckErr(err)
	jobController.Pool = *jobsPool

	t.Log("Respond with status 400 if request body does not contain required values")
	{
		inboundJob.CronSpec = "* * * * *"
		jobByte, err := inboundJob.ToJson()
		misc.CheckErr(err)
		jobStr := string(jobByte)

		req, err := http.NewRequest("POST", "/jobs", strings.NewReader(jobStr))
		if err != nil {
			t.Fatalf("\t\t Cannot create http request")
		}

		w := httptest.NewRecorder()
		jobController.CreateOne(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	t.Log("Respond with status 201 if request body is valid")
	{
		project.Name = "TestJobController_Project"
		project.Description = "TestJobController_Project_Description"
		id, err := project.CreateOne(&jobController.Pool, context.Background())

		if err != nil {
			t.Fatalf("\t\t Cannot create project %v", err)
		}

		project.ID = id
		j1 := dtos.JobDto{}

		j1.CronSpec = "1 * * * *"
		j1.ProjectId = id
		j1.CallbackUrl = "http://random.url"
		j1.StartDate = time.Now().Add(60 * time.Second).UTC().Format(time.RFC3339)
		jobByte, err := j1.ToJson()
		misc.CheckErr(err)
		jobStr := string(jobByte)
		req, err := http.NewRequest("POST", "/jobs", strings.NewReader(jobStr))
		if err != nil {
			t.Fatalf("\t\t Cannot create job %v", err)
		}

		w := httptest.NewRecorder()
		jobController.CreateOne(w, req)
		body, err := ioutil.ReadAll(w.Body)

		if err != nil {
			t.Fatalf("\t\t Could not read response body %v", err)
		}

		var response map[string]interface{}

		if err = json.Unmarshal(body, &response); err != nil {
			t.Fatalf("\t\t Could unmarsha json response %v", err)
		}

		if len(response) < 1 {
			t.Fatalf("\t\t Response payload is empty")
		}

		fmt.Println(response)

		inboundJob.ID = response["data"].(string)
		assert.Equal(t, http.StatusCreated, w.Code)
	}
}

func TestJobController_GetAll(t *testing.T) {
	t.Log("Respond with status 200 and return all created jobs")
	{
		jobModel.ProjectId = project.ID
		jobModel.CronSpec = "1 * * * *"
		jobModel.StartDate = time.Now().Add(60 * time.Second)
		jobModel.CallbackUrl = "some-url"

		rv := reflect.ValueOf(jobModel)
		rt := rv.Type()
		rc := reflect.New(rt)
		rc.Elem().Set(rv)

		jobTwoCopy := rc.Interface().(*domains.JobDomain)

		if _, err := jobModel.CreateOne(&jobController.Pool, context.Background()); err != nil {
			t.Fatalf("\t\t Cannot create job two %v", err)
		}

		if _, err := jobTwoCopy.CreateOne(&jobController.Pool, context.Background()); err != nil {
			t.Fatalf("\t\t Cannot create job three %v", err)
		}

		req, err := http.NewRequest("GET", "/jobs?project_id="+project.ID, nil)

		if err != nil {
			t.Fatalf("\t\t Cannot create http request %v", err)
		}

		w := httptest.NewRecorder()
		jobController.GetAll(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("\t\t Error reading data %v", err)
		}

		fmt.Println(string(body))

		if _, err := jobModel.DeleteOne(&jobController.Pool, context.Background()); err != nil {
			t.Fatalf("\t\t Cannot delete job two %v", err)
		}

		if _, err := jobTwoCopy.DeleteOne(&jobController.Pool, context.Background()); err != nil {
			t.Fatalf("\t\t Cannot delete job two copy %v", err)
		}
	}
}

func TestJobController_UpdateOne(t *testing.T) {
	t.Log("Respond with status 400 if update attempts to change cron spec")
	{
		inboundJob.CronSpec = "3 * * * *"
		jobByte, err := inboundJob.ToJson()
		misc.CheckErr(err)
		jobStr := string(jobByte)
		req, err := http.NewRequest("PUT", "/jobs/"+inboundJob.ID, strings.NewReader(jobStr))
		if err != nil {
			t.Fatalf("\t\t Cannot create http request %v", err)
		}

		w := httptest.NewRecorder()
		jobController.UpdateOne(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	t.Log("Respond with status 200 if update body is valid")
	{
		inboundJob.StartDate = time.Now().UTC().Format(time.RFC3339)
		inboundJob.CronSpec = "1 * * * *"
		inboundJob.Description = "some job description"
		inboundJob.Timezone = "UTC"
		jobByte, err := inboundJob.ToJson()
		misc.CheckErr(err)
		jobStr := string(jobByte)
		req, err := http.NewRequest("PUT", "/jobs/"+inboundJob.ID, strings.NewReader(jobStr))

		if err != nil {
			t.Fatalf("\t\t Cannot create http request %v", err)
		}

		w := httptest.NewRecorder()
		jobController.UpdateOne(w, req)
		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("\t\t Cannot create http request %v", err)
			log.Println("Response body :", string(body))
		}

		assert.Equal(t, http.StatusOK, w.Code)
		log.Println("Response body :", string(body))
	}
}

func TestJobController_DeleteOne(t *testing.T) {
	t.Log("Respond with status 200 after successful deletion")
	{
		req, err := http.NewRequest("DELETE", "/jobs/"+inboundJob.ID, nil)
		if err != nil {
			t.Fatalf("\t\t Cannot create http request %v", err)
		}

		w := httptest.NewRecorder()
		jobController.DeleteOne(w, req)
		assert.Equal(t, w.Code, http.StatusOK)

		if _, err = project.DeleteOne(&jobController.Pool, context.Background()); err != nil {
			t.Fatalf("\t\t Cannot delete project %v", err)
		}
	}

	jobController.Pool.Close()
}
