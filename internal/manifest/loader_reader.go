package manifest

import (
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type decoder interface {
	Decode(into interface{}) error
}

func ReaderLoader(r io.Reader) Loader {
	return &readerLoader{
		d: yaml.NewYAMLOrJSONDecoder(r, 64),
	}
}

type readerLoader struct {
	d decoder
	e error
}

func (l *readerLoader) Next() bool {
	return l.e == nil
}

func (l *readerLoader) Error() error {
	return l.e
}

func (l *readerLoader) ToUnstructured(dst *unstructured.Unstructured) {
	l.e = l.d.Decode(dst)
}
