// Package gauss contains utilities to execute
package gauss

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrTimeout = errors.New("timeout")
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

func sendErrorToChannelOnPanic(errorChannel chan error) {
	if r := recover(); r != nil {
		errorChannel <- fmt.Errorf("%v", r)
	}
}

func writeResultOnPanic(result []Return, index int, wg *sync.WaitGroup) {
	if r := recover(); r != nil {
		result[index] = NewReturn(fmt.Errorf("%v", r))
		wg.Done()
	}
}

// JoinFailOnAnyError Run functions and return when any function fail
func JoinFailOnAnyError(funcs ...Function) ([]Return, error) {
	errorChannel := make(chan error)
	completeChannel := make(chan bool)
	var wg sync.WaitGroup
	returns := make([]Return, len(funcs))
	for index, function := range funcs {
		wg.Add(1)
		go func(returnValues []Return, index int, function Function) {
			defer sendErrorToChannelOnPanic(errorChannel)
			returnValuesByFunction := function()
			returnValues[index] = returnValuesByFunction
			if returnValuesByFunction.Error() != nil {
				errorChannel <- returnValuesByFunction.Error()
			}
			wg.Done()
		}(returns, index, function)
	}

	go waitAndCloseChannel(&wg, completeChannel)

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
		go func(returnValues []Return, index int, function Function, waitGroup *sync.WaitGroup) {
			defer writeResultOnPanic(returnValues, index, waitGroup)
			returnValuesByFunction := function()
			returnValues[index] = returnValuesByFunction
			wg.Done()
		}(returns, index, function, &wg)
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
		go func(returnValues []Return, index int, function Function, waitGroup *sync.WaitGroup) {
			defer writeResultOnPanic(returnValues, index, waitGroup)
			returnValuesByFunction := function()
			returnValues[index] = returnValuesByFunction
			if returnValuesByFunction.Error() == nil {
				finishChannel <- true
			}
			wg.Done()
		}(returns, index, function, &wg)
	}

	go waitAndCloseChannel(&wg, completeChannel)

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

// JoinFailOnErrorOrTimeout Run functions and return when complete or fail if a function fail or timeout
func JoinFailOnErrorOrTimeout(duration time.Duration, funcs ...Function) ([]Return, error) {
	errorChannel := make(chan error)
	completeChannel := make(chan bool)
	var wg sync.WaitGroup
	returns := make([]Return, len(funcs))
	for index, function := range funcs {
		wg.Add(1)
		go func(returnValues []Return, index int, function Function) {
			defer sendErrorToChannelOnPanic(errorChannel)
			returnValuesByFunction := function()
			returnValues[index] = returnValuesByFunction
			if returnValuesByFunction.Error() != nil {
				errorChannel <- returnValuesByFunction.Error()
			}
			wg.Done()
		}(returns, index, function)
	}

	go waitAndCloseChannel(&wg, completeChannel)

	return selectWithCompleteErrorChannelAndTimer(returns, completeChannel, errorChannel, duration)
}

func selectWithCompleteErrorChannelAndTimer(returns []Return, completeChannel chan bool, errorChannel chan error, duration time.Duration) ([]Return, error) {
	select {
	case <-completeChannel:
		return returns, nil
	case err := <-errorChannel:
		return returns, err
	case <-time.After(duration):
		return returns, ErrTimeout
	}
}

func waitAndCloseChannel(wg *sync.WaitGroup, completeChannel chan bool) {
	wg.Wait()
	close(completeChannel)
}
