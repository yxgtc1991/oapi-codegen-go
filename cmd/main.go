package main

import (
	codegen_test "demo/oapi-codegen-go"
	"demo/oapi-codegen-go/app"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	server := app.EchoServer{}
	codegen_test.RegisterHandlersWithBaseURL(e, &server, "")
	err := e.Start(":8090")
	if err != nil {
		panic(err)
	}
}
