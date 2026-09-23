package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/domain/custodykey/service"
	"dlt-ingress/src/main/dltingress/port/repository"
)

type DomainServices struct {
	CustodyKeyExistsService        servicecustodykey.Exists
	CustodyKeyExistMultipleService servicecustodykey.ExistMultiple
}

func SetupDomainServices(
	repositories *repository.DltIngressRepositories,
) *DomainServices {
	return &DomainServices{
		CustodyKeyExistsService:        *servicecustodykey.NewExists(repositories.CustodyKeyRepo),
		CustodyKeyExistMultipleService: *servicecustodykey.NewExistMultiple(repositories.CustodyKeyRepo),
	}
}
