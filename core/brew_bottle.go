package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const awsKeyEnv = "AWS_ACCESS_KEY_ID"
const awsSecretEnv = "AWS_SECRET_ACCESS_KEY"
const awsS3BucketEnv = "AWS_S3_BUCKET"
const awsS3PathEnv = "AWS_S3_PATH"
const awsS3RegionEnv = "AWS_S3_REGION"

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

func brewBottleFind(name string) (string, error) {
	m, err := filepath.Glob(fmt.Sprintf("%s--*.tar.gz", brewAppName(name)))
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(m) > 0 {
		return m[0], nil
	}
	return "", errors.WithStack(os.ErrNotExist)
}

func brewBottleUpload(name string) error {

	// setup aws params
	awsKey := os.Getenv(awsKeyEnv)
	awsSecret := os.Getenv(awsSecretEnv)
	awsBucket := os.Getenv(awsS3BucketEnv)
	//awsPath := os.Getenv(awsS3PathEnv)
	awsRegion := os.Getenv(awsS3RegionEnv)
	if awsKey == "" || awsSecret == "" || awsBucket == "" {
		return errors.New("missing AWS S3 environment variable(s).")
	}
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}
	// get bottle
	fileName, err := brewBottleFind(name)
	if err != nil {
		return err
	}
	uploadName := fmt.Sprintf(
		"%s--%s-%s.tar.gz", brewAppName(name), runtime.GOOS, runtime.GOARCH,
	)
	s, err := session.NewSession(&aws.Config{
		
	})
	// fetch dependency info
	/*info, err := s.Info()
	if err != nil {
		return err
	}


	return nil
}
