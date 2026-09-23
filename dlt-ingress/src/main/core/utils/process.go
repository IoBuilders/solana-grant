package utils

import "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/baseprocess"

func SigningTypeFromOrigin(origin string) baseprocess.SigningType {
	if origin == "EXTERNAL" {
		return baseprocess.NonCustodial
	}
	return baseprocess.Custodial
}
