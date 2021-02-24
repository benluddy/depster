package dot

import (
	"fmt"
	"io"
	"os"

	"github.com/blang/semver/v4"
	"github.com/operator-framework/operator-registry/pkg/client"
	"github.com/spf13/cobra"

	"github.com/benluddy/depster/cmd/internal/commander"
)

type node struct {
	Name      string
	Version   semver.Version
	Channel   string
	Package   string
	Replaces  string
	Skips     []string
	SkipRange string
}

func AddTo(c commander.Interface) {
	var (
		output  string
		pkg     string
		channel string
	)

	dot := &cobra.Command{
		Use:   "dot",
		Short: "Generate a DOT representation of the catalog upgrade graph.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, err := client.NewClient(args[0])
			if err != nil {
				return err
			}
			defer cl.Close()

			var dst io.Writer
			if output == "-" {
				dst = os.Stdout
			} else {
				fd, err := os.Open(output)
				if err != nil {
					return err
				}
				defer fd.Close()
				dst = fd
			}

			it, err := cl.ListBundles(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Fprintf(dst, "digraph catalog {\n")

			pkgs := make(map[string][]node)
			for b := it.Next(); b != nil; b = it.Next() {
				if pkg != "" && b.PackageName != pkg {
					continue
				}
				if channel != "" && b.ChannelName != channel {
					continue
				}

				n := node{
					Name:      b.CsvName,
					Channel:   b.ChannelName,
					Package:   b.PackageName,
					Replaces:  b.Replaces,
					SkipRange: b.SkipRange,
				}

				if b.Version != "" {
					if sv, err := semver.Parse(b.Version); err != nil {
						fmt.Fprintf(os.Stderr, "unable to parse version of %q: %v\n", n.Name, err)
					} else {
						n.Version = sv
					}
				}

				for _, skip := range b.Skips {
					if skip == "" {
						continue // bad
					}
					n.Skips = append(n.Skips, skip)
				}

				pkgs[n.Package] = append(pkgs[n.Package], n)

				fmt.Fprintf(dst, "  %q;\n", n.Name)
				if n.Replaces != "" {
					fmt.Fprintf(dst, "  %q -> %q [label=\"replaces (%s)\"];\n", n.Name, n.Replaces, n.Channel)
				}

				for _, skip := range n.Skips {
					fmt.Fprintf(dst, "  %q -> %q [label=\"skips\"];\n", n.Name, skip)
				}
			}

			for pkg, ns := range pkgs {
				for i, n := range ns {
					if n.SkipRange == "" {
						continue
					}
					rg, err := semver.ParseRange(n.SkipRange)
					if err != nil {
						fmt.Fprintf(os.Stderr, "unable to parse skiprange of %q: %v\n", n.Name, err)
						continue
					}
					// matches nodes without channel entries
					catchall := fmt.Sprintf("%s (%s)", pkg, n.SkipRange)
					fmt.Fprintf(dst, "  %q [style=\"dotted\"];\n", catchall)
					fmt.Fprintf(dst, "  %q -> %q [label=\"skiprange\"];\n", n.Name, catchall)
					for j, m := range ns {
						// len(ns) is probably small
						if j == i || !rg(m.Version) {
							continue
						}
						fmt.Fprintf(dst, "  %q -> %q [label=\"skiprange\"];\n", n.Name, m.Name)
					}
				}
			}

			fmt.Fprintf(dst, "}\n")

			return it.Error()
		},
	}
	dot.Flags().StringVarP(&output, "output", "o", "-", "destination file")
	dot.Flags().StringVarP(&pkg, "package", "p", "", "filter nodes by package")
	dot.Flags().StringVarP(&channel, "channel", "c", "", "filter nodes by channel")

	c.AddCommand(dot)
}
