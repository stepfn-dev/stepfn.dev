package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/glassechidna/lambdahttp/pkg/gowrap"
	"net/http"
	"os"
)

func main() {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody),
	})
	if err != nil {
		panic(err)
	}

	ddb := dynamodb.New(sess)
	table := os.Getenv("TABLE_NAME")

	http.HandleFunc("/sfn", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		key := map[string]*dynamodb.AttributeValue{"pk": {S: &id}}
		getResp, err := ddb.GetItem(&dynamodb.GetItemInput{TableName: &table, Key: key})
		if err != nil || getResp.Item == nil {
			http.NotFound(w, r)
			return
		}

		it := getResp.Item
		msi := map[string]interface{}{
			"Script":     *it["script"].S,
			"Definition": *it["definition"].S,
			"Input":      *it["input"].S,
		}

		j, _ := json.Marshal(msi)
		w.Header().Set("Content-type", "application/json")
		w.Write(j)
	})

	lambda.StartHandler(gowrap.ApiGateway(http.DefaultServeMux))
}
