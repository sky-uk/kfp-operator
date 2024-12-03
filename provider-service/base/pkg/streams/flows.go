package streams

type PassThroughFlow[T, E any] struct {
	in       chan T
	out      chan T
	errorOut chan E
	effect   func(T)
}

func (ptf *PassThroughFlow[T, E]) In() chan<- T {
	return ptf.in
}

func (ptf *PassThroughFlow[T, E]) Out() <-chan T {
	return ptf.out
}

func (ptf *PassThroughFlow[T, E]) ErrOut() <-chan E {
	return ptf.errorOut
}

func (ptf *PassThroughFlow[T, E]) From(outlet Outlet[T]) Flow[T, T, E] {
	go func() {
		for message := range outlet.Out() {
			ptf.In() <- message
		}
	}()
	return ptf
}

func (ptf *PassThroughFlow[T, E]) To(inlet Inlet[T]) {
	go func() {
		for message := range ptf.out {
			inlet.In() <- message
		}
	}()
}

func (ptf *PassThroughFlow[T, E]) Error(inlet Inlet[E]) {
	for errorMessage := range ptf.errorOut {
		inlet.In() <- errorMessage
	}
}

func NewPassThroughFlow[T, E any](effect func(T)) PassThroughFlow[T, E] {
	flow := PassThroughFlow[T, E]{
		in:       make(chan T),
		out:      make(chan T),
		errorOut: make(chan E),
		effect:   effect,
	}

	go func() {
		for msg := range flow.in {
			effect(msg)
			flow.out <- msg
		}
	}()

	return flow
}
