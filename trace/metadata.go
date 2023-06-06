package trace

import (
	"github.com/LabKiko/kiko-gokit/metadata"
	"go.opentelemetry.io/otel/propagation"
)

// MapCarrier is a TextMapCarrier that uses a map held in memory as a storage
// medium for propagated key-value pairs.
type metadataSupplier struct {
	metadata.Metadata
}

// assert that metadataSupplier implements the TextMapCarrier interface
var _ propagation.TextMapCarrier = (*metadataSupplier)(nil)

// Get returns the value associated with the passed key.
func (m *metadataSupplier) Get(key string) string {
	return m.Metadata[key]
}

// Set stores the key-value pair.
func (m *metadataSupplier) Set(key, value string) {
	m.Metadata[key] = value
}

// Keys lists the keys stored in this carrier.
func (m *metadataSupplier) Keys() []string {
	keys := make([]string, 0, len(m.Metadata))
	for k := range m.Metadata {
		keys = append(keys, k)
	}
	return keys
}
