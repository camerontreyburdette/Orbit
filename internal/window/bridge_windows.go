//go:build windows

package window

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var errorInterfaceType = reflect.TypeOf((*error)(nil)).Elem()

type remoteProcedureCallMessageStructure struct {
	Identifier int               `json:"identifier"`
	Method     string            `json:"method"`
	Parameters []json.RawMessage `json:"parameters"`
}

type boundFunction struct {
	value          reflect.Value
	functionType   reflect.Type
	isVariadic     bool
	parameterCount int
}

func newBoundFunction(function interface{}) (boundFunction, error) {
	functionValue := reflect.ValueOf(function)
	if functionValue.Kind() != reflect.Func {
		return boundFunction{}, errors.New("only functions can be bound")
	}
	functionType := functionValue.Type()
	if functionType.NumOut() > 2 {
		return boundFunction{}, errors.New("function may only return a value or a value and error")
	}
	return boundFunction{
		value:          functionValue,
		functionType:   functionType,
		isVariadic:     functionType.IsVariadic(),
		parameterCount: functionType.NumIn(),
	}, nil
}

func (bound boundFunction) parameterType(parameterIndex int) reflect.Type {
	if bound.isVariadic && parameterIndex >= bound.parameterCount-1 {
		return bound.functionType.In(bound.parameterCount - 1).Elem()
	}
	return bound.functionType.In(parameterIndex)
}

func (bound boundFunction) hasMatchingArity(parameterCount int) bool {
	if bound.isVariadic {
		return parameterCount >= bound.parameterCount-1
	}
	return parameterCount == bound.parameterCount
}

func (bound boundFunction) decodeArguments(parameters []json.RawMessage) ([]reflect.Value, error) {
	arguments := make([]reflect.Value, 0, len(parameters))
	for parameterIndex, rawParameter := range parameters {
		argument := reflect.New(bound.parameterType(parameterIndex))
		if unmarshalError := json.Unmarshal(rawParameter, argument.Interface()); unmarshalError != nil {
			return nil, unmarshalError
		}
		arguments = append(arguments, argument.Elem())
	}
	return arguments, nil
}

func interpretReturnValues(returnValues []reflect.Value) (interface{}, error) {
	switch len(returnValues) {
	case 0:
		return nil, nil
	case 1:
		if !returnValues[0].Type().Implements(errorInterfaceType) {
			return returnValues[0].Interface(), nil
		}
		if returnValues[0].Interface() != nil {
			return nil, returnValues[0].Interface().(error)
		}
		return nil, nil
	case 2:
		if !returnValues[1].Type().Implements(errorInterfaceType) {
			return nil, errors.New("second return value must be an error")
		}
		if returnValues[1].Interface() == nil {
			return returnValues[0].Interface(), nil
		}
		return returnValues[0].Interface(), returnValues[1].Interface().(error)
	}
	return nil, errors.New("unexpected number of return values")
}

func (bound boundFunction) invoke(parameters []json.RawMessage) (interface{}, error) {
	if !bound.hasMatchingArity(len(parameters)) {
		return nil, errors.New("function arguments mismatch")
	}
	arguments, decodeError := bound.decodeArguments(parameters)
	if decodeError != nil {
		return nil, decodeError
	}
	return interpretReturnValues(bound.value.Call(arguments))
}

func bindingScriptFor(name string) string {
	encodedName, _ := json.Marshal(name)
	return fmt.Sprintf(`(function() {
		var name = %s;
		var remoteProcedureCall = window._rpc = (window._rpc || {nextSequenceNumber: 1});
		window[name] = function() {
			var sequenceNumber = remoteProcedureCall.nextSequenceNumber++;
			var promise = new Promise(function(resolve, reject) {
				remoteProcedureCall[sequenceNumber] = { resolve: resolve, reject: reject };
			});
			window.external.invoke(JSON.stringify({
				identifier: sequenceNumber,
				method: name,
				parameters: Array.prototype.slice.call(arguments)
			}));
			return promise;
		};
	})();`, string(encodedName))
}

func (windowInstance *WindowsAppWindow) Bind(name string, function interface{}) error {
	bound, bindError := newBoundFunction(function)
	if bindError != nil {
		return bindError
	}

	windowInstance.mutex.Lock()
	windowInstance.bindings[name] = bound
	windowInstance.mutex.Unlock()

	windowInstance.Init(bindingScriptFor(name))
	return nil
}

func (windowInstance *WindowsAppWindow) settleRemoteCall(requestIdentifier string, outcome string, encodedPayload []byte) {
	windowInstance.Dispatch(func() {
		windowInstance.Eval(fmt.Sprintf("window._rpc[%s].%s(%s); window._rpc[%s] = undefined", requestIdentifier, outcome, string(encodedPayload), requestIdentifier))
	})
}

func (windowInstance *WindowsAppWindow) handleMessageCallback(message string) {
	var request remoteProcedureCallMessageStructure
	if unmarshalError := json.Unmarshal([]byte(message), &request); unmarshalError != nil {
		return
	}

	requestIdentifier := strconv.Itoa(request.Identifier)
	result, executionError := windowInstance.executeBoundMethod(request)
	if executionError != nil {
		encodedError, _ := json.Marshal(executionError.Error())
		windowInstance.settleRemoteCall(requestIdentifier, "reject", encodedError)
		return
	}

	encodedResult, marshalError := json.Marshal(result)
	if marshalError != nil {
		encodedError, _ := json.Marshal(marshalError.Error())
		windowInstance.settleRemoteCall(requestIdentifier, "reject", encodedError)
		return
	}

	windowInstance.settleRemoteCall(requestIdentifier, "resolve", encodedResult)
}

func (windowInstance *WindowsAppWindow) executeBoundMethod(request remoteProcedureCallMessageStructure) (interface{}, error) {
	windowInstance.mutex.Lock()
	bound, exists := windowInstance.bindings[request.Method]
	windowInstance.mutex.Unlock()

	if !exists {
		return nil, nil
	}
	return bound.invoke(request.Parameters)
}
