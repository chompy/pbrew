package core

import (
	"fmt"

	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

// Compile precompiles (brew bottle) the service and uploads it to AWS S3.
func (s *Service) Compile() error {
	done := output.Duration(fmt.Sprintf("Bottling %s.", s.DisplayName()))
	if err := brewBottle(s.BrewName); err != nil {
		return err
	}
	if err := brewBottleUpload(s.BrewName); err != nil {
		return err
	}
	done()
	return nil
}
