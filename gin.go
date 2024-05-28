package jsonschema

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func ValidateMiddleware(schema any) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data map[string]any
		if err := c.ShouldBindBodyWith(&data, binding.JSON); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}

		if err := Validate(data, schema); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
		}

		c.Next()
	}
}
