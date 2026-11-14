package mocks

import (
	"context"
	"hash/fnv"
	"strings"
	"sync"

	sharedTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/shared"
)

const EmbedderModel = "fake-embedding-model"

type Embedder struct {
	mutex      sync.Mutex
	Dimensions int
	Err        error
	Calls      [][]string
	Pinned     map[string][]float32
}

var _ sharedTypes.Embedder = (*Embedder)(nil)

func NewEmbedder(dimensions int) *Embedder {
	return &Embedder{Dimensions: dimensions}
}

func (e *Embedder) Model() string {
	return EmbedderModel
}

func (e *Embedder) CreateEmbeddings(_ context.Context, inputs []string) ([][]float32, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.Calls = append(e.Calls, append([]string(nil), inputs...))
	if e.Err != nil {
		return nil, e.Err
	}

	embeddings := make([][]float32, 0, len(inputs))
	for _, input := range inputs {
		embeddings = append(embeddings, e.vectorFor(input))
	}

	return embeddings, nil
}

func (e *Embedder) Pin(marker string, vector []float32) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.Pinned == nil {
		e.Pinned = map[string][]float32{}
	}

	e.Pinned[marker] = vector
}

func (e *Embedder) SetErr(err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.Err = err
}

func (e *Embedder) Reset() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.Err = nil
	e.Calls = nil
	e.Pinned = nil
}

func (e *Embedder) CallCount() int {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return len(e.Calls)
}

func (e *Embedder) InputCount() int {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	total := 0
	for _, call := range e.Calls {
		total += len(call)
	}

	return total
}

func (e *Embedder) vectorFor(input string) []float32 {
	for marker, vector := range e.Pinned {
		if strings.Contains(input, marker) {
			return vector
		}
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(input))
	seed := hasher.Sum32()

	vector := make([]float32, e.Dimensions)
	for index := range vector {
		seed = seed*1664525 + 1013904223
		vector[index] = float32(seed%2000)/1000 - 1
	}

	return vector
}
