package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func SendLogError(username string, page string, errorLog string, jsonRequest string, jsonResponse string, errorCode string, headerAuth string, methodRequest string, path string, ip string, c *gin.Context) {
	db := Connect(c)
	defer db.Close()

	query := fmt.Sprintf("INSERT INTO db_log_error (username, page, error_log, json_request, json_response, error_code, header, method, path, ip, tgl_input) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', NOW());", TrimReplace(username), TrimReplace(page), TrimReplace(errorLog), TrimReplace(jsonRequest), TrimReplace(jsonResponse), TrimReplace(errorCode), TrimReplace(headerAuth), TrimReplace(methodRequest), TrimReplace(path), TrimReplace(ip))
	_, err := db.Exec(query)
	if err != nil {
		return
	}
}