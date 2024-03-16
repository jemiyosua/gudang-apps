package helper

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func GetRole(username string, c *gin.Context) (string, string, string) {
	db := Connect(c)
	defer db.Close()

	roleDB := ""
	query := fmt.Sprintf("SELECT role FROM db_login WHERE username = '%s'", username)
	if err := db.QueryRow(query).Scan(&roleDB); err != nil {
		errorMessage := fmt.Sprintf("Error running %q: %+v", query, err)
		return "1", errorMessage, ""
	}

	return "0", "", roleDB
}
