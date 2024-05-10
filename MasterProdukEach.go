package main

import (
	"bytes"
	"fmt"
	"gudang/helper"
	"io/ioutil"
	"log"
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

type JMasterProdukEachRequest struct {
	Username             string
	ParamKey             string
	Method               string
	ProdukId             string
	NamaProduk           string
	KategoriIdProduk     string
	HargaBeliProduk      int
	HargaJualProduk      int
	UnitProduk           string
	QtyProduk            int
	IsiProduk            int
	UserInput            string
	TanggalExpiredProduk string
	Page                 int
	RowPage              int
	OrderBy              string
	Order                string
}

type JMasterProdukEachResponse struct {
	ProdukId             string
	NamaProduk           string
	KategoriIdProduk     string
	HargaBeliProduk      int
	HargaJualProduk      int
	UnitProduk           string
	QtyProduk            int
	IsiProduk            int
	UserInput            string
	TanggalExpiredProduk string
	TanggalInput         string
}

func MasterProdukEach(c *gin.Context) {
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

	reqBody := JMasterProdukEachRequest{}
	//jMasterProdukEachResponse := JMasterProdukEachResponse{}
	jMasterProdukEachResponses := []JMasterProdukEachResponse{}

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
	logFile = logFile + "MasterProdukEach_" + dateNow + ".log"
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
		dataLogMasterProdukEach(jMasterProdukEachResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogMasterProdukEach(jMasterProdukEachResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			errorMessage = "Error, Bind Json Data"
			dataLogMasterProdukEach(jMasterProdukEachResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
			return
		} else {
			username := reqBody.Username
			paramKey := reqBody.ParamKey
			method := reqBody.Method
			produkId := reqBody.ProdukId
			namaProduk := reqBody.NamaProduk
			kategoriIdProduk := reqBody.KategoriIdProduk
			hargaBeliProduk := reqBody.HargaBeliProduk
			hargaJualProduk := reqBody.HargaJualProduk
			// unitProduk := reqBody.UnitProduk
			qtyProduk := reqBody.QtyProduk
			isiProduk := reqBody.IsiProduk
			// tanggalExpiredProduk := reqBody.TanggalExpiredProduk
			userInput := reqBody.UserInput

			errorCodeRole, errorMessageRole, role := helper.GetRole(username, c)
			if errorCodeRole == "1" {
				dataLogMasterProdukEach(jMasterProdukEachResponses, reqBody.Username, errorCodeRole, errorMessageRole, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
				dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
			// ------ end of Param Validation ------

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(username, paramKey, c)
			if checkAccessVal != "1" {
				dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
				if produkId == "" {
					errorMessage += "Kode Produk can't null value"
				}

				if namaProduk == "" {
					errorMessage += "Method can't null value"
				}

				if kategoriIdProduk == "" {
					errorMessage += "Ketegori Id Produk can't null value"
				}

				if strconv.Itoa(hargaBeliProduk) == "" {
					errorMessage += "Harga Beli Produk can't null value"
				}

				if strconv.Itoa(hargaJualProduk) == "" {
					errorMessage += "Harga Jual Produk can't null value"
				}

				if strconv.Itoa(qtyProduk) == "" {
					errorMessage += "Quantity Produk can't null value"
				}

				if strconv.Itoa(isiProduk) == "" {
					errorMessage += "Isi Produk can't null value"
				}

				if errorMessage != "" {
					dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				cntKodeProdukDB := 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_product WHERE produk_id = '%s'", produkId)
				if err := db.QueryRow(query).Scan(&cntKodeProdukDB); err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				if cntKodeProdukDB == 0 {
					query := fmt.Sprintf("INSERT INTO db_master_product (produk_id, nama_produk, cat_produk, status, user_input, tgl_update, tgl_input) VALUES ('%s', '%s', '%s', '1', '%s', NOW(), NOW())", produkId, namaProduk, kategoriIdProduk, userInput)
					if _, err = db.Exec(query); err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					margin := hargaJualProduk - hargaBeliProduk
					query2 := fmt.Sprintf("INSERT INTO db_master_product_harga (produk_id, harga_beli, harga_jual, margin, status, user_input, tgl_input) VALUES ('%s', '%d', '%d', '%d', '1', '%s', NOW())", produkId, hargaBeliProduk, hargaJualProduk, margin, userInput)
					if _, err = db.Exec(query2); err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query2, err)
						dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					query3 := fmt.Sprintf("INSERT INTO db_master_product_stok (produk_id, qty, isi_produk, user_input, tgl_input) VALUES ('%s', '%d', '%d', '%s', NOW())", produkId, qtyProduk, isiProduk, userInput)
					if _, err = db.Exec(query3); err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query3, err)
						dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

				} else {
					errorMessage = "Product Exist!"
					dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("INSERT PRODUCT : %s at %s : %s %s by %s", produkId, hour, minute, state, username)
				helper.LogActivity(username, "MASTER-PRODUCT-EACH", ip, bodyString, method, Log, errorCode, role, c)
				dataLogMasterProdukEach(jMasterProdukEachResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "UPDATE" {

			} else if method == "DELETE" {

			} else if method == "SELECT" {

			} else {
				errorMessage = "Method undifined!"
				dataLogMasterProdukEach(jMasterProdukEachResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
		}
	}
}

func dataLogMasterProdukEach(jMasterProdukEachResponses []JMasterProdukEachResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "MASTER PRODUCT EACH", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnMasterProdukEach(jMasterProdukEachResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
	return
}

func returnMasterProdukEach(jMasterProdukEachResponses []JMasterProdukEachResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result":       jMasterProdukEachResponses,
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
