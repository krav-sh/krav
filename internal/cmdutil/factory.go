package cmdutil

import (
	"github.com/tbhb/arci/internal/config"
	"github.com/tbhb/arci/internal/config/settings"
	"github.com/tbhb/arci/internal/logging"
)

type Factory struct {
	Config func() (*settings.Config, error)
	Logger func() logging.Logger
	Paths  func() (config.Paths, error)
}
