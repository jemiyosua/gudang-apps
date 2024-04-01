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

type JProductStockRequest struct {
	Username       string
	ParamKey       string
	Method         string
	Id             string
	KodeProduk     string
	UnitProduk     string
	Quantity       int
	IsiProduk      int
	TotalProduk    int
	TanggalExpired string
	Page           int
	RowPage        int
	OrderBy        string
	Order          string
}

type JProductStockResponse struct {
	Id             string
	KodeProduk     string
	UnitProduk     string
	Quantity       int
	IsiProduk      int
	TotalProduk    int
	TanggalExpired string
	TanggalInput   string
}

func ProductStock(c *gin.Context) {
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

	reqBody := JProductStockRequest{}
	jProductStockResponse := JProductStockResponse{}
	jProductStockResponses := []JProductStockResponse{}

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
	logFile = logFile + "ProductStock_" + dateNow + ".log"
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
		dataLogProductStock(jProductStockResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogProductStock(jProductStockResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			errorMessage = "Error, Bind Json Data"
			dataLogProductStock(jProductStockResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
			return
		} else {
			username := reqBody.Username
			paramKey := reqBody.ParamKey
			method := reqBody.Method
			id := reqBody.Id
			kodeProduk := reqBody.KodeProduk
			unitProduk := reqBody.UnitProduk
			quantity := reqBody.Quantity
			isiProduk := reqBody.IsiProduk
			totalProduk := reqBody.TotalProduk
			tanggalExpired := reqBody.TanggalExpired
			page := reqBody.Page
			rowPage := reqBody.RowPage

			errorCodeRole, errorMessageRole, role := helper.GetRole(username, c)
			if errorCodeRole == "1" {
				dataLogProductStock(jProductStockResponses, reqBody.Username, errorCodeRole, errorMessageRole, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

			if page == 0 {
				errorMessage += "Page can't null or 0 value"
			}

			if rowPage == 0 {
				errorMessage += "Page can't null or 0 value"
			}

			if errorMessage != "" {
				dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
			// ------ end of Param Validation ------

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(username, paramKey, c)
			if checkAccessVal != "1" {
				dataLogProductStock(jProductStockResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
				if kodeProduk == "" {
					errorMessage += "Kode Produk can't null value"
				} else {
					cntKodeProdukDB := 0
					query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_product WHERE produk_id = '%s'", kodeProduk)
					if err := db.QueryRow(query).Scan(&cntKodeProdukDB); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					if cntKodeProdukDB < 1 {
						errorMessage += fmt.Sprintf("Kode Produk %s does not exists!", kodeProduk)
					}
				}

				if errorMessage != "" {
					dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

				query := fmt.Sprintf("INSERT INTO db_master_product_stok(produk_id, unit_produk, qty, isi_produk, total_produk, tgl_expired, user_input, tgl_input)  VALUES ('%s', '%s', '%d', '%d', '%d', '%s', '%s', sysdate() + interval 7 hour )", kodeProduk, unitProduk, quantity, isiProduk, totalProduk, tanggalExpired, username)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("INSERT NEW STOCK : %s at %s : %s %s by %s", kodeProduk, hour, minute, state, username)
				helper.LogActivity(username, "PRODUCT STOCK", ip, bodyString, method, Log, errorCode, role, c)
				dataLogProductStock(jProductStockResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "UPDATE" {

			} else if method == "DELETE" {

				if id == "" {
					errorMessage += "ID Stock can't null value"
					dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				query := fmt.Sprintf("DELETE FROM db_master_product_stok WHERE id = '%s'", id)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("DELETE PRODUCT STOCK : %s at %s : %s %s by %s", kodeProduk, hour, minute, state, username)
				helper.LogActivity(username, "PRODUCT STOCK", ip, bodyString, method, Log, errorCode, role, c)
				dataLogProductStock(jProductStockResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "SELECT" {
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

				if kodeProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" produk_id = '%s' ", kodeProduk)
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_product_stok %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				query1 := fmt.Sprintf(`SELECT id, produk_id, unit_produk, qty, isi_produk, total_produk, to_char(tgl_expired,'YYYY/MM/DD'), tgl_input  FROM db_master_product_stok %s %s`, queryWhere, queryLimit)
				fmt.Println(query1)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jProductStockResponse.Id,
						&jProductStockResponse.KodeProduk,
						&jProductStockResponse.UnitProduk,
						&jProductStockResponse.Quantity,
						&jProductStockResponse.IsiProduk,
						&jProductStockResponse.TotalProduk,
						&jProductStockResponse.TanggalExpired,
						&jProductStockResponse.TanggalInput,
					)

					jProductStockResponses = append(jProductStockResponses, jProductStockResponse)

					if err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
						dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
				}

				fmt.Println("OK")

				dataLogProductStock(jProductStockResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				errorMessage = "Method undifined!"
				dataLogProductStock(jProductStockResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
		}
	}
}

func dataLogProductStock(jProductStockResponses []JProductStockResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "PRODUCT STOCK", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnProductStock(jProductStockResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
	return
}

func returnProductStock(jProductStockResponses []JProductStockResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result":       jProductStockResponses,
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
