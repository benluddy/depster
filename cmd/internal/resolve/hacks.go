package resolve

import (
	"fmt"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	listers "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/listers/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry"
	"github.com/operator-framework/operator-registry/pkg/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type phonyCatalogSourceLister struct{}

func (phonyCatalogSourceLister) List(selector labels.Selector) ([]*v1alpha1.CatalogSource, error) {
	return nil, nil
}

func (l phonyCatalogSourceLister) CatalogSources(namespace string) listers.CatalogSourceNamespaceLister {
	return l
}

func (phonyCatalogSourceLister) Get(name string) (*v1alpha1.CatalogSource, error) {
	return nil, fmt.Errorf("not implemented")
}

type InputBuilder struct {
	namespaces    []*v1.Namespace
	subscriptions []*v1alpha1.Subscription
	catalogs      []*v1alpha1.CatalogSource
	csvs          []*v1alpha1.ClusterServiceVersion
}

func (b *InputBuilder) Add(in *unstructured.Unstructured) error {
	switch in.GroupVersionKind() {
	case v1alpha1.SchemeGroupVersion.WithKind(v1alpha1.ClusterServiceVersionKind):
		var csv v1alpha1.ClusterServiceVersion
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(in.UnstructuredContent(), &csv); err != nil {
			return fmt.Errorf("failed to convert manifest: %w", err)
		}
		b.csvs = append(b.csvs, &csv)
	case v1alpha1.SchemeGroupVersion.WithKind("CatalogSource"):
		var c v1alpha1.CatalogSource
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(in.UnstructuredContent(), &c); err != nil {
			return fmt.Errorf("failed to convert manifest: %w", err)
		}
		if c.Spec.SourceType != v1alpha1.SourceTypeGrpc {
			return fmt.Errorf("unsupported catalog source type: %q", c.Spec.SourceType)
		}
		b.catalogs = append(b.catalogs, &c)
	case v1.SchemeGroupVersion.WithKind("Namespace"):
		var n v1.Namespace
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(in.UnstructuredContent(), &n); err != nil {
			return fmt.Errorf("failed to convert manifest: %w", err)
		}
		b.namespaces = append(b.namespaces, &n)
	case v1alpha1.SchemeGroupVersion.WithKind(v1alpha1.SubscriptionKind):
		var s v1alpha1.Subscription
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(in.UnstructuredContent(), &s); err != nil {
			return fmt.Errorf("failed to convert manifest: %w", err)
		}
		b.subscriptions = append(b.subscriptions, &s)
	default:
		return fmt.Errorf("gvk %q not recognized", in.GroupVersionKind())
	}
	return nil
}

type phonyRegistryClientProvider struct {
	clients map[registry.CatalogKey]client.Interface
}

func (p phonyRegistryClientProvider) ClientsForNamespaces(namespaces ...string) map[registry.CatalogKey]client.Interface {
	return p.clients
}
