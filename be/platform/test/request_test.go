package test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
)

func TestRequestHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/items/:item_id", func(c *gin.Context) {
		if got := request.Param(c, "item_id"); got != "item-1" {
			t.Fatalf("Param() = %q", got)
		}
		if got := request.BearerToken(c); got != "token" {
			t.Fatalf("BearerToken() = %q", got)
		}
		params, err := request.ListQuery(c)
		if err != nil || params.Limit != 30 || params.Offset != 2 {
			t.Fatalf("ListQuery() = %#v, %v", params, err)
		}
	})
	request := httptest.NewRequest("GET", "/items/item-1?limit=30&offset=2", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("status = %d", response.Code)
	}
}
