// gauss contains utilities to execute
package gauss

import (
	"sync"
)

type Result struct {
	Err    error
	Result []interface{}
}

func (_self *Result) AddResult(result interface{}) {
	_self.Result = append(_self.Result, result)
}

type Function func(result *Result)

func JoinFailOnAnyError(funcs ...Function) ([]Result, error) {
	errorChannel := make(chan error)
	completeChannel := make(chan bool)
	var wg sync.WaitGroup
	execResults := make([]Result, len(funcs))
	for index, function := range funcs {
		wg.Add(1)
		execResults[index] = Result{Result: make([]interface{}, 0)}
		go func(execResult *Result, f Function) {
			f(execResult)
			wg.Done()
			if execResult.Err != nil {
				errorChannel <- execResult.Err
			}
		}(&execResults[index], function)
	}

	go func() {
		wg.Wait()
		close(completeChannel)
	}()

	select {
	case <-completeChannel:
		return execResults, nil
	case err := <-errorChannel:
		return execResults, err
	}
}
