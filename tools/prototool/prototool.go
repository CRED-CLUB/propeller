package main

import (
	"context"

	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/CRED-CLUB/propeller/tools/prototool/java"
)

func main() {
	ctx := context.Background()
	_, err := logger.NewLogger("dev", make(map[string]interface{}), nil)
	if err != nil {
		logger.Ctx(ctx).Fatal(err)
	}
	javaHandler, err := java.NewHandler()
	if err != nil {
		logger.Ctx(ctx).Fatal(err)
	}
	err = javaHandler.BuildAndPublish(ctx, "0.0.3")
	if err != nil {
		logger.Ctx(ctx).Fatal(err.Error())
	}
}
