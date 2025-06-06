package store

type InMemoryStore struct {
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{}
}
