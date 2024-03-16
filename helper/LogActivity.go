package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func LogActivity(username string, page string, ip string, jsonRequest string, method string, log string, logStatus string, role string, c *gin.Context) {
	db := Connect(c)
	defer db.Close()

	query := fmt.Sprintf("INSERT into db_log_activity (username, page, ip, json_request, method, log, log_status, role, tgl_input) values ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', NOW())", TrimReplace(username), TrimReplace(page), TrimReplace(ip), TrimReplace(jsonRequest), TrimReplace(method), TrimReplace(log), TrimReplace(logStatus), TrimReplace(role))
	fmt.Println(query)
	_, err := db.Exec(query)
	if err != nil {
		return
	}
}