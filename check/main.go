package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pivotalservices/pipeline-status-resource/driver"
	"github.com/pivotalservices/pipeline-status-resource/models"
)

func main() {
	var request models.CheckRequest
	err := json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	if driver.IsDebug(request.Source) {
		if tmpFile, err := ioutil.TempFile("/tmp", "checkdbg"); err == nil {
			json.NewEncoder(tmpFile).Encode(request)
		} else {
			fmt.Fprintf(os.Stderr, "Error writing debug output: %v\n", err)
		}
	}

	driver, err := driver.FromSource(request.Source)
	if err != nil {
		fatal("constructing driver", err)
	}

	versions, err := driver.Check(request.Version.Number)
	if err != nil {
		fatal("checking for new versions", err)
	}

	delta := models.CheckResponse{}
	for _, v := range versions {
		delta = append(delta, models.Version{
			Number: v,
		})
	}

	json.NewEncoder(os.Stdout).Encode(delta)
}

func fatal(doing string, err error) {
	println("error " + doing + ": " + err.Error())
	os.Exit(1)
}
