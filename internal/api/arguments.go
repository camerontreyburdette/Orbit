package api

import "encoding/json"

type methodArguments []json.RawMessage

func (arguments methodArguments) decode(index int, target interface{}) bool {
	if index >= len(arguments) {
		return false
	}
	return json.Unmarshal(arguments[index], target) == nil
}

func (arguments methodArguments) int64At(index int) int64 {
	var value int64
	arguments.decode(index, &value)
	return value
}

func (arguments methodArguments) intAt(index int) int {
	var value int
	arguments.decode(index, &value)
	return value
}

func (arguments methodArguments) stringAt(index int) string {
	var value string
	arguments.decode(index, &value)
	return value
}

func (arguments methodArguments) boolAt(index int) bool {
	var value bool
	arguments.decode(index, &value)
	return value
}

func (arguments methodArguments) int64SliceAt(index int) []int64 {
	var value []int64
	arguments.decode(index, &value)
	return value
}

func (arguments methodArguments) fieldsAt(index int) map[string]interface{} {
	var value map[string]interface{}
	arguments.decode(index, &value)
	return value
}
