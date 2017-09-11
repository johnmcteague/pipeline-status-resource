package driver

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/adammck/venv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pivotalservices/pipeline-status-resource/models"
	"github.com/pivotalservices/pipeline-status-resource/state"
)

type Servicer interface {
	GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

type S3Driver struct {
	InitialVersion string

	Env                  venv.Env
	Svc                  Servicer
	BucketName           string
	Key                  string
	ServerSideEncryption string
}

func (driver *S3Driver) Start() (status *models.PipelineStatus, err error) {
	pipelineName := driver.Env.Getenv("BUILD_PIPELINE_NAME")
	teamName := driver.Env.Getenv("BUILD_TEAM_NAME")

	status = &models.PipelineStatus{}
	err = driver.load(status)
	if err == nil {
		if status.Pipeline != pipelineName {
			err = fmt.Errorf("State file is already associated with pipeline %s but is trying to be associated with pipeline %s",
				status.Pipeline, pipelineName)

			return nil, err
		}

		if status.Team != teamName {
			err = fmt.Errorf("State file is already associated with team %s but is trying to be associated with tea, %s",
				status.Team, teamName)

			return nil, err
		}
	} else if s3err, ok := err.(awserr.RequestFailure); ok && s3err.StatusCode() == 404 {
		status = &models.PipelineStatus{}
		status.Pipeline = pipelineName
		status.Team = teamName
		status.BuildNumber = driver.InitialVersion
	} else {
		status = nil
	}

	driver.changeAndPersistState(status, models.StateRunning, nil)
	return
}

func (driver *S3Driver) Finish() (status *models.PipelineStatus, err error) {
	return driver.makeReady(nil)
}

func (driver *S3Driver) Fail() (status *models.PipelineStatus, err error) {
	failure := &models.BuildFailure{}

	failure.JobName = os.Getenv("BUILD_JOB_NAME")
	failure.BuildName = os.Getenv("BUILD_NAME")
	failure.DetailsURL = fmt.Sprintf("%s/teams/%s/pipelines/%s/jobs/%s/builds/%s",
		os.Getenv("ATC_EXTERNAL_URL"),
		os.Getenv("BUILD_TEAM_NAME"),
		os.Getenv("BUILD_PIPELINE_NAME"),
		os.Getenv("BUILD_JOB_NAME"),
		os.Getenv("BUILD_NAME"))

	return driver.makeReady(failure)
}

func (driver *S3Driver) Check(cursor string) ([]string, error) {
	status := &models.PipelineStatus{}
	err := driver.load(status)

	versions := make([]string, 0, 1)

	if err == nil {
		switch status.State {
		case "":
			if cursor == "" {
				if driver.InitialVersion != "" {
					versions = append(versions, driver.InitialVersion)
				} else {
					versions = append(versions, "1")
				}
			}
		case models.StateReady:
			if strings.Compare(status.BuildNumber, cursor) >= 0 {
				versions = append(versions, status.BuildNumber)
			}
		}
	}

	return versions, err
}

func (driver *S3Driver) load(status *models.PipelineStatus) error {
	resp, err := driver.Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(driver.BucketName),
		Key:    aws.String(driver.Key),
	})

	if resp != nil && err == nil {
		var statusYaml []byte
		statusYaml, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		err = yaml.Unmarshal(statusYaml, status)
		if err != nil {
			return err
		}
	}

	return nil
}

func (driver *S3Driver) makeReady(failure *models.BuildFailure) (status *models.PipelineStatus, err error) {
	status = &models.PipelineStatus{}
	err = driver.load(status)
	if err == nil {
		if ok, err := driver.changeAndPersistState(status, models.StateReady, failure); ok {
			return status, err
		}
	}

	return nil, err
}

func (driver *S3Driver) changeAndPersistState(status *models.PipelineStatus,
	pipelineState models.PipelineState,
	failure *models.BuildFailure) (ok bool, err error) {

	if status != nil && err == nil {
		status.Failure = failure
		status, err = state.ChangeState(status, pipelineState, nil)

		if err == nil {
			outputYaml, marshalError := yaml.Marshal(status)
			if marshalError != nil {
				err = marshalError
			}

			if err == nil {
				params := &s3.PutObjectInput{
					Bucket:      aws.String(driver.BucketName),
					Key:         aws.String(driver.Key),
					ContentType: aws.String("text/plain"),
					Body:        bytes.NewReader(outputYaml),
					ACL:         aws.String(s3.ObjectCannedACLPrivate),
				}

				if len(driver.ServerSideEncryption) > 0 {
					params.ServerSideEncryption = aws.String(driver.ServerSideEncryption)
				}

				_, err = driver.Svc.PutObject(params)
			}
		}
	} else {
		err = fmt.Errorf("status was nil")
	}

	ok = (err == nil)
	return
}
