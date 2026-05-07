package app

import (
	"github.com/Skyinfi/app-manager/internal/config"
	"github.com/Skyinfi/app-manager/internal/httpapi"
	"github.com/Skyinfi/app-manager/internal/service"
	"github.com/Skyinfi/app-manager/internal/store"
)

func DefaultStore() *store.Store {
	return store.Default()
}

func NewAuthService(cfg config.Config) *service.AuthService {
	return service.NewAuth(cfg)
}

func NewServer(st *store.Store, auth *service.AuthService, opts ...httpapi.Option) *httpapi.Server {
	return httpapi.New(service.New(st), auth, opts...)
}
