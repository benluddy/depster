package resolve

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type ManifestLoader interface {
	// todo: remove one-to-one url/manifest assumption
	LoadManifest(src *url.URL, dst *unstructured.Unstructured) error
}

type SchemelessLoader struct{}

func (l SchemelessLoader) LoadManifest(src *url.URL, dst *unstructured.Unstructured) error {
	abs, err := filepath.Abs(src.String())
	if err != nil {
		return fmt.Errorf("could not interpret schemeless argument as file path: %w", err)
	}
	src = &url.URL{
		Scheme: "file",
		Path:   abs,
	}
	return FileLoader{}.LoadManifest(src, dst)
}

type FileLoader struct{}

func (l FileLoader) LoadManifest(src *url.URL, dst *unstructured.Unstructured) error {
	fd, err := os.Open(src.Path)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", src.Path, err)
	}
	defer fd.Close()

	dec := yaml.NewYAMLOrJSONDecoder(fd, 64)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("failed to decode file %q: %w", src.Path, err)
	}

	return nil
}
