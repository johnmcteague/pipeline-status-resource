package state

import (
	"strconv"
	"time"

	"github.com/pivotalservices/pipeline-status-resource/models"
)

func ChangeState(status *models.PipelineStatus,
	buildState models.PipelineState,
	failure *models.BuildFailure) (newStatus *models.PipelineStatus, err error) {

	newStatus = &models.PipelineStatus{}
	*newStatus = *status

	switch {
	case buildState == models.StateRunning && newStatus.State != models.StateRunning:
		buildNum, _ := strconv.Atoi(status.BuildNumber)

		newStatus.State = buildState
		newStatus.BuildNumber = strconv.Itoa(buildNum + 1)
		newStatus.Failure = nil
		modifyStatus(newStatus)
	case buildState == models.StateReady && newStatus.State != models.StateReady:
		newStatus.State = buildState
		newStatus.Failure = failure
		modifyStatus(newStatus)
	}

	return
}

func modifyStatus(s *models.PipelineStatus) {
	now := time.Now()
	s.LastModified = now.Format("2006-01-02T15:04:05-0700")
}
