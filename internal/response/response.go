package response

import "github.com/gin-gonic/gin"

func body(status, message string, data []interface{}) gin.H {
	h := gin.H{"status": status, "message": message}
	if len(data) > 0 && data[0] != nil {
		h["data"] = data[0]
	}
	return h
}

func Success(c *gin.Context, code int, message string, data ...interface{}) {
	c.JSON(code, body("success", message, data))
}

func Fail(c *gin.Context, code int, message string, data ...interface{}) {
	c.JSON(code, body("fail", message, data))
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, body("error", message, nil))
}

func AbortFail(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, body("fail", message, nil))
}

func AbortError(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, body("error", message, nil))
}
