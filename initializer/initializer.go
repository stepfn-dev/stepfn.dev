package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
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
		entropy: ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0),
	}

	lambda.Start(i.handle)
}

type InitializerInput struct {
	Script     string
	Definition string
	Input      json.RawMessage
}

type InitializerOutput struct {
	StateMachineArn string
}

type initializer struct {
	sfn     sfniface.SFNAPI
	ddb     dynamodbiface.DynamoDBAPI
	table   string
	role    string
	entropy io.Reader
	funcArn string
}

func (i *initializer) handle(ctx context.Context, input *InitializerInput) (*InitializerOutput, error) {
	id := "S" + ulid.MustNew(ulid.Timestamp(time.Now()), i.entropy).String()
	_, err := i.ddb.PutItem(&dynamodb.PutItemInput{
		TableName: &i.table,
		Item: map[string]*dynamodb.AttributeValue{
			"pk":     {S: &id},
			"script": {S: &input.Script},
		},
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	transformed := input.Definition
	// TODO: validate state names
	gjson.Get(input.Definition, "States").ForEach(func(key, value gjson.Result) bool {
		if value.Get("Type").Str != "Task" {
			return true
		}

		if strings.HasPrefix(value.Get("Resource").Str, "arn:aws:states:::lambda:invoke") {
			handler := gjson.Get(value.Raw, "Parameters.FunctionName").Str
			cc := &stepfndev.ClientContext{Id: id, Handler: handler}
			transformedValue, _ := sjson.Set(value.Raw, "Parameters.ClientContext", cc.Encode())
			transformedValue, _ = sjson.Set(transformedValue, "Parameters.FunctionName", i.funcArn)
			transformedValue, _ = sjson.Delete(transformedValue, "Parameters.FunctionName\\.$")
			transformedValue, _ = sjson.Delete(transformedValue, "Parameters.Qualifier")
			transformedValue, _ = sjson.Delete(transformedValue, "Parameters.Qualifier\\.$")

			transformed, _ = sjson.Set(transformed, "States."+key.Str, json.RawMessage(transformedValue))
		} else {
			// TODO: old style tasks with lambda arn resource
		}

		return true
	})

	fmt.Println(transformed)

	resp, err := i.sfn.CreateStateMachine(&sfn.CreateStateMachineInput{
		Name:       aws.String(fmt.Sprintf("stepfn-%s", id)),
		Definition: &transformed,
		RoleArn:    &i.role,
		Type:       aws.String(sfn.StateMachineTypeExpress),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &InitializerOutput{StateMachineArn: *resp.StateMachineArn}, nil
}
