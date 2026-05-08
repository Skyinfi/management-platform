package app

import (
	"github.com/Skyinfi/management-platform/app-manager/internal/config"
	"github.com/Skyinfi/management-platform/app-manager/internal/httpapi"
	"github.com/Skyinfi/management-platform/app-manager/internal/service"
	"github.com/Skyinfi/management-platform/app-manager/internal/store"
)

type Dependencies struct {
	Store   *store.Store
	Auth    *service.AuthService
	Docker  *service.DockerService
	Process *service.ProcessService
	Audit   *service.AuditLog
	Service *service.Service
}

func Bootstrap(cfg config.Config) Dependencies {
	st := store.Default()
	auth := service.NewAuth(cfg)
	dockerSvc := service.NewDockerService(cfg.Docker)
	scannerClient := service.NewScannerClient(cfg.Scanner.Addr)
	processSvc := service.NewProcessService(cfg.Services, scannerClient)
	audit := service.NewAuditLog()
	svc := service.New(st,
		service.WithDocker(dockerSvc),
		service.WithProcess(processSvc),
	)

	return Dependencies{
		Store:   st,
		Auth:    auth,
		Docker:  dockerSvc,
		Process: processSvc,
		Audit:   audit,
		Service: svc,
	}
}

func NewServer(deps Dependencies, opts ...httpapi.Option) *httpapi.Server {
	return httpapi.New(
		deps.Service,
		deps.Auth,
		deps.Docker,
		deps.Process,
		deps.Audit,
		opts...,
	)
}
