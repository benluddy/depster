package resolve

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/benluddy/depster/cmd/internal/commander"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/cache"
	"github.com/operator-framework/operator-registry/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func AddTo(c commander.Interface) {
	resolve := &cobra.Command{
		Use:   "resolve",
		Short: "Perform dependency resolution.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var b InputBuilder
			for _, arg := range args {
				loaders := map[string]ManifestLoader{
					"":     SchemelessLoader{},
					"file": FileLoader{},
				}

				src, err := url.Parse(arg)
				if err != nil {
					return fmt.Errorf("failed to parse argument %q as url: %w", arg, err)
				}

				loader, ok := loaders[src.Scheme]
				if !ok {
					return fmt.Errorf("no manifest loader for scheme %q", src.Scheme)
				}

				var u unstructured.Unstructured
				if err := loader.LoadManifest(src, &u); err != nil {
					return fmt.Errorf("error loading manifest from %q: %w", src, err)
				}

				if err := b.Add(&u); err != nil {
					return fmt.Errorf("error adding input: %w", err)
				}
			}

			log := logrus.New()
			log.SetOutput(ioutil.Discard)
			if verbose, err := cmd.Flags().GetBool("verbose"); err != nil {
				return fmt.Errorf("error reading flag: %w", err)
			} else if verbose {
				log.SetOutput(os.Stderr)
				log.SetLevel(logrus.DebugLevel)
			}

			p := phonyRegistryClientProvider{
				clients: make(map[registry.CatalogKey]client.Interface),
			}

			for _, catalog := range b.catalogs {
				key := registry.CatalogKey{
					Namespace: catalog.GetNamespace(),
					Name:      catalog.GetName(),
				}
				if _, ok := p.clients[key]; ok {
					return fmt.Errorf("duplicate catalog source: %s/%s", key.Namespace, key.Name)
				}
				if c, err := client.NewClient(catalog.Spec.Address); err != nil {
					return fmt.Errorf("error creating registry client: %w", err)
				} else {
					p.clients[key] = c
				}

			}

			sp := &mergedSourceProvider{
				sps: []cache.SourceProvider{
					resolver.SourceProviderFromRegistryClientProvider(p, log),
					resolver.NewCSVSourceProvider(
						b.ClusterServiceVersionLister(),
						b.SubscriptionLister(),
						b.OperatorGroupLister(),
						log,
					),
				},
			}

			r := resolver.NewDefaultResolver(sp, b.PriorityProvider(), log)

			var nsnames []string
			for _, ns := range b.namespaces {
				nsnames = append(nsnames, ns.GetName())
			}

			operators, err := r.Resolve(nsnames, b.subscriptions)
			if err != nil {
				return fmt.Errorf("resolution failed: %w", err)
			}

			printer := makeTabularPrinter(cmd.OutOrStdout())
			for _, each := range operators {
				printer.Print(each)
			}
			return printer.Close()
		},
	}
	c.AddCommand(resolve)
}

type mergedSourceProvider struct {
	sps    []cache.SourceProvider
	logger logrus.StdLogger
}

func (msp *mergedSourceProvider) Sources(namespaces ...string) map[cache.SourceKey]cache.Source {
	result := make(map[cache.SourceKey]cache.Source)
	for _, sp := range msp.sps {
		for key, source := range sp.Sources(namespaces...) {
			if _, ok := result[key]; ok {
				msp.logger.Printf("warning: duplicate sourcekey: %q\n", key)
			}
			result[key] = source
		}
	}
	return result
}
