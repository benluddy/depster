package manifest

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Loader interface {
	Next() bool
	Error() error
	Close() error
	ToUnstructured(dst *unstructured.Unstructured)
}
