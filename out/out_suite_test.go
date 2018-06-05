package main_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var outPath string

var accessKeyID = os.Getenv("STATUS_TESTING_ACCESS_KEY_ID")
var secretAccessKey = os.Getenv("STATUS_TESTING_SECRET_ACCESS_KEY")
var sessionToken = os.Getenv("STATUS_TESTING_SESSION_TOKEN")
var bucketName = os.Getenv("STATUS_TESTING_BUCKET")
var regionName = os.Getenv("STATUS_TESTING_REGION")

var pipelineName, pnExists = os.LookupEnv("BUILD_PIPELINE_NAME")
var teamName, tnExists = os.LookupEnv("BUILD_TEAM_NAME")

var _ = BeforeSuite(func() {
	var err error

	os.Setenv("BUILD_PIPELINE_NAME", "test-pipeline")
	os.Setenv("BUILD_TEAM_NAME", "test-team")

	Expect(accessKeyID).NotTo(BeEmpty(), "must specify $STATUS_TESTING_ACCESS_KEY_ID")
	Expect(secretAccessKey).NotTo(BeEmpty(), "must specify $STATUS_TESTING_SECRET_ACCESS_KEY")
	Expect(bucketName).NotTo(BeEmpty(), "must specify $STATUS_TESTING_BUCKET")
	Expect(regionName).NotTo(BeEmpty(), "must specify $STATUS_TESTING_REGION")

	outPath, err = gexec.Build("github.com/pivotalservices/pipeline-status-resource/out")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	if pnExists {
		os.Setenv("BUILD_PIPELINE_NAME", pipelineName)
	} else {
		os.Unsetenv("BUILD_PIPELINE_NAME")
	}

	if tnExists {
		os.Setenv("BUILD_TEAM_NAME", teamName)
	} else {
		os.Unsetenv("BUILD_TEAM_NAME")
	}

	gexec.CleanupBuildArtifacts()
})

func TestOut(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Out Suite")
}
