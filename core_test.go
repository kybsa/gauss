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
	return NewReturn(nil)
}

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

func Test_GivenSuccessFunctions_WhenJoinCompleteAll_ThenReturnTrue(t *testing.T) {
	_, isSuccess := JoinCompleteAll(successFunction, successFunction)
	assert.True(t, isSuccess, "JoinCompleteAll must return second value equals to true")
}

func Test_GivenFailFunction_WhenJoinCompleteAll_ThenReturnFalse(t *testing.T) {
	_, isSuccess := JoinCompleteAll(successFunction, errorFunction)
	assert.False(t, isSuccess, "JoinCompleteAll must return second value equals to false")
}

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
