package response

import "github.com/gin-gonic/gin"

type Success struct {
	statusCode int
	message    string
}

func NewSuccess(message string, status int) Success {
	return Success{
		statusCode: status,
		message:    message,
	}
}

func (s Success) JSON(c *gin.Context) {
	c.JSON(s.statusCode, gin.H{
		"message": s.message,
	})
}
