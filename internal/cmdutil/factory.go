package cmdutil

import (
	"pkg.krav.sh/krav/internal/config"
	"pkg.krav.sh/krav/internal/config/settings"
	"pkg.krav.sh/krav/internal/logging"
)

type Factory struct {
	Config func() (*settings.Config, error)
	Logger func() logging.Logger
	Paths  func() (config.Paths, error)
}
