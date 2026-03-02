package root

import (
	"github.com/spf13/cobra"

	"go.krav.sh/krav/internal/cmdutil"
)

func NewCmdRoot(_ *cmdutil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "krav <command> [flags]",
		Short: "Krav CLI",
		Long:  "Krav.",
	}

	return cmd, nil
}
