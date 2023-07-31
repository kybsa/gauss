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
)

func successFunction(result *Result) {
	result.Err = nil
	result.AddResult("Value")
}

func errorFunction(result *Result) {
	result.Err = errNormal
}

func errorFunctionAfter200Ms(result *Result) {
	result.Err = errAfter200Ms
	time.After(time.Millisecond * 200)
}

func Test_GivenSucessFuction_WhenJoinFailOnAnyError_ThenReturnNillError(t *testing.T) {
	_, err := JoinFailOnAnyError(successFunction, successFunction)
	assert.Nil(t, err, "JoinFailOnAnyError must return nil error")
}

func Test_GivenOneFunctonFail_WhenJoinFailOnAnyError_ThenReturnExpectedError(t *testing.T) {
	_, err := JoinFailOnAnyError(successFunction, errorFunction)
	assert.EqualError(t, err, errNormal.Error(), "JoinFailOnAnyError must return expected error")
}

func Test_GivenOneFunctionFailFirst_WhenJoinFailOnAnyError_ThenReturnExpectedError(t *testing.T) {
	_, err := JoinFailOnAnyError(errorFunction, errorFunctionAfter200Ms)
	assert.EqualError(t, err, errNormal.Error())
}
