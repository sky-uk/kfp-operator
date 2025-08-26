package pkg

import _ "github.com/sky-uk/kfp-operator/pkg/common"

type OnCompleteHandlers struct {
	OnSuccessHandler              func()
	OnRecoverableFailureHandler   func()
	OnUnrecoverableFailureHandler func()
}

func (onc OnCompleteHandlers) OnSuccess() {
	if onc.OnSuccessHandler != nil {
		onc.OnSuccessHandler()
	}
}

func (onc OnCompleteHandlers) OnRecoverableFailure() {
	if onc.OnRecoverableFailureHandler != nil {
		onc.OnRecoverableFailureHandler()
	}
}

func (onc OnCompleteHandlers) OnUnrecoverableFailure() {
	if onc.OnUnrecoverableFailureHandler != nil {
		onc.OnUnrecoverableFailureHandler()
	}
}

type StreamMessage[T any] struct {
	Message T
	OnCompleteHandlers
}
