package response

import "github.com/gin-gonic/gin"

type Success struct {
	StatusCode int
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
}

func NewSuccess(message string, status int) Success {
	return Success{
		StatusCode: status,
		Message:    message,
	}
}

func NewSuccessWithData(message string, status int, data interface{}) Success {
	return Success{
		StatusCode: status,
		Message:    message,
		Data:       data,
	}
}

func (s Success) JSON(c *gin.Context) {
	c.JSON(s.StatusCode, gin.H{
		"message": s.Message,
		"data":    s.Data,
	})
}
