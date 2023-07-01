package broker

import (
	"sync"

	"github.com/google/uuid"
)

type Broker[E any] interface {
	Publisher[E]
	Subscriber[E]
}

type Subscriber[E any] interface {
	Subscribe(handler func(e E)) (dispose func())
}

type Publisher[E any] interface {
	Publish(e E)
}

func New[E any]() Broker[E] {
	return &broker[E]{
		subs: map[string]func(E){},
	}
}

type broker[E any] struct {
	lock sync.Mutex
	subs map[string]func(E)
}

func (o *broker[E]) Subscribe(h func(e E)) (dispose func()) {
	id := uuid.NewString()

	o.lock.Lock()
	defer o.lock.Unlock()

	o.subs[id] = h

	return func() {
		o.dispose(id)
	}
}

func (o *broker[E]) Publish(e E) {
	o.lock.Lock()
	defer o.lock.Unlock()

	for _, h := range o.subs {
		if h != nil {
			h(e)
		}
	}
}

func (o *broker[E]) dispose(id string) {
	o.lock.Lock()
	defer o.lock.Unlock()

	delete(o.subs, id)
}
