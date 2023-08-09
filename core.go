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
type SuccessFunction func(returns []Return)
type FailFunction func(returns []Return, err error)

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
	runFunctionFailOnAnyError(returns, errorChannel, &wg, funcs)

	go waitAndCloseChannel(&wg, completeChannel)

	return selectWithCompleteErrorChannel(returns, completeChannel, errorChannel)
}

// JoinFailOnAnyErrorSuccessFailFunction Run functions and execute successFunction if success or call failFunction if any function fail
func JoinFailOnAnyErrorSuccessFailFunction(successFunction SuccessFunction, failFunction FailFunction, funcs ...Function) {
	errorChannel := make(chan error)
	completeChannel := make(chan bool)
	var wg sync.WaitGroup
	returns := make([]Return, len(funcs))
	runFunctionFailOnAnyError(returns, errorChannel, &wg, funcs)

	go waitAndCloseChannel(&wg, completeChannel)

	select {
	case <-completeChannel:
		successFunction(returns)
	case err := <-errorChannel:
		failFunction(returns, err)
	}
}

func runFunctionFailOnAnyError(returns []Return, errorChannel chan error, wg *sync.WaitGroup, funcs []Function) {
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
}

func selectWithCompleteErrorChannel(returns []Return, completeChannel chan bool, errorChannel chan error) ([]Return, error) {
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

// JoinCompleteAllSuccessFailFunction Run functions and call complete functions if success or
// call failFunction if any fail
func JoinCompleteAllSuccessFailFunction(successFunction SuccessFunction, failFunction FailFunction, funcs ...Function) {
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
	var err error = nil
	for _, result := range returns {
		if result.Error() != nil {
			isSuccess = false
			err = result.Error()
			break
		}
	}

	if isSuccess {
		successFunction(returns)
	} else {
		failFunction(returns, err)
	}
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

	return selectWithCompleteFinishChannel(returns, completeChannel, finishChannel)
}

func JoinCompleteOnAnySuccessSuccessFailFunction(successFunction SuccessFunction, failFunction FailFunction, funcs ...Function) {
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

	selectWithCompleteFinishChannelCallSuccessFailFunction(successFunction, failFunction, returns, completeChannel, finishChannel)
}

func selectWithCompleteFinishChannelCallSuccessFailFunction(successFunction SuccessFunction, failFunction FailFunction, returns []Return, completeChannel chan bool, finishChannel chan bool) {
	select {
	case <-completeChannel:
		if existSuccessResult(returns) {
			successFunction(returns)
		} else {
			failFunction(returns, getFirstError(returns))
		}
	case <-finishChannel:
		successFunction(returns)
	}
}

func selectWithCompleteFinishChannel(returns []Return, completeChannel chan bool, finishChannel chan bool) ([]Return, bool) {
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

func getFirstError(returns []Return) error {
	for _, r := range returns {
		if r.Error() != nil {
			return r.Error()
		}
	}
	return nil
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

func JoinFailOnErrorOrTimeoutSuccessFailFunction(successFunction SuccessFunction, failFunction FailFunction, duration time.Duration, funcs ...Function) {
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

	selectWithCompleteErrorChannelAndTimerSuccessFailFunction(successFunction, failFunction, returns, completeChannel, errorChannel, duration)
}

func selectWithCompleteErrorChannelAndTimerSuccessFailFunction(successFunction SuccessFunction, failFunction FailFunction, returns []Return, completeChannel chan bool, errorChannel chan error, duration time.Duration) {
	select {
	case <-completeChannel:
		successFunction(returns)
	case err := <-errorChannel:
		failFunction(returns, err)
	case <-time.After(duration):
		failFunction(returns, ErrTimeout)
	}
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
