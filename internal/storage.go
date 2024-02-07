package internal

type MemStorage struct {
	values map[string]interface{}
}

type Storage interface {
	AddValue()
}
