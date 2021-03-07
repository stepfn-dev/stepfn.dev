package main

import (
	"embed"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/glassechidna/lambdahttp/pkg/gowrap"
	"io/fs"
	"net/http"
	"os"
)

//go:embed build/*
var assets embed.FS

func main() {
	buildfs, err := fs.Sub(assets, "build")
	if err != nil {
		panic(err)
	}

	httpfs := http.FileServer(http.FS(buildfs))

	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		lambda.StartHandler(gowrap.ApiGateway(httpfs))
	} else {
		http.ListenAndServe(":8080", httpfs)
	}
}
