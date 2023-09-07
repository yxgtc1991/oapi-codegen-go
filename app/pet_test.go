package app

import (
	codegenTest "demo/oapi-codegen-go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestAddPet(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	codegenTest.RegisterHandlersWithBaseURL(e, &EchoServer{}, "/echo_test")
	tag := "tag1"
	body := codegenTest.AddPetJSONRequestBody{
		Name: "baby",
		Tag:  &tag,
	}
	request, _ := codegenTest.NewAddPetRequest("/echo_test/", body)
	e.ServeHTTP(recorder, request)
	response, err := codegenTest.ParseAddPetResponse(recorder.Result())
	assert.Nil(t, err)
	assert.Equal(t, body.Name, response.JSON200.Name)
}
