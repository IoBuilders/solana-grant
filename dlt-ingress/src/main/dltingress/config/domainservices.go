package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey/service"
)

type DomainServices struct {
	CustodyKeyExistsService        servicecustodykey.Exists
	CustodyKeyExistMultipleService servicecustodykey.ExistMultiple
}

func SetupDomainServices(
	repositories *DltIngressRepositories,
) *DomainServices {
	return &DomainServices{
		CustodyKeyExistsService:        *servicecustodykey.NewExists(repositories.CustodyKeyRepo),
		CustodyKeyExistMultipleService: *servicecustodykey.NewExistMultiple(repositories.CustodyKeyRepo),
	}
}
