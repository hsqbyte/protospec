package checksum

import (
	"github.com/hsqbyte/protospec/src/errors"
)

// ChecksumFunc is the signature for a checksum algorithm function.
// It takes a byte slice and returns a checksum value as uint64.
type ChecksumFunc func(data []byte) uint64

// ChecksumRegistry manages registered checksum algorithms.
type ChecksumRegistry struct {
	algorithms map[string]ChecksumFunc
}

// NewChecksumRegistry creates a new empty ChecksumRegistry.
func NewChecksumRegistry() *ChecksumRegistry {
	return &ChecksumRegistry{
		algorithms: make(map[string]ChecksumFunc),
	}
}

// Register adds a checksum algorithm to the registry.
func (r *ChecksumRegistry) Register(name string, fn ChecksumFunc) {
	r.algorithms[name] = fn
}

// Get retrieves a checksum algorithm by name.
// Returns AlgorithmNotFoundError if the algorithm is not registered.
func (r *ChecksumRegistry) Get(name string) (ChecksumFunc, error) {
	fn, ok := r.algorithms[name]
	if !ok {
		return nil, &errors.AlgorithmNotFoundError{Name: name}
	}
	return fn, nil
}

// Has checks whether a checksum algorithm is registered.
func (r *ChecksumRegistry) Has(name string) bool {
	_, ok := r.algorithms[name]
	return ok
}
