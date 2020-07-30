package commander

import (
	"github.com/spf13/cobra"
)

type Interface interface {
	AddCommand(...*cobra.Command)
}
