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
	Subscribe() (events <-chan E, dispose func())
}

type Publisher[E any] interface {
	Publish(e E)
}

func New[E any]() Broker[E] {
	return &broker[E]{
		subs: make(map[string]chan E),
	}
}

type broker[E any] struct {
	mux  sync.Mutex
	subs map[string]chan E
}

func (o *broker[E]) Subscribe() (events <-chan E, dispose func()) {
	id := uuid.NewString()
	c := make(chan E)

	o.mux.Lock()
	defer o.mux.Unlock()

	o.subs[id] = c

	return c, func() {
		o.mux.Lock()
		defer o.mux.Unlock()

		if c, ok := o.subs[id]; ok {
			close(c)
			delete(o.subs, id)
		}
	}
}

func (o *broker[E]) Publish(e E) {
	o.mux.Lock()
	defer o.mux.Unlock()

	for _, c := range o.subs {
		go func(c chan E) {
			c <- e
		}(c)
	}
}
