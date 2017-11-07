package base

type Storer interface {
}

type Persister interface {
	Read() ([]byte, error)
	Write([]byte) (int64, error)
}

type Publisher interface {
	Publish([]byte) (int64, error)
	Unpublish() error
}

type Connecter interface {
	Connect() (Storer, error)
	Close() error
}

type HeartBeater interface {
	Beat() error
}

type Operationor interface {
	Persister
	Publisher
	Connecter
	HeartBeater
}
