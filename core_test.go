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

// JoinFailOnAnyErrorSuccessFailFunction

func Test_GivenSucessFuctions_WhenJoinFailOnAnyErrorSuccessFailFunction_ThenCallSuccessFunction(t *testing.T) {
	JoinFailOnAnyErrorSuccessFailFunction(func(returnValues []Return) {
		assert.Equal(t, returnValues[0].ReturnValues()[0], successValue, "JoinFailOnAnyErrorSuccessFailFunction must return expected value")
	}, func(returns []Return, err error) {
		assert.False(t, true, "JoinFailOnAnyErrorSuccessFailFunction must no call failFunction")
	}, successFunction, successFunction)
}

func Test_GivenOneFunctonFail_WhenJoinFailOnAnyErrorSuccessFailFunction_ThenCallFailFunction(t *testing.T) {
	JoinFailOnAnyErrorSuccessFailFunction(func(returnValues []Return) {
		assert.False(t, true, "JoinFailOnAnyErrorSuccessFailFunction must no call completeFunction")
	}, func(returns []Return, err error) {
		assert.EqualError(t, err, errNormal.Error(), "JoinFailOnAnyErrorSuccessFailFunction must return expected error")
	}, successFunction, errorFunction)
}

func Test_GivenOneFunctionFailFirst_WhenJoinFailOnAnyErrorSuccessFailFunction_ThenCallFailFunction(t *testing.T) {
	JoinFailOnAnyErrorSuccessFailFunction(func(returnValues []Return) {
		assert.False(t, true, "JoinFailOnAnyErrorSuccessFailFunction must no call completeFunction")
	}, func(returns []Return, err error) {
		assert.EqualError(t, err, errNormal.Error(), "JoinFailOnAnyErrorSuccessFailFunction must call FailFunction with expected error")
	}, errorFunction, errorFunctionAfter200Ms)
}

func Test_GivenFunctionDoPanic_WhenJoinFailOnAnyErrorSuccessFailFunction_ThenCallFailFunction(t *testing.T) {
	JoinFailOnAnyErrorSuccessFailFunction(func(returnValues []Return) {
		assert.False(t, true, "JoinFailOnAnyErrorSuccessFailFunction must no call completeFunction")
	}, func(returns []Return, err error) {
		assert.Error(t, err, "JoinFailOnAnyErrorSuccessFailFunction must call failFunction with expected error value")
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

// JoinCompleteAllCompleteFailFunction Tests

func Test_GivenSuccessFunctions_WhenJJoinCompleteAllSuccessFailFunction_ThenCallSuccessFunction(t *testing.T) {
	JoinCompleteAllSuccessFailFunction(func(returnValues []Return) {
		assert.True(t, true, "JoinCompleteAllSuccessFailFunction must call sucess function")
	}, func(returns []Return, err error) {
		assert.True(t, false, "JoinCompleteAllSuccessFailFunction must no call failFunction")
	}, successFunction, successFunction)
}

func Test_GivenFailFunction_WhenJoinCompleteAllSuccessFailFunction_ThenCallFailFunction(t *testing.T) {
	JoinCompleteAllSuccessFailFunction(func(returnValues []Return) {
		assert.True(t, false, "JoinCompleteAllSuccessFailFunction must no call sucessFunction")
	}, func(returns []Return, err error) {
		assert.True(t, true, "JoinCompleteAllSuccessFailFunction must call failFunction")
	}, successFunction, errorFunction)
}

func Test_GivenFunctionDoPanic_WhenJoinCompleteAllSuccessFailFunctionn_ThenCallFailFunction(t *testing.T) {
	JoinCompleteAllSuccessFailFunction(func(returnValues []Return) {
		assert.False(t, true, "JoinCompleteAllSuccessFailFunction must no call completeFunction")
	}, func(returns []Return, err error) {
		assert.Error(t, err, "JoinCompleteAllSuccessFailFunction must call failFunction with error")
	}, panicFunction)
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

// JoinCompleteOnAnySuccessSuccessFailFunction Tests

func Test_GivenSuccessFunctions_WhenJoinCompleteOnAnySuccessSuccessFailFunction_ThenReturnTrue(t *testing.T) {
	JoinCompleteOnAnySuccessSuccessFailFunction(func(returnValues []Return) {
		assert.True(t, true, "JoinCompleteOnAnySuccessSuccessFailFunction must call success function")
	}, func(returns []Return, err error) {
		assert.True(t, false, "JoinCompleteOnAnySuccessSuccessFailFunction must no call fail function")
	}, successFunction, successFunction)
}

func Test_GivenOneSuccessFunctionAndFailFunction_WhenJoinCompleteOnAnySuccessSuccessFailFunction_ThenReturnTrue(t *testing.T) {
	JoinCompleteOnAnySuccessSuccessFailFunction(func(returnValues []Return) {
		assert.True(t, true, "JoinCompleteOnAnySuccessSuccessFailFunction must call success function")
	}, func(returns []Return, err error) {
		assert.True(t, false, "JoinCompleteOnAnySuccessSuccessFailFunction must no call fail function")
	}, successFunction, errorFunction)
}

func Test_GivenFailFunctions_WhenJoinCompleteOnAnySuccessSuccessFailFunction_ThenReturnTrue(t *testing.T) {
	JoinCompleteOnAnySuccessSuccessFailFunction(func(returnValues []Return) {
		assert.True(t, false, "JoinCompleteOnAnySuccessSuccessFailFunction must no call success function")
	}, func(returns []Return, err error) {
		assert.True(t, true, "JoinCompleteOnAnySuccessSuccessFailFunction must no call fail function")
	}, errorFunction, errorFunction)
}

func Test_GivenOneFailFunctionAndOneFunctionSuccessAfter200ms_WhenJoinCompleteOnAnySuccessSuccessFailFunction_ThenReturnTrue(t *testing.T) {
	JoinCompleteOnAnySuccessSuccessFailFunction(func(returnValues []Return) {
		assert.True(t, true, "JoinCompleteOnAnySuccessSuccessFailFunction must call success function")
	}, func(returns []Return, err error) {
		assert.True(t, false, "JoinCompleteOnAnySuccessSuccessFailFunction must no call fail function")
	}, errorFunction, successFunctionAfter200Ms)
}

func Test_GivenFunctionDoPanic_WhenJoinCompleteOnAnySuccessSuccessFailFunction_ThenReturnError(t *testing.T) {
	JoinCompleteOnAnySuccessSuccessFailFunction(func(returnValues []Return) {
		assert.True(t, false, "JoinCompleteOnAnySuccessSuccessFailFunction must no call success function")
	}, func(returns []Return, err error) {
		assert.True(t, true, "JoinCompleteOnAnySuccessSuccessFailFunction must call fail function")
	}, panicFunction)
}

// getFirstError tests

func Test_GivenArrayReturnWithoutErrors_WhenGetFirstError_ThenReturnNil(t *testing.T) {
	returns := []Return{NewReturn(nil)}
	err := getFirstError(returns)
	assert.Nil(t, err, "getFirstError must return a nil error")
}

func Test_GivenArrayReturnWithErrors_WhenGetFirstError_ThenReturnError(t *testing.T) {
	returns := []Return{NewReturn(errors.New("error"))}
	err := getFirstError(returns)
	assert.Error(t, err, "getFirstError must return an error")
}

// selectWithCompleteFinishChannelCallSuccessFailFunction tests

func Test_GivenCloseCompleteChannelAndNoErrors_WhenSelectWithCompleteFinishChannelCallSuccessFailFunction_ThenCallSucessFunction(t *testing.T) {
	// Given
	returns := []Return{NewReturn(nil)}
	completeChannel := make(chan bool)
	close(completeChannel)
	finishChannel := make(chan bool)
	// When
	selectWithCompleteFinishChannelCallSuccessFailFunction(func(returns []Return) {
		// Then
		assert.True(t, true, "selectWithCompleteFinishChannelCallSuccessFailFunction must call success function")
	}, func(returns []Return, err error) {
		assert.True(t, false, "selectWithCompleteFinishChannelCallSuccessFailFunction must no call fail function")
	}, returns, completeChannel, finishChannel)
}

func Test_GivenCloseCompleteChannelAndErrors_WhenSelectWithCompleteFinishChannelCallSuccessFailFunction_ThenCallFailFunction(t *testing.T) {
	// Given
	returns := []Return{NewReturn(errors.New("error"))}
	completeChannel := make(chan bool)
	close(completeChannel)
	finishChannel := make(chan bool)
	// When
	selectWithCompleteFinishChannelCallSuccessFailFunction(func(returns []Return) {
		assert.True(t, false, "selectWithCompleteFinishChannelCallSuccessFailFunction must no call success function")
	}, func(returns []Return, err error) {
		// Then
		assert.True(t, true, "selectWithCompleteFinishChannelCallSuccessFailFunction must call fail function")
	}, returns, completeChannel, finishChannel)
}

// existSuccessResult tests

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
