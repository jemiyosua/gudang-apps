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
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/manucorporat/try"
)

type JTransaksiJualRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	TransaksiId string
	IdProduk string
	HargaProduk string
	StatusProduk string
	TransaksiJualtList []JTransaksiJualListRequest
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JTransaksiJualListRequest struct {
	IdProduk string
	KategoriProduk string
	HargaJualProduk float64
	Qty int
	Total int
}

type JTransaksiJualResponse struct {
	Id string
	TransaksiId string
	JumlahTransaksi string
	TotalTransaksi string
	UserInput string
	TanggalInput string
}

func TransaksiJual(c *gin.Context) {
	db := helper.Connect(c)
	defer db.Close()
	startTime := time.Now()
	startTimeString := startTime.String()

	var (
		bodyBytes []byte
		xRealIp   string
		ip        string
		logFile   string
		totalRecords float64
		totalPage float64
	)

	reqBody := JTransaksiJualRequest{}
	jTransaksiJualResponse := JTransaksiJualResponse{}
	jTransaksijualResponses := []JTransaksiJualResponse{}

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
	logFile = logFile + "TransaksiJual_" + dateNow + ".log"
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
		dataLogTransaksiJual(jTransaksijualResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogTransaksiJual(jTransaksijualResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			errorMessage = "Error, Bind Json Data"
			dataLogTransaksiJual(jTransaksijualResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
			return
		} else {
			username := reqBody.Username
			paramKey := reqBody.ParamKey
			method := reqBody.Method
			transaksiJualList := reqBody.TransaksiJualtList
			id := reqBody.Id
			idProduk := reqBody.IdProduk
			transaksiId := reqBody.TransaksiId
			page := reqBody.Page
			rowPage := reqBody.RowPage

			errorCodeRole, errorMessageRole, _ := helper.GetRole(username, c)
			if errorCodeRole == "1" {
				dataLogTransaksiJual(jTransaksijualResponses, reqBody.Username, errorCodeRole, errorMessageRole, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
				dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
			// ------ end of Param Validation ------

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(username, paramKey, c)
			if checkAccessVal != "1" {
				dataLogTransaksiJual(jTransaksijualResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}

			// currentTime := time.Now()
			// timeNow := currentTime.Format("15:04:05")
			// timeNowSplit := strings.Split(timeNow, ":")
			// hour := timeNowSplit[0]
			// minute := timeNowSplit[1]
			// state := ""
			// if hour < "12" {
			// 	state = "AM"
			// } else {
			// 	state = "PM"
			// }

			if method == "INSERT" {

				currentTime := time.Now()
				timeNow := currentTime.Format("01/02/2006 15:04:05")
				timeNowSplit := strings.Split(timeNow, " ")
				date := timeNowSplit[0]
				time := timeNowSplit[1]
				dateSplit := strings.Split(date, "/")
				day := dateSplit[0]
				month := dateSplit[1]
				year := dateSplit[2]
				timeSplit := strings.Split(time, ":")
				hour := timeSplit[0]
				minute := timeSplit[1]
				second := timeSplit[2]
				
				transactionId := "TRX_OUT_" + day + month + year + "_" + hour + minute + second

				sliceLength := len(transaksiJualList)

				query1 := fmt.Sprintf("INSERT INTO db_transaksi_jual (transaksi_id, jumlah_transaksi, user_input, tgl_input) VALUES ('%s', '%d', '%s', NOW())", transactionId, sliceLength, username)
				if _, err = db.Exec(query1); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
					dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				}

				var wg sync.WaitGroup
				wg.Add(sliceLength)

				for i := 0; i < sliceLength; i++ {
					go func(i int) {
						defer wg.Done()

						try.This(func() {

							idProdukList := transaksiJualList[i].IdProduk
							hargaJualProdukList := transaksiJualList[i].HargaJualProduk
							qtyProdukList := transaksiJualList[i].Qty
							totalProdukList := transaksiJualList[i].Total

							query := fmt.Sprintf("INSERT INTO db_transaksi_beli_detail (transaksi_id, produk_id, harga_jual, qty, total, user_input, tgl_input) VALUES ('%s', '%s', '%f', '%d', '%d', '%s', NOW())", transactionId, idProdukList, hargaJualProdukList, qtyProdukList, totalProdukList, username)
							if _, err = db.Exec(query); err != nil {
								errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
								dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								// return
							}

						}).Finally(func() {
						}).Catch(func(e try.E) {
							// Print crash
						})
					}(i)
				}
				wg.Wait()

				runtime.GC()

				dataLogTransaksiJual(jTransaksijualResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "UPDATE" {

			} else if method == "DELETE" {

			} else if method == "SELECT" {

				// ------ start Param Validation ------
				if page == 0 {
					errorMessage += "Page can't null or 0 value"
				}
	
				if rowPage == 0 {
					errorMessage += "Page can't null or 0 value"
				}

				if errorMessage != "" {
					dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}
				// ------ end of Param Validation ------

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

				if idProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}
					
					queryWhere += fmt.Sprintf(" produk_id = '%s' ", idProduk)
				}

				if transaksiId != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}
					
					queryWhere += fmt.Sprintf(" transaksi_id LIKE '%%%s%%' ", transaksiId)
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_transaksi_jual %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				query1 := fmt.Sprintf(`SELECT id, transaksi_id, jumlah_transaksi, user_input, tgl_input FROM db_transaksi_jual %s %s`, queryWhere, queryLimit)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jTransaksiJualResponse.Id,
						&jTransaksiJualResponse.TransaksiId,
						&jTransaksiJualResponse.JumlahTransaksi,
						&jTransaksiJualResponse.UserInput,
						&jTransaksiJualResponse.TanggalInput,
					)

					query := fmt.Sprintf("SELECT sum(total) as total_transaksi FROM db_transaksi_jual_detail WHERE transaksi_id = '%s'", jTransaksiJualResponse.TransaksiId)
					if err := db.QueryRow(query).Scan(&jTransaksiJualResponse.TotalTransaksi); err != nil {
						errorMessage = "Error running, " + err.Error()
						dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					jTransaksijualResponses = append(jTransaksijualResponses, jTransaksiJualResponse)

					if err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
						dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
				}

				dataLogTransaksiJual(jTransaksijualResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				errorMessage = "Method undifined!"
				dataLogTransaksiJual(jTransaksijualResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
		}
	}
}

func dataLogTransaksiJual(jTransaksijualResponses []JTransaksiJualResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "MASTER ITEM", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnTransaksiJual(jTransaksijualResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
	return
}

func returnTransaksiJual(jTransaksijualResponses []JTransaksiJualResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"DateTime":   currentTime1,
			"TotalRecords":   totalRecords,
			"TotalPage":   totalPage,
			"Result": jTransaksijualResponses, 
		})
	}

	startTime := time.Now()

	rex := regexp.MustCompile(`\r?\n`)
	endTime := time.Now()
	codeError := "200"

	diff := endTime.Sub(startTime)

	logDataNew := rex.ReplaceAllString(logData + codeError + "~" + endTime.String() + "~" + diff.String() + "~" + errorMessage, "")
	log.Println(logDataNew)

	runtime.GC()

	return
}
