package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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

var _ = Describe("Out", func() {
	var key string

	var source string

	var outCmd *exec.Cmd

	var stdOutString string

	var yamlTemplate string = `
---
pipeline: test-pipeline
team: test-team
build: %s
last_modified: %s
state: %s
`

	BeforeEach(func() {
		var err error

		source, err = ioutil.TempDir("", "out-source")
		Expect(err).NotTo(HaveOccurred())

		outCmd = exec.Command(outPath, source)
	})

	AfterEach(func() {
		os.RemoveAll(source)
	})

	Context("when executed", func() {
		var request models.OutRequest
		var response models.OutResponse

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

			request = models.OutRequest{
				Version: models.Version{},
				Source: models.Source{
					Bucket:          bucketName,
					Key:             key,
					AccessKeyID:     accessKeyID,
					SecretAccessKey: secretAccessKey,
					RegionName:      regionName,
				},
				Params: models.OutParams{},
			}

			response = models.OutResponse{}
		})

		AfterEach(func() {
			_, err := svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		getStatus := func() models.PipelineStatus {
			resp, err := svc.GetObject(&s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
			Expect(err).NotTo(HaveOccurred())

			contents, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			s := models.PipelineStatus{}
			err = yaml.Unmarshal(contents, &s)
			Expect(err).NotTo(HaveOccurred())

			//fmt.Println(string(contents))

			return s
		}

		putStatus := func(buildNum string, buildState models.PipelineState) string {
			now := time.Now().Format(models.ISO8601DateFormat)
			yaml := fmt.Sprintf(yamlTemplate, buildNum, now, buildState)

			_, err := svc.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(bucketName),
				Key:         aws.String(key),
				ContentType: aws.String("text/plain"),
				Body:        bytes.NewReader([]byte(yaml)),
				ACL:         aws.String(s3.ObjectCannedACLPrivate),
			})
			Expect(err).NotTo(HaveOccurred())

			return now
		}

		runAndExpectSuccess := func() {
			stdin, err := outCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(outCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 15*time.Second).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &response)
			Expect(err).NotTo(HaveOccurred())
		}

		runAndExpectFailure := func() {
			stdin, err := outCmd.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			session, err := gexec.Start(outCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			err = json.NewEncoder(stdin).Encode(request)
			Expect(err).NotTo(HaveOccurred())

			// account for roundtrip to s3
			Eventually(session, 5*time.Second).Should(gexec.Exit(1))

			stdOutString = string(session.Out.Contents())
		}

		Context("when starting a build", func() {
			JustBeforeEach(runAndExpectSuccess)

			BeforeEach(func() {
				request.Params.Action = models.Start
			})

			Context("for the first time", func() {
				Context("with an initial version", func() {
					BeforeEach(func() {
						request.Source.InitialVersion = "10"
					})

					It("reports the state as running with the initial version", func() {
						status := getStatus()
						Expect(status.State).Should(Equal(models.StateRunning))
						Expect(status.BuildNumber).Should(Equal(request.Source.InitialVersion))
					})
				})

				Context("without an initial version", func() {
					It("reports the state as running with version 1", func() {
						status := getStatus()
						Expect(status.State).Should(Equal(models.StateRunning))
						Expect(status.BuildNumber).Should(Equal("1"))
					})
				})

				Context("with locking enabled", func() {
					BeforeEach(func() {
						request.Source.RequireReady = true
						request.Source.RetryAfter = "2s"
					})

					It("reports the state as running with version 1", func() {
						status := getStatus()
						Expect(status.State).Should(Equal(models.StateRunning))
						Expect(status.BuildNumber).Should(Equal("1"))
					})
				})
			})

			Context("subsequent times", func() {
				Context("when the state is currently running", func() {
					var timestamp string
					BeforeEach(func() {
						timestamp = putStatus("1", models.StateRunning)
					})

					Context("without locking enabled", func() {
						It("nothing should change", func() {
							status := getStatus()
							Expect(status.State).Should(Equal(models.StateRunning))
							Expect(status.BuildNumber).Should(Equal("1"))
							Expect(status.LastModified).Should(Equal(timestamp))
						})
					})

					Context("with locking enabled", func() {
						var timestamp string
						BeforeEach(func() {
							request.Source.RequireReady = true
							request.Source.RetryAfter = "5s"

							timestamp = putStatus("5", models.StateRunning)
							go func() {
								time.Sleep(7 * time.Second)
								_ = putStatus("5", models.StateReady)
							}()
						})

						It("should start the build after 6 seconds", func() {
							status := getStatus()
							originalTime, _ := time.Parse(models.ISO8601DateFormat, timestamp)
							newTime, _ := time.Parse(models.ISO8601DateFormat, status.LastModified)

							atLeast := originalTime.Add(10 * time.Second)
							notMoreThan := originalTime.Add(15 * time.Second)

							Expect(newTime).Should(BeTemporally(">=", atLeast, 1*time.Second))
							Expect(newTime).Should(BeTemporally("<", notMoreThan, 1*time.Second))
							Expect(status.State).Should(Equal(models.StateRunning))
							Expect(status.BuildNumber).Should(Equal("6"))
						})
					})
				})

				Context("when the state is currently ready", func() {
					var timestamp string
					BeforeEach(func() {
						timestamp = putStatus("3", models.StateReady)
						time.Sleep(2 * time.Second)
					})

					It("should increment the build, change the state, and change the last modified date", func() {
						status := getStatus()
						Expect(status.State).Should(Equal(models.StateRunning))
						Expect(status.BuildNumber).Should(Equal("4"))
						Expect(strings.Compare(status.LastModified, timestamp)).Should(BeNumerically(">", 0))
					})
				})
			})
		})

		Context("when finishing a build", func() {
			BeforeEach(func() {
				request.Params.Action = models.Finish
			})

			Context("without an existing status", func() {
				JustBeforeEach(runAndExpectFailure)
				It("should fail", func() {})
			})

			Context("with an existing status", func() {
				JustBeforeEach(runAndExpectSuccess)

				Context("which is currently running", func() {
					var lastMod string
					BeforeEach(func() {
						lastMod = putStatus("10", models.StateRunning)
						time.Sleep(2 * time.Second)
					})

					It("should change the state and last modified but not the build number", func() {
						status := getStatus()
						Expect(status.BuildNumber).To(Equal("10"))
						Expect(status.State).To(Equal(models.StateReady))
						Expect(strings.Compare(status.LastModified, lastMod)).To(BeNumerically(">", 0))
					})
				})

				Context("which is currently ready", func() {
					var lastMod string
					BeforeEach(func() {
						lastMod = putStatus("10", models.StateReady)
						time.Sleep(2 * time.Second)
					})

					It("should change nothing", func() {
						status := getStatus()
						Expect(status.BuildNumber).To(Equal("10"))
						Expect(status.State).To(Equal(models.StateReady))
						Expect(status.LastModified).To(Equal(lastMod))
					})
				})
			})
		})

		Context("when failing a build", func() {
			BeforeEach(func() {
				os.Setenv("ATC_EXTERNAL_URL", "https://concourse.example.com")
				os.Setenv("BUILD_JOB_NAME", "test-job")
				os.Setenv("BUILD_NAME", "10")
				request.Params.Action = models.Fail
			})

			AfterEach(func() {
				os.Unsetenv("ATC_EXTERNAL_URL")
				os.Unsetenv("BUILD_JOB_NAME")
				os.Unsetenv("BUILD_NAME")
			})

			Context("without an existing status", func() {
				JustBeforeEach(runAndExpectFailure)
				It("should fail", func() {})
			})

			Context("with an existing status", func() {

				Context("which is currently running", func() {
					var lastMod string
					BeforeEach(func() {
						lastMod = putStatus("10", models.StateRunning)
						time.Sleep(2 * time.Second)
					})

					JustBeforeEach(runAndExpectSuccess)

					It("should change the state and last modified but not the build number", func() {
						status := getStatus()
						Expect(status.BuildNumber).To(Equal("10"))
						Expect(status.State).To(Equal(models.StateReady))
						Expect(strings.Compare(status.LastModified, lastMod)).To(BeNumerically(">", 0))
						Expect(status.Failure).ToNot(BeNil())
						Expect(status.Failure.DetailsURL).To(
							Equal("https://concourse.example.com/teams/test-team/pipelines/test-pipeline/jobs/test-job/builds/10"))
					})
				})

				Context("which is currently ready", func() {
					BeforeEach(func() {
						putStatus("10", models.StateReady)
					})

					JustBeforeEach(runAndExpectFailure)
					It("should fail", func() {})
				})
			})
		})
	})

})
