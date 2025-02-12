package pkg

type OnCompleteHandlers struct {
	OnSuccessHandler func()
	OnFailureHandler func()
}

func (onc OnCompleteHandlers) OnSuccess() {
	if onc.OnSuccessHandler != nil {
		onc.OnSuccessHandler()
	}
}

func (onc OnCompleteHandlers) OnFailure() {
	if onc.OnFailureHandler != nil {
		onc.OnFailureHandler()
	}
}

type StreamMessage[T any] struct {
	Message T
	OnCompleteHandlers
}
