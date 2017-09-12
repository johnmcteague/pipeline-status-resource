package main_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var checkPath string

var accessKeyID = os.Getenv("STATUS_TESTING_ACCESS_KEY_ID")
var secretAccessKey = os.Getenv("STATUS_TESTING_SECRET_ACCESS_KEY")
var bucketName = os.Getenv("STATUS_TESTING_BUCKET")
var regionName = os.Getenv("STATUS_TESTING_REGION")

var _ = BeforeSuite(func() {
	var err error

	Expect(accessKeyID).NotTo(BeEmpty(), "must specify $STATUS_TESTING_ACCESS_KEY_ID")
	Expect(secretAccessKey).NotTo(BeEmpty(), "must specify $STATUS_TESTING_SECRET_ACCESS_KEY")
	Expect(bucketName).NotTo(BeEmpty(), "must specify $STATUS_TESTING_BUCKET")
	Expect(regionName).NotTo(BeEmpty(), "must specify $STATUS_TESTING_REGION")

	checkPath, err = gexec.Build("github.com/pivotalservices/pipeline-status-resource/check")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestCheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Check Suite")
}
