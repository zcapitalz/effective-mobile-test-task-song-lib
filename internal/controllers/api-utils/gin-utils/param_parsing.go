package ginutils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func CreateParamParsingMiddleware(
	requestParamName,
	contextParamName string,
	parser func(param string) (any, error),
) func(c *gin.Context) {

	return func(c *gin.Context) {
		parsedParam, err := parser(c.Param(requestParamName))
		if err != nil {
			BindURIError(c, fmt.Errorf("parse %s: %s", requestParamName, err))
			return
		}

		c.Set(contextParamName, parsedParam)
		c.Next()
	}
}
