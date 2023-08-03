package gauss

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	errNormal     = errors.New("err-normal")
	errAfter200Ms = errors.New("err-after-200ms")
	successValue  = "Value"
)

func successFunction() Return {
	return NewReturn(nil, successValue)
}

func errorFunction() Return {
	return NewReturn(errNormal)
}

func errorFunctionAfter200Ms() Return {
	time.Sleep(time.Millisecond * 200)
	return NewReturn(errAfter200Ms)
}

func successFunctionAfter200Ms() Return {
	time.Sleep(time.Millisecond * 200)
	return NewReturn(nil, successValue)
}

func panicFunction() Return {
	panic("panic")
}

// JoinFailOnAnyError Tests

func Test_GivenSucessFuctions_WhenJoinFailOnAnyError_ThenReturnNillError(t *testing.T) {
	returnValues, err := JoinFailOnAnyError(successFunction, successFunction)
	assert.Nil(t, err, "JoinFailOnAnyError must return nil error")
	assert.Equal(t, returnValues[0].ReturnValues()[0], successValue)
}

func Test_GivenOneFunctonFail_WhenJoinFailOnAnyError_ThenReturnExpectedError(t *testing.T) {
	_, err := JoinFailOnAnyError(successFunction, errorFunction)
	assert.EqualError(t, err, errNormal.Error(), "JoinFailOnAnyError must return expected error")
}

func Test_GivenOneFunctionFailFirst_WhenJoinFailOnAnyError_ThenReturnExpectedError(t *testing.T) {
	_, err := JoinFailOnAnyError(errorFunction, errorFunctionAfter200Ms)
	assert.EqualError(t, err, errNormal.Error())
}

func Test_GivenFunctionDoPanic_WhenJoinFailOnAnyError_ThenReturnError(t *testing.T) {
	_, err := JoinFailOnAnyError(panicFunction)
	assert.Error(t, err)
}

// JoinFailOnAnyErrorCompleteFailFunction

func Test_GivenSucessFuctions_WhenJoinFailOnAnyErrorCompleteFailFunction_ThenCallCompleteFunction(t *testing.T) {
	JoinFailOnAnyErrorCompleteFailFunction(func(returnValues []Return) {
		assert.Equal(t, returnValues[0].ReturnValues()[0], successValue, "JoinFailOnAnyErrorCompleteFailFunction must return expected value")
	}, func(returns []Return, err error) {
		assert.False(t, true, "JoinFailOnAnyErrorCompleteFailFunction must no call failFunction")
	}, successFunction, successFunction)
}

func Test_GivenOneFunctonFail_WhenJJoinFailOnAnyErrorCompleteFailFunction_ThenCallFailFunction(t *testing.T) {
	JoinFailOnAnyErrorCompleteFailFunction(func(returnValues []Return) {
		assert.False(t, true, "JoinFailOnAnyErrorCompleteFailFunction must no call completeFunction")
	}, func(returns []Return, err error) {
		assert.EqualError(t, err, errNormal.Error(), "JoinFailOnAnyError must return expected error")
	}, successFunction, errorFunction)
}

func Test_GivenOneFunctionFailFirst_WhenJoinFailOnAnyErrorCompleteFailFunction_ThenCallFailFunction(t *testing.T) {
	JoinFailOnAnyErrorCompleteFailFunction(func(returnValues []Return) {
		assert.False(t, true, "JoinFailOnAnyErrorCompleteFailFunction must no call completeFunction")
	}, func(returns []Return, err error) {
		assert.EqualError(t, err, errNormal.Error(), "JoinFailOnAnyErrorCompleteFailFunction must call FailFunction with expected error")
	}, errorFunction, errorFunctionAfter200Ms)
}

func Test_GivenFunctionDoPanic_WhenJoinFailOnAnyErrorCompleteFailFunction_ThenCallFailFunction(t *testing.T) {
	JoinFailOnAnyErrorCompleteFailFunction(func(returnValues []Return) {
		assert.False(t, true, "JoinFailOnAnyErrorCompleteFailFunction must no call completeFunction")
	}, func(returns []Return, err error) {
		assert.Error(t, err, "JoinFailOnAnyErrorCompleteFailFunction must call failFunction with expected error value")
	}, panicFunction)
}

// JoinCompleteAll Tests

func Test_GivenSuccessFunctions_WhenJoinCompleteAll_ThenReturnTrue(t *testing.T) {
	_, isSuccess := JoinCompleteAll(successFunction, successFunction)
	assert.True(t, isSuccess, "JoinCompleteAll must return second value equals to true")
}

func Test_GivenFailFunction_WhenJoinCompleteAll_ThenReturnFalse(t *testing.T) {
	_, isSuccess := JoinCompleteAll(successFunction, errorFunction)
	assert.False(t, isSuccess, "JoinCompleteAll must return second value equals to false")
}

func Test_GivenFunctionDoPanic_WhenJoinCompleteAll_ThenReturnError(t *testing.T) {
	_, isSuccess := JoinCompleteAll(panicFunction)
	assert.False(t, isSuccess)
}

// JoinCompleteOnAnySuccess Tests

func Test_GivenSuccessFunctions_WhenJoinCompleteOnAnySuccess_ThenReturnTrue(t *testing.T) {
	_, isSuccess := JoinCompleteOnAnySuccess(successFunction, successFunction)
	assert.True(t, isSuccess, "JoinCompleteOnAnySuccess must return second value equals to true")
}

func Test_GivenOneSuccessFunctionAndFailFunction_WhenJoinCompleteOnAnySuccess_ThenReturnTrue(t *testing.T) {
	_, isSuccess := JoinCompleteOnAnySuccess(successFunction, errorFunction)
	assert.True(t, isSuccess, "JoinCompleteOnAnySuccess must return second value equals to true")
}

func Test_GivenFailFunctions_WhenJoinCompleteOnAnySuccess_ThenReturnTrue(t *testing.T) {
	_, isSuccess := JoinCompleteOnAnySuccess(errorFunction, errorFunction)
	assert.False(t, isSuccess, "JoinCompleteOnAnySuccess must return second value equals to false")
}

func Test_GivenOneFailFunctionAndOneFunctionSuccessAfter200ms_WhenJoinCompleteOnAnySuccess_ThenReturnTrue(t *testing.T) {
	_, isSuccess := JoinCompleteOnAnySuccess(errorFunction, successFunctionAfter200Ms)
	assert.True(t, isSuccess, "JoinCompleteOnAnySuccess must return second value equals to true")
}

func Test_GivenFunctionDoPanic_WhenJoinCompleteOnAnySuccess_ThenReturnError(t *testing.T) {
	_, isSuccess := JoinCompleteOnAnySuccess(panicFunction)
	assert.False(t, isSuccess)
}

func TestOneReturnFailAndOneSuccess_WhenExistSuccessResult_ThenReturnTrue(t *testing.T) {
	resturns := []Return{NewReturn(nil), NewReturn(errors.New("Error"))}
	assert.True(t, existSuccessResult(resturns))
}

// JoinFailOnErrorOrTimeout tests

func Test_GivenSucessFuctionAfter200ms_WhenJoinFailOnErrorOrTimeoutWithTimeout300Ms_ThenReturnNilError(t *testing.T) {
	_, err := JoinFailOnErrorOrTimeout(300*time.Millisecond, successFunctionAfter200Ms)
	assert.Nil(t, err, "JoinFailOnErrorOrTimeout must return a nil error")
}

func Test_GivenSucessFuctionAfter200ms_WhenJoinFailOnErrorOrTimeoutWithTimeout100Ms_ThenReturnError(t *testing.T) {
	_, err := JoinFailOnErrorOrTimeout(100*time.Millisecond, successFunctionAfter200Ms)
	assert.Error(t, err, "JoinFailOnErrorOrTimeout must return an error")
}

func Test_GivenErrorFuction_WhenJoinFailOnErrorOrTimeoutWithTimeout100Ms_ThenReturnError(t *testing.T) {
	_, err := JoinFailOnErrorOrTimeout(100*time.Millisecond, errorFunction)
	assert.Error(t, err, "JoinFailOnErrorOrTimeout must return an error")
}
