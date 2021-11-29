package core

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"os"
	"path/filepath"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const awsS3BucketEnv = "AWS_S3_BUCKET"
const awsS3PathEnv = "AWS_S3_PATH"
const awsS3RegionEnv = "AWS_S3_REGION"
const bottleURLPrefix = "https://platform-cc-releases.s3.amazonaws.com/_pbrew_bottles/"

func brewBottle(name string) error {
	info, err := brewInfo(name)
	if err != nil {
		return err
	}
	if len(info["installed"].([]interface{})) == 0 {
		if err := brewCommand("install", name, "--build-bottle"); err != nil {
			return err
		}
	} else {
		installInfo := info["installed"].([]interface{})[0].(map[string]interface{})
		if !installInfo["built_as_bottle"].(bool) {
			output.LogInfo("Service was not built as bottle, reinstalling...")
			if err := brewCommand("uninstall", name); err != nil {
				return err
			}
			if err := brewCommand("install", name, "--build-bottle"); err != nil {
				return err
			}
		}
	}
	brewCommand("services", "stop", name)
	if err := brewCommand("bottle", name); err != nil {
		return err
	}
	return nil
}

func brewBottleFindLocal(name string) (string, error) {
	m, err := filepath.Glob(fmt.Sprintf("%s--*.tar.gz", brewAppName(name)))
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(m) > 0 {
		return m[0], nil
	}
	return "", errors.WithStack(os.ErrNotExist)
}

func brewBottleUploadName(name string) string {
	archStr := runtime.GOARCH
	switch runtime.GOARCH {
	case "amd64":
		{
			archStr = "x86_64"
		}
	}
	return fmt.Sprintf(
		"%s--0.%s_%s.bottle.1.tar.gz", brewAppName(name), archStr, runtime.GOOS,
	)
}

func brewBottleUpload(name string) error {
	awsBucket := os.Getenv(awsS3BucketEnv)
	awsPath := os.Getenv(awsS3PathEnv)
	awsRegion := os.Getenv(awsS3RegionEnv)
	if awsBucket == "" {
		return errors.New("missing AWS S3 bucket environment variable.")
	}
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}
	fileName, err := brewBottleFindLocal(name)
	if err != nil {
		return err
	}
	uploadName := brewBottleUploadName(name)
	s, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      aws.String(awsRegion),
	})
	if err != nil {
		return errors.WithStack(err)
	}
	bottleFile, err := os.Open(fileName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer bottleFile.Close()
	if _, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(awsBucket),
		Key:    aws.String(strings.Trim(strings.Trim(awsPath, "/")+"/"+uploadName, "/")),
		Body:   bottleFile,
	}); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func brewBottleRemoteURL(name string) string {
	return bottleURLPrefix + brewBottleUploadName(name)
}

func brewBottleDownloadPath(name string) string {
	return filepath.Join(GetDir(BottleDir), brewBottleUploadName(name))
}

func brewBottleDownload(name string) error {
	resp, err := http.Get(brewBottleRemoteURL(name))
	if err != nil {
		return errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.WithStack(os.ErrNotExist)
	}
	bottleFile, err := os.Create(brewBottleDownloadPath(name))
	if err != nil {
		return errors.WithStack(err)
	}
	defer bottleFile.Close()
	if _, err := io.Copy(bottleFile, resp.Body); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
