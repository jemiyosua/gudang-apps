package helper

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func ValidateHeader(requestBody string, c *gin.Context) bool {

	var result gin.H

	pageGo := "VALIDATE-HEADER"

	signature := c.GetHeader("Signature")
	contentType := c.GetHeader("Content-Type")

	enckey := os.Getenv("ENCKEY")
	base64String := base64.StdEncoding.EncodeToString([]byte(requestBody))
	signatureKey := fmt.Sprintf("%x", md5.Sum([]byte(enckey+base64String)))

	if contentType == "" {
		errorMessage := "Error, Header - Content-Type is not application/json or empty value "
		SendLogError("", pageGo, errorMessage, "", "", "1", "", "", "", "", c)
		result = gin.H{
			"ErrCode":    "1",
			"ErrMessage": errorMessage,
			"Result":     "",
		}
		c.JSON(http.StatusOK, result)
		return false
	}

	if signature == "" {
		errorMessage := "Header Signature can not null"
		SendLogError("", pageGo, errorMessage, "", "", "1", "", "", "", "", c)
		result = gin.H{
			"ErrCode":    "1",
			"ErrMessage": errorMessage,
			"Result":     "",
		}
		c.JSON(http.StatusOK, result)
		return false
	} else {
		fmt.Println("signature : " + signature)
		fmt.Println("SignatureKey : " + signatureKey)

		if signature == signatureKey {
			return true
		} else {
			errorMessage := "Header Signature invalid "
			SendLogError("", pageGo, errorMessage, "", "", "1", "", "", "", "", c)
			result = gin.H{
				"ErrCode":    "1",
				"ErrMessage": errorMessage,
				"Result":     "",
			}
			c.JSON(http.StatusOK, result)
			return false
		}
	}
}
