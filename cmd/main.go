package main

import (
	codegenTest "demo/oapi-codegen-go"
	"demo/oapi-codegen-go/app"
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	"net/http"
)

func main() {
	e := echo.New()
	server := app.EchoServer{}
	codegenTest.RegisterHandlersWithBaseURL(e, &server, "/james")
	// swagger 对象
	swagger, err := codegenTest.GetSwaggerWithPrefix("/james")
	if err != nil {
		panic(err)
	}
	// demo 1: 增加参数校验中间件：对 swagger 对象解析，并对请求参数校验
	// e.Use(middleware.OapiRequestValidator(swagger))
	// demo 2: 自定义参数校验
	options := middleware.Options{
		ErrorHandler: func(c echo.Context, err *echo.HTTPError) error {
			println(err.Error())
			return c.String(http.StatusOK, "bad request")
		},
	}
	e.Use(middleware.OapiRequestValidatorWithOptions(swagger, &options))
	// demo 3: Swagger UI 结合 Echo：在 echo 中配置中间件，对 Swagger 路由特殊预处理
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/james/swagger.yaml" {
				return c.File("./demo.yaml") // 返回 Swagger YAML 文件
			}
			if c.Request().URL.Path == "/james/docs" {
				return c.File("./index.html")
			}
			return next(c)
		}
	})
	// 使用 oapi-codegen 校验
	e.Use(middleware.OapiRequestValidator(swagger))
	err = e.Start(":8090")
	if err != nil {
		panic(err)
	}
}
