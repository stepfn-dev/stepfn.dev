package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sfn/sfniface"
	"github.com/oklog/ulid"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"math/rand"
	"os"
	"stepfndev/stepfndev"
	"strings"
	"time"
)

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody),
	})
	if err != nil {
		panic(err)
	}

	i := &initializer{
		sfn:     sfn.New(sess),
		ddb:     dynamodb.New(sess),
		table:   os.Getenv("TABLE_NAME"),
		role:    os.Getenv("SFN_ROLE_ARN"),
		funcArn: os.Getenv("EXECJS_FUNCTION"),
		entropy: ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0),
	}

	lambda.Start(i.handle)
}

type InitializerInput struct {
	Script     string
	Definition string
	Input      json.RawMessage
	Id         string
	Key        string
}

type InitializerOutput struct {
	StateMachineArn string
	TraceHeader     string
	TraceId         string
	Id              string
}

type initializer struct {
	sfn     sfniface.SFNAPI
	ddb     dynamodbiface.DynamoDBAPI
	table   string
	role    string
	funcArn string
	entropy io.Reader
}

func validId(id string) bool {
	return len(id) == 27 && id[0] == 'S'
}

func (i *initializer) handle(ctx context.Context, input *InitializerInput) (*InitializerOutput, error) {
	lctx, _ := lambdacontext.FromContext(ctx)
	traceId, traceHeader := traceIdAndHeader()
	fmt.Printf(`{"func":"initializer","requestId":"%s","traceId":"%s"}`+"\n", lctx.AwsRequestID, traceId)

	machineArn := ""

	if !validId(input.Id) {
		input.Id = makeId(i.entropy)

		var err error
		machineArn, err = i.createMachine(input)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		machineArn, err = i.updateMachine(input)
		if err == ErrIncorrectMachineKey {
			input.Id = makeId(i.entropy)
			machineArn, err = i.createMachine(input)
		}

		if err != nil {
			return nil, err
		}
	}

	return &InitializerOutput{
		StateMachineArn: machineArn,
		TraceHeader:     traceHeader,
		TraceId:         traceId,
		Id:              input.Id,
	}, nil
}

var ErrIncorrectMachineKey = errors.New("incorrect machine key")

func (i *initializer) updateMachine(input *InitializerInput) (string, error) {
	transformed := normalizeStateMachineDefinition(input.Definition, input.Id, i.funcArn)

	_, err := i.ddb.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: &i.table,
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {S: &input.Id},
		},
		UpdateExpression:    aws.String("SET #script = :script, #definition = :definition, #input = :input"),
		ConditionExpression: aws.String("writeKey = :writeKey"),
		ExpressionAttributeNames: map[string]*string{
			"#script":     aws.String("script"),
			"#definition": aws.String("definition"),
			"#input":      aws.String("input"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":script":     {S: &input.Script},
			":definition": {S: &input.Definition},
			":input":      {S: aws.String(string(input.Input))},
			":writeKey":   {S: aws.String(input.Key)},
		},
	})
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			if err.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
				return "", ErrIncorrectMachineKey
			}
		}
		return "", errors.WithStack(err)
	}

	machineArn := fmt.Sprintf("arn:aws:states:%s:%s:stateMachine:stepfn-%s", os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCOUNT_ID"), input.Id)
	_, err = i.sfn.UpdateStateMachine(&sfn.UpdateStateMachineInput{
		StateMachineArn: aws.String(machineArn),
		Definition:      &transformed,
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return machineArn, nil
}

func (i *initializer) createMachine(input *InitializerInput) (string, error) {
	transformed := normalizeStateMachineDefinition(input.Definition, input.Id, i.funcArn)

	_, err := i.ddb.PutItem(&dynamodb.PutItemInput{
		TableName: &i.table,
		Item: map[string]*dynamodb.AttributeValue{
			"pk":         {S: &input.Id},
			"script":     {S: &input.Script},
			"definition": {S: &input.Definition},
			"input":      {S: aws.String(string(input.Input))},
			"writeKey":   {S: &input.Key},
		},
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	resp, err := i.sfn.CreateStateMachine(&sfn.CreateStateMachineInput{
		Name:       aws.String(fmt.Sprintf("stepfn-%s", input.Id)),
		Definition: &transformed,
		RoleArn:    &i.role,
		Type:       aws.String(sfn.StateMachineTypeExpress),
	})
	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.StateMachineArn, nil
}

func makeId(entropy io.Reader) string {
	return "S" + ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func normalizeStateMachineDefinition(definition string, dynamoId string, funcArn string) string {
	transformed := definition
	// TODO: validate state names
	gjson.Get(definition, "States").ForEach(func(key, value gjson.Result) bool {
		if value.Get("Type").Str != "Task" {
			return true
		}

		if strings.HasPrefix(value.Get("Resource").Str, "arn:aws:states:::lambda:invoke") {
			handler := gjson.Get(value.Raw, "Parameters.FunctionName").Str
			cc := &stepfndev.ClientContext{Id: dynamoId, Handler: handler}
			transformedValue, _ := sjson.Set(value.Raw, "Parameters.ClientContext", cc.Encode())
			transformedValue, _ = sjson.Set(transformedValue, "Parameters.FunctionName", funcArn)
			transformedValue, _ = sjson.Delete(transformedValue, "Parameters.FunctionName\\.$")
			transformedValue, _ = sjson.Delete(transformedValue, "Parameters.Qualifier")
			transformedValue, _ = sjson.Delete(transformedValue, "Parameters.Qualifier\\.$")

			transformed, _ = sjson.Set(transformed, "States."+key.Str, json.RawMessage(transformedValue))
		} else {
			// TODO: old style tasks with lambda arn resource
		}

		return true
	})

	return transformed
}

func traceIdAndHeader() (string, string) {
	traceHeader := os.Getenv("_X_AMZN_TRACE_ID")
	m := map[string]string{}

	keyvalues := strings.Split(traceHeader, ";")
	for _, keyvalue := range keyvalues {
		split := strings.SplitN(keyvalue, "=", 2)
		key := split[0]
		value := split[1]
		m[key] = value
	}

	traceId := m["Root"]
	return traceId, traceHeader
}
