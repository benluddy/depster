package resolve

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/cache"
)

type tabularPrinter struct {
	w   *tabwriter.Writer
	err error
}

func makeTabularPrinter(w io.Writer) *tabularPrinter {
	p := &tabularPrinter{
		w: tabwriter.NewWriter(w, 0, 0, 1, ' ', 0),
	}
	_, p.err = fmt.Fprintf(p.w, "NAME\tPACKAGE\tCHANNEL\tCATALOG\tIMAGE\t\n")
	return p
}

func (p tabularPrinter) Print(e *cache.Entry) {
	if p.err != nil {
		return
	}

	s := &cache.OperatorSourceInfo{}
	if e.SourceInfo != nil {
		s = e.SourceInfo
	}
	_, p.err = fmt.Fprintf(p.w, "%s\t%s\t%s\t%s/%s\t%s\t\n", e.Name, s.Package, s.Channel, s.Catalog.Namespace, s.Catalog.Name, e.BundlePath)
}

func (p tabularPrinter) Close() error {
	err := p.w.Flush()
	if p.err != nil {
		return p.err
	}
	p.err = err
	if p.err == nil {
		p.err = fmt.Errorf("printer closed")
	}
	return err
}
