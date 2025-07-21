package server

import (
	"travel-api/internal/config"
	"travel-api/internal/injector"
)

func CreateContainer(cfg config.Config) (*injector.Container, error) {
	factory := injector.NewFactory()
	return factory.CreateProductionContainer(cfg)
}
