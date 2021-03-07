package stepfndev

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

type ClientContext struct {
	Id      string
	Handler string
}

func (cc *ClientContext) Encode() string {
	msi := map[string]interface{}{"custom": cc}
	j, _ := json.Marshal(msi)
	return base64.StdEncoding.EncodeToString(j)
}

func Decode(ctx context.Context) *ClientContext {
	lctx, _ := lambdacontext.FromContext(ctx)
	m := lctx.ClientContext.Custom
	return &ClientContext{
		Id:      m["Id"],
		Handler: m["Handler"],
	}
}
