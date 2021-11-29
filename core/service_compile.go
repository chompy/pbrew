package core

import (
	"errors"
	"fmt"
	"os"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

const awsKeyEnv = "AWS_ACCESS_KEY_ID"
const awsSecretEnv = "AWS_SECRET_ACCESS_KEY"
const awsS3BucketEnv = "AWS_S3_BUCKET"
const awsS3PathEnv = "AWS_S3_PATH"
const awsS3RegionEnv = "AWS_S3_REGION"

// Compile precompiles (brew bottle) the service and uploads it to AWS S3.
func (s *Service) Compile() error {
	done := output.Duration(fmt.Sprintf("Bottling %s.", s.BrewName))
	// connect to aws s3
	awsKey := os.Getenv(awsKeyEnv)
	awsSecret := os.Getenv(awsSecretEnv)
	awsBucket := os.Getenv(awsS3BucketEnv)
	//awsPath := os.Getenv(awsS3PathEnv)
	awsRegion := os.Getenv(awsS3RegionEnv)
	if awsKey == "" || awsSecret == "" || awsBucket == "" {
		return errors.New("missing environment variable(s).")
	}
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	if err := brewBottle(s.BrewName); err != nil {
		return err
	}

	//s, err := session.NewSession(&aws.Config)
	// fetch dependency info
	/*info, err := s.Info()
	if err != nil {
		return err
	}
	for _, depName := range info["dependencies"].([]interface{}) {
		done2 := output.Duration(fmt.Sprintf("Bottle dependency %s.", depName))
		if err := brewCommand("bottle", depName.(string)); err != nil {
			output.Warn(err.Error())
		}
		done2()
	}
	done2 := output.Duration(fmt.Sprintf("Bottle %s.", s.BrewName))
	if err := brewCommand("bottle", s.BrewName); err != nil {
		output.Warn(err.Error())
	}*/
	//done2()
	done()
	return nil
}
