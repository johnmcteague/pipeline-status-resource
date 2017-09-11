package driver_test

import (
	"io/ioutil"
	"strings"

	"github.com/adammck/venv"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/pipeline-status-resource/driver"
)

var mockEnv venv.Env = venv.Mock()
var _ = BeforeSuite(func() {
	mockEnv.Setenv("BUILD_PIPELINE_NAME", "bar")
	mockEnv.Setenv("BUILD_TEAM_NAME", "foo")
})

var _ = Describe("S3 Driver", func() {
	Context("with encryption", func() {
		It("sets it when enabled", func() {
			s := &service{}
			d := driver.S3Driver{
				Svc:                  s,
				Env:                  mockEnv,
				ServerSideEncryption: "my-encryption-schema",
			}
			d.Start()
			Expect(*s.params.ServerSideEncryption).To(Equal("my-encryption-schema"))
		})
		It("leaves it empty when disabled", func() {
			s := &service{}
			d := driver.S3Driver{
				Svc: s,
				Env: mockEnv,
			}
			d.Start()
			Expect(s.params.ServerSideEncryption).To(BeNil())
		})
	})
})

type service struct {
	params *s3.PutObjectInput
}

func (*service) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	sampleYaml := `
---
team: foo
pipeline: bar
build: 3
last_modified: 2017-03-14T23:33:45
state: READY
`
	out := &s3.GetObjectOutput{}
	out.Body = ioutil.NopCloser(strings.NewReader(sampleYaml))

	return out, nil
}

func (s *service) PutObject(p *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	s.params = p
	return nil, nil
}
