package cmdutil

import (
	"go.krav.sh/krav/internal/config"
	"go.krav.sh/krav/internal/config/settings"
	"go.krav.sh/krav/internal/logging"
)

type Factory struct {
	Config func() (*settings.Config, error)
	Logger func() logging.Logger
	Paths  func() (config.Paths, error)
}
