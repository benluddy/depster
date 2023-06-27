package resolve

import (
	"fmt"

	v1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	v1listers "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/listers/operators/v1"
	v1alpha1listers "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/listers/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/cache"
	"github.com/operator-framework/operator-registry/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgocache "k8s.io/client-go/tools/cache"
)

type phonySourcePriorityProvider struct {
	catalogs []*v1alpha1.CatalogSource
}

func (p *phonySourcePriorityProvider) Priority(k cache.SourceKey) int {
	for _, cs := range p.catalogs {
		if cs.Name == k.Name && cs.Namespace == k.Namespace {
			return cs.Spec.Priority
		}
	}
	return 0
}

type InputBuilder struct {
	namespaces    []*corev1.Namespace
	subscriptions []*v1alpha1.Subscription
	catalogs      []*v1alpha1.CatalogSource
	csvs          []*v1alpha1.ClusterServiceVersion
	ogs           []*v1.OperatorGroup
}

func (b *InputBuilder) PriorityProvider() cache.SourcePriorityProvider {
	return &phonySourcePriorityProvider{catalogs: b.catalogs}
}

func (b *InputBuilder) Add(in *unstructured.Unstructured) error {
	switch in.GroupVersionKind() {
	case v1.SchemeGroupVersion.WithKind(v1.OperatorGroupKind):
		var og v1.OperatorGroup
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(in.UnstructuredContent(), &og); err != nil {
			return fmt.Errorf("failed to convert manifest: %w", err)
		}
		b.ogs = append(b.ogs, &og)
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
	case corev1.SchemeGroupVersion.WithKind("Namespace"):
		var n corev1.Namespace
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

func (b *InputBuilder) ClusterServiceVersionLister() v1alpha1listers.ClusterServiceVersionLister {
	indexer := clientgocache.NewIndexer(clientgocache.MetaNamespaceKeyFunc, clientgocache.Indexers{clientgocache.NamespaceIndex: clientgocache.MetaNamespaceIndexFunc})
	for _, csv := range b.csvs {
		indexer.Add(csv)
	}
	return v1alpha1listers.NewClusterServiceVersionLister(indexer)
}

func (b *InputBuilder) SubscriptionLister() v1alpha1listers.SubscriptionLister {
	indexer := clientgocache.NewIndexer(clientgocache.MetaNamespaceKeyFunc, clientgocache.Indexers{clientgocache.NamespaceIndex: clientgocache.MetaNamespaceIndexFunc})
	for _, s := range b.subscriptions {
		indexer.Add(s)
	}
	return v1alpha1listers.NewSubscriptionLister(indexer)
}

func (b *InputBuilder) OperatorGroupLister() v1listers.OperatorGroupLister {
	indexer := clientgocache.NewIndexer(clientgocache.MetaNamespaceKeyFunc, clientgocache.Indexers{clientgocache.NamespaceIndex: clientgocache.MetaNamespaceIndexFunc})
	for _, og := range b.ogs {
		indexer.Add(og)
	}
	return v1listers.NewOperatorGroupLister(indexer)
}

type phonyRegistryClientProvider struct {
	clients map[registry.CatalogKey]client.Interface
}

func (p phonyRegistryClientProvider) ClientsForNamespaces(namespaces ...string) map[registry.CatalogKey]client.Interface {
	return p.clients
}
