package main

import (
	"bytes"
	"fmt"
	"gudang/helper"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type JCategoryRequest struct {
	Username       string
	ParamKey       string
	Method         string
	Id             string
	NamaKategori   string
	StatusKategori string
	Page           int
	RowPage        int
	OrderBy        string
	Order          string
}

type JCategoryResponse struct {
	Id             string
	NamaKategori   string
	StatusKategori string
	TanggalInput   string
}

func Category(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	startTime := time.Now()
	startTimeString := startTime.String()

	var (
		bodyBytes    []byte
		xRealIp      string
		ip           string
		logFile      string
		totalRecords float64
		totalPage    float64
	)

	reqBody := JCategoryRequest{}
	jCategoryResponse := JCategoryResponse{}
	jCategoryResponses := []JCategoryResponse{}

	errorCode := "1"
	errorMessage := ""
	errorCodeSession := "2"
	errorMessageSession := "Session Expired"

	allHeader := helper.ReadAllHeader(c)
	logFile = os.Getenv("LOGFILE")
	method := c.Request.Method
	path := c.Request.URL.EscapedPath()

	// ---------- start get ip ----------
	if Values, _ := c.Request.Header["X-Real-Ip"]; len(Values) > 0 {
		xRealIp = Values[0]
	}

	if xRealIp != "" {
		ip = xRealIp
	} else {
		ip = c.ClientIP()
	}
	// ---------- end of get ip ----------

	// ---------- start log file ----------
	dateNow := startTime.Format("2006-01-02")
	logFile = logFile + "Category_" + dateNow + ".log"
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	// ---------- end of log file ----------

	// ------ start body json validation ------
	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	bodyString := string(bodyBytes)

	bodyJson := helper.TrimReplace(string(bodyString))
	logData := startTimeString + "~" + ip + "~" + method + "~" + path + "~" + allHeader + "~"
	rex := regexp.MustCompile(`\r?\n`)
	logData = logData + rex.ReplaceAllString(bodyJson, "") + "~"

	if string(bodyString) == "" {
		errorMessage = "Error, Body is empty"
		dataLogCategory(jCategoryResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogCategory(jCategoryResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			errorMessage = "Error, Bind Json Data"
			dataLogCategory(jCategoryResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
			return
		} else {
			username := reqBody.Username
			paramKey := reqBody.ParamKey
			method := reqBody.Method
			id := reqBody.Id
			namaKategori := strings.TrimSpace(reqBody.NamaKategori)
			statusKategori := reqBody.StatusKategori
			page := reqBody.Page
			rowPage := reqBody.RowPage

			errorCodeRole, errorMessageRole, role := helper.GetRole(username, c)
			if errorCodeRole == "1" {
				dataLogCategory(jCategoryResponses, reqBody.Username, errorCodeRole, errorMessageRole, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}

			// ------ Param Validation ------
			if username == "" {
				errorMessage += "Username can't null value"
			}

			if paramKey == "" {
				errorMessage += "ParamKey can't null value"
			}

			if method == "" {
				errorMessage += "Method can't null value"
			}

			if errorMessage != "" {
				dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
			// ------ end of Param Validation ------

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(username, paramKey, c)
			if checkAccessVal != "1" {
				dataLogCategory(jCategoryResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}

			currentTime := time.Now()
			timeNow := currentTime.Format("15:04:05")
			timeNowSplit := strings.Split(timeNow, ":")
			hour := timeNowSplit[0]
			minute := timeNowSplit[1]
			state := ""
			if hour < "12" {
				state = "AM"
			} else {
				state = "PM"
			}

			if method == "INSERT" {

				// ------ Param Validation ------
				if namaKategori == "" {
					errorMessage += "Nama Kategori can't null value"
				} else {
					cntNamaKategoriDB := 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_category_product WHERE upper(cat_name) = upper('%s')", namaKategori)
					if err := db.QueryRow(query).Scan(&cntNamaKategoriDB); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					if cntNamaKategoriDB > 0 {
						errorMessage += fmt.Sprintf("Nama Kategori %s already exist!", namaKategori)
					}
				}

				if errorMessage != "" {
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				query := fmt.Sprintf("INSERT INTO db_category_product(cat_name, status, user_input, tgl_input) VALUES ('%s', 1, '%s', sysdate() + interval 7 hour)", namaKategori, username)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("INSERT NEW CATEGORY : %s at %s : %s %s by %s", namaKategori, hour, minute, state, username)
				helper.LogActivity(username, "CATEGORY", ip, bodyString, method, Log, errorCode, role, c)
				dataLogCategory(jCategoryResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "UPDATE" {

				if id == "" {
					errorMessage += "ID Category can't null value"
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				queryUpdate := ""

				if namaKategori != "" {
					cntNamaKategoriDB := 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_category_product WHERE upper(cat_name) = upper('%s')", namaKategori)
					if err := db.QueryRow(query).Scan(&cntNamaKategoriDB); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					if cntNamaKategoriDB > 0 {
						errorMessage += fmt.Sprintf("Nama Kategori %s already exist!", namaKategori)
					} else {
						if queryUpdate != "" {
							queryUpdate += fmt.Sprintf(" , cat_name = '%s' ", namaKategori)
						} else {
							queryUpdate += fmt.Sprintf(" cat_name = '%s' ", namaKategori)
						}

					}
				}

				if errorMessage != "" {
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if statusKategori != "" {
					if queryUpdate != "" {
						queryUpdate += fmt.Sprintf(" , status = '%s' ", statusKategori)
					} else {
						queryUpdate += fmt.Sprintf(" status = '%s' ", statusKategori)
					}
				}

				query := fmt.Sprintf("UPDATE db_category_product SET %s WHERE id = '%s'", queryUpdate, id)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("UPDATE CATEGORY : %s at %s : %s %s by %s", namaKategori, hour, minute, state, username)
				helper.LogActivity(username, "CATEGORY", ip, bodyString, method, Log, errorCode, role, c)
				dataLogCategory(jCategoryResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "DELETE" {

				if id == "" {
					errorMessage += "ID Category can't null value"
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				query := fmt.Sprintf("DELETE FROM db_category_product WHERE id = '%s'", id)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("DELETE CATEGORY : %s at %s : %s %s by %s", id, hour, minute, state, username)
				helper.LogActivity(username, "CATEGORY", ip, bodyString, method, Log, errorCode, role, c)
				dataLogCategory(jCategoryResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "SELECT" {
				if page == 0 {
					errorMessage += "Page can't null or 0 value"
				}
	
				if rowPage == 0 {
					errorMessage += "Page can't null or 0 value"
				}

				if errorMessage != "" {
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				pageNow := (page - 1) * rowPage
				pageNowString := strconv.Itoa(pageNow)
				queryLimit := ""

				queryWhere := ""
				if id != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" id = '%s' ", id)
				}

				if namaKategori != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" cat_name LIKE '%%%s%%' ", namaKategori)
				}

				if statusKategori != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" status = '%s' ", statusKategori)
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_category_product %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if rowPage == -1 {
					queryLimit = ""
					totalPage = 1
				} else {
					rowPageString := strconv.Itoa(rowPage)
					queryLimit = "LIMIT " + pageNowString + "," + rowPageString
					totalPage = math.Ceil(float64(totalRecords) / float64(rowPage))
				}

				query1 := fmt.Sprintf(`SELECT id, cat_name, status, tgl_input FROM db_category_product %s %s`, queryWhere, queryLimit)
				fmt.Println(query1)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jCategoryResponse.Id,
						&jCategoryResponse.NamaKategori,
						&jCategoryResponse.StatusKategori,
						&jCategoryResponse.TanggalInput,
					)

					jCategoryResponses = append(jCategoryResponses, jCategoryResponse)

					if err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
						dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
				}

				fmt.Println("OK")

				dataLogCategory(jCategoryResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				errorMessage = "Method undifined!"
				dataLogCategory(jCategoryResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
		}
	}
}

func dataLogCategory(jCategoryResponses []JCategoryResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "CATEGORY", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnCategory(jCategoryResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
	return
}

func returnCategory(jCategoryResponses []JCategoryResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

	if strings.Contains(errorMessage, "Error running") {
		errorMessage = "Error Execute data"
	}

	if errorCode == "504" {
		c.String(http.StatusUnauthorized, "")
	} else {
		currentTime := time.Now()
		currentTime1 := currentTime.Format("01/02/2006 15:04:05")

		c.PureJSON(http.StatusOK, gin.H{
			"ErrorCode":    errorCode,
			"ErrorMessage": errorMessage,
			"DateTime":     currentTime1,
			"TotalRecords": totalRecords,
			"TotalPage":    totalPage,
			"Result":       jCategoryResponses,
		})
	}

	startTime := time.Now()

	rex := regexp.MustCompile(`\r?\n`)
	endTime := time.Now()
	codeError := "200"

	diff := endTime.Sub(startTime)

	logDataNew := rex.ReplaceAllString(logData+codeError+"~"+endTime.String()+"~"+diff.String()+"~"+errorMessage, "")
	log.Println(logDataNew)

	runtime.GC()

	return
}
