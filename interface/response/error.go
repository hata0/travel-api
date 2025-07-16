package response

import "github.com/gin-gonic/gin"

type Error struct {
	statusCode int
	message    string
}

func NewError(err error, status int) Error {
	return Error{
		statusCode: status,
		message:    err.Error(),
	}
}

func (e Error) JSON(c *gin.Context) {
	c.JSON(e.statusCode, gin.H{
		"message": e.message,
	})
}
