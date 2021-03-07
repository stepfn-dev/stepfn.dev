package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/pkg/errors"
	"os"
	"rogchap.com/v8go"
	"stepfndev/stepfndev"
)

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody),
	})
	if err != nil {
		panic(err)
	}

	h := newHandler(dynamodb.New(sess), os.Getenv("TABLE_NAME"))
	lambda.Start(h.handle)
}

type handler struct {
	iso   *v8go.Isolate
	ddb   dynamodbiface.DynamoDBAPI
	table string
}

func newHandler(ddb dynamodbiface.DynamoDBAPI, table string) *handler {
	iso, err := v8go.NewIsolate()
	if err != nil {
		panic(err)
	}

	return &handler{iso: iso, ddb: ddb, table: table}
}

type dynamoItem struct {
	Id     string `dynamodbav:"pk"`
	Script string `dynamodbav:"script"`
}

func (h *handler) handle(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	cc := stepfndev.Decode(ctx)

	script, err := h.script(cc.Id)
	if err != nil {
		return nil, err
	}

	v8ctx, err := v8go.NewContext(h.iso)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer v8ctx.Close()

	_, err = v8ctx.RunScript(script, "index.js")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	val, err := v8ctx.RunScript(fmt.Sprintf("%s(%s)", cc.Handler, string(input)), "index.js")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	str, err := v8go.JSONStringify(v8ctx, val)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return json.RawMessage(str), nil
}

func (h *handler) script(ddbId string) (string, error) {
	key := map[string]*dynamodb.AttributeValue{"pk": {S: &ddbId}}
	getResp, err := h.ddb.GetItem(&dynamodb.GetItemInput{TableName: &h.table, Key: key})
	if err != nil {
		return "", errors.WithStack(err)
	}

	item := dynamoItem{}
	err = dynamodbattribute.UnmarshalMap(getResp.Item, &item)
	if err != nil {
		return "", errors.WithStack(err)
	}

	script := item.Script
	fmt.Println(ddbId, script)
	return script, nil
}
