package pkg

type Inlet[I any] interface {
	In() chan<- I
}

type Outlet[O any] interface {
	Out() <-chan O
}

type ErrorOutlet[E any] interface {
	ErrOut() <-chan E
}

type Source[O any] interface {
	Outlet[O]
}

type Flow[I any, O any, E any] interface {
	Inlet[I]
	Outlet[O]
	ErrorOutlet[E]
	From(Outlet[I]) Flow[I, O, E]
	To(Sink[O])
	Error(Sink[E])
}

type Sink[I any] interface {
	Inlet[I]
}
