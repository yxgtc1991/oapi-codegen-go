package app

import (
	. "demo/oapi-codegen-go"
	"github.com/labstack/echo/v4"
	"net/http"
)

type EchoServer struct{}

func (e *EchoServer) FindPets(ctx echo.Context, params FindPetsParams) error {
	// todo 业务逻辑
	return ctx.JSON(http.StatusOK, "find list")
}

func (e *EchoServer) AddPet(ctx echo.Context) error {
	pet := AddPetJSONRequestBody{}
	_ = ctx.Bind(&pet)
	resp := Pet{
		Id:   1,
		Name: pet.Name,
		Tag:  pet.Tag,
	}
	return ctx.JSON(http.StatusOK, resp)
}

func (e *EchoServer) DeletePet(ctx echo.Context, id int64) error {
	// todo 业务逻辑
	return ctx.JSON(http.StatusOK, "delete")
}

func (e *EchoServer) FindPetById(ctx echo.Context, id int64) error {
	// todo 业务逻辑
	return ctx.JSON(http.StatusOK, "find single")
}
