package scheduler

import (
	"os"
	"github.com/sidra-api/sidra/dto"
	"github.com/jasonlvhit/gocron"	
)

type Job struct {
	dataSet *dto.DataPlane
	controlPlaneHost string
}

func NewJob(dataSet *dto.DataPlane) *Job {
	os.Mkdir("/tmp/gs", 0755)
	controlPlaneHost := os.Getenv("CONTROL_PLANE_HOST")
	if os.Getenv("env") == "local" {
		controlPlaneHost = "http://host.docker.internal:8086"
	}
	if controlPlaneHost == "" {
		controlPlaneHost = "https://portal.sidra.id"
	}
	return &Job{		
		dataSet,
		controlPlaneHost,
	}
}

func (j *Job) Run() {
	gocron.Every(15).Second().Do(j.register())
	gocron.Every(15).Second().Do(j.storeConfig())	
	<-gocron.Start()
}