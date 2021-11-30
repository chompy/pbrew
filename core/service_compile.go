package core

import (
	"fmt"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// Compile precompiles (brew bottle) the service and uploads it to AWS S3.
func (s *Service) Compile() error {
	done := output.Duration(fmt.Sprintf("Bottling %s.", s.BrewName))

	if err := brewBottle(s.BrewName); err != nil {
		return err
	}
	if err := brewBottleUpload(s.BrewName); err != nil {
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
