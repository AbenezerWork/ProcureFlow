package health

import (
	"context"

	domainhealth "github.com/AbenezerWork/ProcureFlow/internal/domain/health"
)

type Service struct {
	appName     string
	environment string
	version     string
}

func NewService(appName, environment, version string) Service {
	return Service{
		appName:     appName,
		environment: environment,
		version:     version,
	}
}

func (s Service) Check(context.Context) domainhealth.Status {
	return domainhealth.Status{
		Name:        s.appName,
		Environment: s.environment,
		Version:     s.version,
		Status:      "ok",
	}
}
