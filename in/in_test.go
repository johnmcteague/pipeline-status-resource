package main_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotalservices/pipeline-status-resource/models"
)

var _ = Describe("In", func() {
	var key string

	var tmpdir string
	var destination string

	var inCmd *exec.Cmd

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "in-destination")
		Expect(err).NotTo(HaveOccurred())

		destination = path.Join(tmpdir, "in-dir")

		inCmd = exec.Command(inPath, destination)
	})

	AfterEach(func() {
		os.RemoveAll(tmpdir)
	})

	Context("when executed", func() {
		var request models.InRequest
		var response models.InResponse

		var svc *s3.S3

		BeforeEach(func() {
			guid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())

			key = guid.String()

			creds := credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
			awsConfig := &aws.Config{
				Region:           aws.String(regionName),
				Credentials:      creds,
				S3ForcePathStyle: aws.Bool(true),
				MaxRetries:       aws.Int(12),
			}

			svc = s3.New(session.New(awsConfig))

			status := &models.PipelineStatus{
				Team:         "test-team",
				Pipeline:     "test-pipeline",
				BuildNumber:  "3",
				State:        models.StateReady,
				LastModified: "2017-09-10T20:27:00",
			}

			yaml, _ := yaml.Marshal(status)

			_, err = svc.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(bucketName),
				Key:         aws.String(key),
				ContentType: aws.String("text/plain"),
				Body:        bytes.NewReader(yaml),
				ACL:         aws.String(s3.ObjectCannedACLPrivate),
			})
			Expect(err).NotTo(HaveOccurred())

			request = models.InRequest{
				Version: models.Version{
					Number: "3",
				},
				Source: models.Source{
					Bucket:          bucketName,
					Key:             key,
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
					RegionName:      regionName,
				},
				Params: models.InParams{},
			}

			response = models.InResponse{}
		})

		AfterEach(func() {
			_, err := svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			stdin, err := inCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(inCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 5*time.Second).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should have created a file", func() {
			checkFile := path.Join(destination, "status")
			_, err := os.Stat(checkFile)
			Expect(os.IsNotExist(err)).Should(BeFalse())
		})
	})
})
