package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"

	"github.com/pivotalservices/pipeline-status-resource/driver"
	"github.com/pivotalservices/pipeline-status-resource/models"
)

func main() {
	if len(os.Args) < 2 {
		println("usage: " + os.Args[0] + " <destination>")
		os.Exit(1)
	}

	destination := os.Args[1]

	err := os.MkdirAll(destination, 0755)
	if err != nil {
		fatal("creating destination", err)
	}

	var request models.InRequest
	err = json.NewDecoder(os.Stdin).Decode(&request)
	if err != nil {
		fatal("reading request", err)
	}

	if driver.IsDebug(request.Source) {
		if tmpFile, err := ioutil.TempFile("/tmp", "indbg"); err == nil {
			json.NewEncoder(tmpFile).Encode(request)
		} else {
			fmt.Fprintf(os.Stderr, "Error writing debug output: %v\n", err)
		}
	}

	driver, err := driver.FromSource(request.Source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	status := &models.PipelineStatus{}
	ok, err := driver.Load(status)
	if !ok {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fileName := path.Join(destination, "status")

	if data, err := yaml.Marshal(status); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		ioutil.WriteFile(fileName, data, 0644)
	}

	json.NewEncoder(os.Stdout).Encode(models.InResponse{
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
