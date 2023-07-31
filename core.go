// Package gauss contains utilities to execute
package gauss

import (
	"sync"
)

type Return interface {
	Error() error
	ReturnValues() []interface{}
}

type returnImpl struct {
	error        error
	returnValues []interface{}
}

func (_self *returnImpl) Error() error {
	return _self.error
}
func (_self *returnImpl) ReturnValues() []interface{} {
	return _self.returnValues
}

func NewReturn(err error, returnValues ...interface{}) Return {
	return &returnImpl{error: err, returnValues: returnValues}
}

type Function func() Return

func JoinFailOnAnyError(funcs ...Function) ([]Return, error) {
	errorChannel := make(chan error)
	completeChannel := make(chan bool)
	var wg sync.WaitGroup
	returns := make([]Return, len(funcs))
	for index, function := range funcs {
		wg.Add(1)
		go func(returnValues []Return, index int, function Function) {
			returnValuesByFunction := function()
			wg.Done()
			returnValues[index] = returnValuesByFunction
			if returnValuesByFunction.Error() != nil {
				errorChannel <- returnValuesByFunction.Error()
			}
		}(returns, index, function)
	}

	go func() {
		wg.Wait()
		close(completeChannel)
	}()

	select {
	case <-completeChannel:
		return returns, nil
	case err := <-errorChannel:
		return returns, err
	}
}

// JoinCompleteAll Run functions and return when complete all functions, first return value contain
// return values and second value return true if success operation, false otherwise.
func JoinCompleteAll(funcs ...Function) ([]Return, bool) {
	var wg sync.WaitGroup
	returns := make([]Return, len(funcs))
	for index, function := range funcs {
		wg.Add(1)
		go func(returnValues []Return, index int, function Function) {
			returnValuesByFunction := function()
			returnValues[index] = returnValuesByFunction
			wg.Done()
		}(returns, index, function)
	}
	wg.Wait()
	isSuccess := true
	for _, result := range returns {
		if result.Error() != nil {
			isSuccess = false
			break
		}
	}
	return returns, isSuccess
}

// JoinCompleteOnAnySuccess run function and return when any success, if all function return error
// then return second value equals to false, true otherwise
func JoinCompleteOnAnySuccess(funcs ...Function) ([]Return, bool) {
	finishChannel := make(chan bool)
	completeChannel := make(chan bool)
	var wg sync.WaitGroup
	returns := make([]Return, len(funcs))
	for index, function := range funcs {
		wg.Add(1)
		go func(returnValues []Return, index int, function Function) {
			returnValuesByFunction := function()
			wg.Done()
			returnValues[index] = returnValuesByFunction
			if returnValuesByFunction.Error() == nil {
				finishChannel <- true
			}
		}(returns, index, function)
	}

	go func() {
		wg.Wait()
		close(completeChannel)
	}()

	select {
	case <-completeChannel:
		return returns, existSuccessResult(returns)
	case <-finishChannel:
		return returns, true
	}
}

func existSuccessResult(returns []Return) bool {
	isSuccess := false
	for _, r := range returns {
		if r.Error() == nil {
			isSuccess = true
			break
		}
	}
	return isSuccess
}
