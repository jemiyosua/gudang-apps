package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func CheckSession(username string, paramKey string, c *gin.Context) string {

	db := Connect(c)
	defer db.Close()

	ParamKeyDB := ""
	query := fmt.Sprintf("SELECT paramkey FROM db_login_session WHERE username = '%s' AND paramkey = '%s' AND tgl_input >= NOW() LIMIT 1", username, paramKey)
	if err := db.QueryRow(query).Scan(&ParamKeyDB); err != nil {
		errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
		return errorMessage
	}

	if ParamKeyDB != "" {
		query1 := fmt.Sprintf("UPDATE db_login_session SET tgl_input = (ADDTIME(NOW(), '0:20:0')) WHERE username = '%s' AND paramkey = '%s' AND tgl_input >= NOW() LIMIT 1", username, paramKey)
		_, err1 := db.Exec(query1)
		if err1 != nil {
			errorMessage := fmt.Sprintf("Error running %q: %+v", query1, err1)
			return errorMessage
		}
		return "1"
	}
	return "2"
}
