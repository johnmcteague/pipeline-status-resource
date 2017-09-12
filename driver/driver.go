package driver

import (
	"fmt"

	"github.com/adammck/venv"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pivotalservices/pipeline-status-resource/models"
)

type Driver interface {
	Check(lastModCursor string) ([]string, error)
	Load(status *models.PipelineStatus) error
	Start() (*models.PipelineStatus, error)
	Finish() (*models.PipelineStatus, error)
	Fail() (*models.PipelineStatus, error)
}

const maxRetries = 12

func FromSource(source models.Source) (Driver, error) {
	initialVersion := source.InitialVersion

	switch source.Driver {
	case models.DriverUnspecified, models.DriverS3:
		var creds *credentials.Credentials

		if source.AccessKeyID == "" && source.SecretAccessKey == "" {
			creds = credentials.AnonymousCredentials
		} else {
			creds = credentials.NewStaticCredentials(source.AccessKeyID, source.SecretAccessKey, "")
		}

		regionName := source.RegionName
		if len(regionName) == 0 {
			regionName = "us-east-1"
		}

		awsConfig := &aws.Config{
			Region:           aws.String(regionName),
			Credentials:      creds,
			S3ForcePathStyle: aws.Bool(true),
			MaxRetries:       aws.Int(maxRetries),
			DisableSSL:       aws.Bool(source.DisableSSL),
		}

		if len(source.Endpoint) != 0 {
			awsConfig.Endpoint = aws.String(source.Endpoint)
		}

		svc := s3.New(session.New(awsConfig))

		return &S3Driver{
			InitialVersion: initialVersion,

			Env:                  venv.OS(),
			Svc:                  svc,
			BucketName:           source.Bucket,
			Key:                  source.Key,
			ServerSideEncryption: source.ServerSideEncryption,
		}, nil

		/*
			THESE ARE CURRENTLY UNSUPPORTED

			case models.DriverGit:
				return &GitDriver{
					InitialVersion: initialVersion,

					URI:        source.URI,
					Branch:     source.Branch,
					PrivateKey: source.PrivateKey,
					Username:   source.Username,
					Password:   source.Password,
					File:       source.File,
					GitUser:    source.GitUser,
				}, nil

			case models.DriverSwift:
				return NewSwiftDriver(&source)

			case models.DriverGCS:
				servicer := &GCSIOServicer{
					JSONCredentials: source.JSONKey,
				}

				return &GCSDriver{
					InitialVersion: initialVersion,

					Servicer:   servicer,
					BucketName: source.Bucket,
					Key:        source.Key,
				}, nil
		*/
	default:
		return nil, fmt.Errorf("unknown driver: %s", source.Driver)
	}
}
