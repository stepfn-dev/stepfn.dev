package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"stepfndev/stepfndev"
	"strings"
)

func normalizeStateMachineDefinition(definition string, dynamoId string, funcArn string) string {
	transformed := definition
	// TODO: validate state names
	gjson.Get(definition, "States").ForEach(func(key, value gjson.Result) bool {
		newValue := value.Raw
		switch value.Get("Type").Str {
		case "Task":
			newValue = normalizeTaskState(value, dynamoId, funcArn)
		case "Map":
			iterator := normalizeStateMachineDefinition(value.Get("Iterator").Raw, dynamoId, funcArn)
			newValue, _ = sjson.Set(value.Raw, "Iterator", json.RawMessage(iterator))
		case "Parallel":
			idx := 0
			value.Get("Branches").ForEach(func(_, branch gjson.Result) bool {
				normalized := normalizeStateMachineDefinition(branch.Raw, dynamoId, funcArn)
				newValue, _ = sjson.Set(newValue, fmt.Sprintf("Branches.%d", idx), json.RawMessage(normalized))
				idx++
				return true
			})
		}

		transformed, _ = sjson.Set(transformed, "States."+key.Str, json.RawMessage(newValue))
		return true
	})

	return transformed
}

func normalizeTaskState(value gjson.Result, dynamoId, funcArn string) string {
	if strings.HasPrefix(value.Get("Resource").Str, "arn:aws:states:::lambda:invoke") {
		handler := gjson.Get(value.Raw, "Parameters.FunctionName").Str
		cc := &stepfndev.ClientContext{Id: dynamoId, Handler: handler}
		transformedValue, _ := sjson.Set(value.Raw, "Parameters.ClientContext", cc.Encode())
		transformedValue, _ = sjson.Set(transformedValue, "Parameters.FunctionName", funcArn)
		transformedValue, _ = sjson.Delete(transformedValue, "Parameters.FunctionName\\.$")
		transformedValue, _ = sjson.Delete(transformedValue, "Parameters.Qualifier")
		transformedValue, _ = sjson.Delete(transformedValue, "Parameters.Qualifier\\.$")

		return transformedValue
	} else {
		// TODO: old style tasks with lambda arn resource
		return ""
	}
}
