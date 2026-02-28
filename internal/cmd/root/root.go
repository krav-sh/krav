package root

import (
	"github.com/spf13/cobra"

	"github.com/tbhb/arci/internal/cmdutil"
)

func NewCmdRoot(_ *cmdutil.Factory) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "arci <command> [flags]",
		Short: "ARCI CLI",
		Long:  "Agentic Requirements Composition & Integration.",
	}

	return cmd, nil
}
