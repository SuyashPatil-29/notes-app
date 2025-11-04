package utils

import (
	"backend/internal/types"

	"github.com/gin-gonic/gin"
)

// SendErrorResponse sends a structured error response
func SendErrorResponse(c *gin.Context, err *types.APIError) {
	response := gin.H{
		"error": gin.H{
			"code":    err.Code,
			"message": err.Message,
		},
	}

	if err.Details != "" {
		response["error"].(gin.H)["details"] = err.Details
	}

	if err.Suggestion != "" {
		response["error"].(gin.H)["suggestion"] = err.Suggestion
	}

	c.JSON(err.HTTPStatus, response)
}

// SendSuccessResponse sends a structured success response
func SendSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    data,
	})
}

// SendSuccessMessageResponse sends a success response with a message
func SendSuccessMessageResponse(c *gin.Context, message string, data interface{}) {
	response := gin.H{
		"success": true,
		"message": message,
	}

	if data != nil {
		response["data"] = data
	}

	c.JSON(200, response)
}
