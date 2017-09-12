package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/pivotalservices/pipeline-status-resource/driver"
	"github.com/pivotalservices/pipeline-status-resource/models"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <source>")
		os.Exit(1)
	}

	// sources := os.Args[1]

	var request models.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	driver, err := driver.FromSource(request.Source)
	if err != nil {
		fatal("constructing driver", err)
	}

	status := &models.PipelineStatus{}

	switch request.Params.Action {
	case models.Start:
		ok, err := driver.Load(status)
		if !ok && err != nil {
			fatal("fetching status", err)
		}

		if request.Source.RequireReady {
			retryDuration, err := time.ParseDuration(request.Source.RetryAfter)
			if err != nil {
				retryDuration = models.DefaultRetryPeriod
			}

			for {
				if status.State == models.StateReady {
					break
				}

				fmt.Fprint(os.Stderr, ".")
				time.Sleep(retryDuration)

				ok, err = driver.Load(status)
				if !ok && err != nil {
					fatal("fetching status", err)
				}
			}
		}

		status, err = driver.Start()
	case models.Finish:
		status, err = driver.Finish()
	case models.Fail:
		status, err = driver.Fail()
	}

	if err != nil {
		fatal(fmt.Sprintf("%sing pipeline", request.Params.Action), err)
	}

	json.NewEncoder(os.Stdout).Encode(models.OutResponse{
		Version: models.Version{
			Number: status.BuildNumber,
		},
		Metadata: models.Metadata{
			{"number", status.BuildNumber},
		},
	})
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
