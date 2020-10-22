package resolve

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestIsAHack(t *testing.T) {
	var b InputBuilder
	b.Add(&unstructured.Unstructured{})
}
