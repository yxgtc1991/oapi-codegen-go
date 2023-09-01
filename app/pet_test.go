package app

import (
	"github.com/labstack/echo/v4"
	"net/http/httptest"
	"testing"
)

func TestAddPet(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()

}
