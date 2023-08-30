package app

import (
	codegenTest "demo/oapi-codegen-go"
	"github.com/labstack/echo/v4"
	"net/http"
)

type EchoServer struct{}

func (e *EchoServer) FindPets(ctx echo.Context, params codegenTest.FindPetsParams) error {
	// todo 业务逻辑
	return ctx.JSON(http.StatusOK, "find list")
}

func (e *EchoServer) AddPet(ctx echo.Context) error {
	// todo 业务逻辑
	return ctx.JSON(http.StatusOK, "add")
}

func (e *EchoServer) DeletePet(ctx echo.Context, id int64) error {
	// todo 业务逻辑
	return ctx.JSON(http.StatusOK, "delete")
}

func (e *EchoServer) FindPetById(ctx echo.Context, id int64) error {
	// todo 业务逻辑
	return ctx.JSON(http.StatusOK, "find single")
}
