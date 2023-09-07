package main

import (
	codegenTest "demo/oapi-codegen-go"
	"demo/oapi-codegen-go/app"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"net/http"
)

const test = "Get order info failed: %v."

func main() {
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)
	server := app.EchoServer{}
	codegenTest.RegisterHandlersWithBaseURL(e, &server, "/james")
	// swagger 对象
	swagger, err := codegenTest.GetSwaggerWithPrefix("/james")
	if err != nil {
		panic(err)
	}
	// 增加参数校验中间件：对 swagger 对象解析，并对请求参数校验
	// e.Use(middleware.OapiRequestValidator(swagger))
	// 自定义参数校验
	options := middleware.Options{
		ErrorHandler: func(c echo.Context, err *echo.HTTPError) error {
			println(err.Error())
			return c.String(http.StatusOK, "bad request")
		},
	}
	e.Use(middleware.OapiRequestValidatorWithOptions(swagger, &options))
	err = e.Start(":8090")
	if err != nil {
		panic(err)
	}
}
