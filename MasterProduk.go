package main

import (
	"bytes"
	"database/sql"
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

type JMasterProdukRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	ProdukId string
	NamaProduk string
	KategoriIdProduk string
	Status string
	UserInput string
	MasterProdukList []JMasterProdukDetailRequest
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JMasterProdukDetailRequest struct {
	ProdukId string `json:"produk_id"`
	NamaProduk string `json:"nama_produk"`
	KategoriIdProduk string `json:"kategori_produk"`
	HargaBeliProduk int `json:"harga_beli"`
	HargaJualProduk int `json:"harga_jual"`
	UnitProduk string `json:"unit"`
	QtyProduk int `json:"qty"`
	IsiProduk int `json:"isi"`
	UserInputProduk string `json:"user_input"`
	TanggalExpiredProduk string `json:"tanggal_expired"`
}

type JMasterProdukResponse struct {
	Id string
	ProdukId string
	NamaProduk string
	KategoriIdProduk string
	KategoriNamaProduk string
	Status string
	UserInput string
	TanggalUpdate string
	TanggalInput string
	TotalStokProduk string
}

func MasterProduk(c *gin.Context) {
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

	reqBody := JMasterProdukRequest{}
	jMasterProdukResponse := JMasterProdukResponse{}
	jMasterProdukResponses := []JMasterProdukResponse{}

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
	logFile = logFile + "MasterProduk_" + dateNow + ".log"
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
		dataLogMasterProduk(jMasterProdukResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogMasterProduk(jMasterProdukResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			errorMessage = "Error, Bind Json Data"
			dataLogMasterProduk(jMasterProdukResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
			return
		} else {
			username := reqBody.Username
			paramKey := reqBody.ParamKey
			method := reqBody.Method
			idProduk := reqBody.Id
			produkId := reqBody.ProdukId
			namaProduk := reqBody.NamaProduk
			kategoriIdProduk := reqBody.KategoriIdProduk
			statusProduk := reqBody.Status
			userInput := reqBody.UserInput
			page := reqBody.Page
			rowPage := reqBody.RowPage

			errorCodeRole, errorMessageRole, role := helper.GetRole(username, c)
			if errorCodeRole == "1" {
				dataLogMasterProduk(jMasterProdukResponses, reqBody.Username, errorCodeRole, errorMessageRole, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
				dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
			// ------ end of Param Validation ------

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(username, paramKey, c)
			if checkAccessVal != "1" {
				dataLogMasterProduk(jMasterProdukResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				if len(reqBody.MasterProdukList) > 0 {
					sliceLength := len(reqBody.MasterProdukList)

					var wg sync.WaitGroup
					wg.Add(sliceLength)

					for i := 0; i < sliceLength; i++ {
						go func(i int) {
							defer wg.Done()

							try.This(func() {

								idProdukList := reqBody.MasterProdukList[i].ProdukId
								namaProdukList := reqBody.MasterProdukList[i].NamaProduk
								catProdukList := reqBody.MasterProdukList[i].KategoriIdProduk
								hargaBeliProdukList := reqBody.MasterProdukList[i].HargaBeliProduk
								hargaJualProdukList := reqBody.MasterProdukList[i].HargaJualProduk
								// unitProdukList := reqBody.MasterProdukList[i].UnitProduk
								// qtyProdukList := reqBody.MasterProdukList[i].QtyProduk
								// isiProdukList := reqBody.MasterProdukList[i].IsiProduk
								userInputProdukList := reqBody.MasterProdukList[i].UserInputProduk
								// tanggalExpiredProdukList := reqBody.MasterProdukList[i].TanggalExpiredProduk // 2026-12-03 (yyyy-mm-dd)
								// totalProdukList := 0

								cntKodeProdukDB := 0
								query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_product WHERE produk_id = '%s'", idProdukList)
								if err := db.QueryRow(query).Scan(&cntKodeProdukDB); err != nil {
									errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
									dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
									// return
								}

								if cntKodeProdukDB == 0 {
									query := fmt.Sprintf("INSERT INTO db_master_product (produk_id, nama_produk, cat_produk, status, user_input, tgl_update, tgl_input) VALUES ('%s', '%s', '%s', '1', '%s', NOW(), NOW())", idProdukList, namaProdukList, catProdukList, userInputProdukList)
									if _, err = db.Exec(query); err != nil {
										errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
										dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
										// return
									}
								}

								// cek tabel master produk harga
								var hargaBeliDBNS sql.NullString
								query2 := fmt.Sprintf("SELECT harga_beli FROM db_master_product_harga WHERE produk_id = '%s'", idProdukList)
								if err := db.QueryRow(query2).Scan(&hargaBeliDBNS); err != nil && err != sql.ErrNoRows {
									errorMessage = fmt.Sprintf("Error running %q: %+v", query2, err)
									dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
									// return
								}

								hargaBeliDB := hargaBeliDBNS.String
								hargaBeliDBInt, _ := strconv.Atoi(hargaBeliDB) 
								if hargaBeliDBInt != hargaBeliProdukList {
									margin := hargaJualProdukList - hargaBeliProdukList
									query := fmt.Sprintf("INSERT INTO db_master_product_harga (produk_id, harga_beli, harga_jual, margin, status, user_input, tgl_input) VALUES ('%s', '%d', '%d', '%d', '1', '%s', NOW())", idProdukList, hargaBeliProdukList, hargaJualProdukList, margin, userInputProdukList)
									if _, err = db.Exec(query); err != nil {
										errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
										dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
										// return
									}
								}

								// if unitProdukList == "pcs" {
								// 	totalProdukList = isiProdukList
								// } else if unitProdukList == "pack" {
								// 	totalProdukList = qtyProdukList * isiProdukList
								// }

								// cntTglExpired := 0
								// query1 := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_product_stok WHERE produk_id = '%s'", reqBody.MasterProdukList[i].ProdukId)
								// if err := db.QueryRow(query1).Scan(&cntTglExpired); err != nil {
								// 	errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
								// 	dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								// 	// return
								// }

								// if cntTglExpired > 0 {
								// 	tglExpiredProduct := ""
								// 	query1 := fmt.Sprintf("SELECT tgl_expired FROM db_master_product_stok WHERE produk_id = '%s'", reqBody.MasterProdukList[i].ProdukId)
								// 	if err := db.QueryRow(query1).Scan(&tglExpiredProduct); err != nil {
								// 		errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
								// 		dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								// 		// return
								// 	}

								// 	if (tanggalExpiredProdukList != tglExpiredProduct) {
								// 		query3 := fmt.Sprintf("INSERT INTO db_master_product_stok (produk_id, unit_produk, qty, isi_produk, total_produk, tgl_expired, user_input, tgl_input) VALUES ('%s', '%s', '%d', '%d', '%d', '%s', '%s', NOW())", idProdukList, unitProdukList, qtyProdukList, isiProdukList, totalProdukList, tanggalExpiredProdukList, userInputProdukList)
								// 		if _, err = db.Exec(query3); err != nil {
								// 			errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
								// 			dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								// 			// return
								// 		}
								// 	}
								// } else {
								// 	query3 := fmt.Sprintf("INSERT INTO db_master_product_stok (produk_id, unit_produk, qty, isi_produk, total_produk, tgl_expired, user_input, tgl_input) VALUES ('%s', '%s', '%d', '%d', '%d', '%s', '%s', NOW())", idProdukList, unitProdukList, qtyProdukList, isiProdukList, totalProdukList, tanggalExpiredProdukList, userInputProdukList)
								// 	if _, err = db.Exec(query3); err != nil {
								// 		errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
								// 		dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								// 		// return
								// 	}
								// }

							}).Finally(func() {
							}).Catch(func(e try.E) {
								// Print crash
							})
						}(i)
					}
					wg.Wait()

					runtime.GC()

					// Log := fmt.Sprintf("INSERT NEW ITEM : %s at %s : %s %s by %s", kodeProduk, hour, minute, state, username)
					// helper.LogActivity(username, "MASTER-ITEM", ip, bodyString, method, Log, errorCode, role, c)
					dataLogMasterProduk(jMasterProdukResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				} else {
					errorMessage := "Import data cannot null list"
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				}

			} else if method == "UPDATE" {

				if produkId == "" {
					errorMessage += "Product Id can't null value"
				}
	
				if errorMessage != "" {
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				queryUpdate := ""

				if namaProduk != "" {
					queryUpdate += fmt.Sprintf(" , nama_produk = '%s' ", namaProduk)
				}

				if kategoriIdProduk != "" {
					queryUpdate += fmt.Sprintf(" , cat_produk = '%s' ", kategoriIdProduk)
				}

				if statusProduk != "" {
					queryUpdate += fmt.Sprintf(" , status = '%s' ", statusProduk)
				}

				query := fmt.Sprintf("UPDATE db_master_product SET tgl_update = NOW() %s WHERE produk_id = '%s'", queryUpdate, produkId)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					// return
				}

				Log := fmt.Sprintf("UPDATE PRODUCT : %s at %s : %s %s by %s", produkId, hour, minute, state, username)
				helper.LogActivity(username, "MASTER-PRODUCT", ip, bodyString, method, Log, "0", role, c)
				dataLogMasterProduk(jMasterProdukResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "DELETE" {

				if produkId == "" {
					errorMessage += "Product Id can't null value"
				}
	
				if errorMessage != "" {
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				query := fmt.Sprintf("DELETE FROM db_master_product WHERE produk_id = '%s'", produkId)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					// return
				}

				query2 := fmt.Sprintf("DELETE FROM db_master_product_harga WHERE produk_id = '%s'", produkId)
				if _, err = db.Exec(query2); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					// return
				}

				query3 := fmt.Sprintf("DELETE FROM db_master_product_stok WHERE produk_id = '%s'", produkId)
				if _, err = db.Exec(query3); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					// return
				}

				Log := fmt.Sprintf("DELETE PRODUCT : %s at %s : %s %s by %s", produkId, hour, minute, state, username)
				helper.LogActivity(username, "MASTER-PRODUCT", ip, bodyString, method, Log, "0", role, c)
				dataLogMasterProduk(jMasterProdukResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "SELECT" {

				if page == 0 {
					errorMessage += "Page can't null or 0 value"
				}
	
				if rowPage == 0 {
					errorMessage += "Row Page can't null or 0 value"
				}

				if errorMessage != "" {
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				pageNow := (page - 1) * rowPage
				pageNowString := strconv.Itoa(pageNow)
				queryLimit := ""

				queryWhere := ""
				if idProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" dbmp.id = '%s' ", idProduk)
				}

				if produkId != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" produk_id = '%s' ", produkId)
				}

				if namaProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" nama_produk LIKE '%%%s%%' ", namaProduk)
				}

				if kategoriIdProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" dbmp.cat_produk = '%s' ", kategoriIdProduk)
				}

				if statusProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" dbmp.status = '%s' ", statusProduk)
				}

				if userInput != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}

					queryWhere += fmt.Sprintf(" dbmp.user_input LIKE '%%%s%%' ", userInput)
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_product dbmp LEFT JOIN db_category_product dbcp ON dbmp.cat_produk = dbcp.id %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				// ---------- start query get menu ----------
				query1 := fmt.Sprintf(`SELECT dbmp.id, produk_id, nama_produk, dbmp.cat_produk, IFNULL(dbcp.cat_name, '') cat_name, dbmp.status, dbmp.user_input, tgl_update, dbmp.tgl_input FROM db_master_product dbmp LEFT JOIN db_category_product dbcp ON dbmp.cat_produk = dbcp.id %s %s`, queryWhere, queryLimit)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jMasterProdukResponse.Id,
						&jMasterProdukResponse.ProdukId,
						&jMasterProdukResponse.NamaProduk,
						&jMasterProdukResponse.KategoriIdProduk,
						&jMasterProdukResponse.KategoriNamaProduk,
						&jMasterProdukResponse.Status,
						&jMasterProdukResponse.UserInput,
						&jMasterProdukResponse.TanggalUpdate,
						&jMasterProdukResponse.TanggalInput,
					)

					tanggalUpdateSplit := strings.Split(jMasterProdukResponse.TanggalUpdate, " ")
					tanggalUpdateDate := tanggalUpdateSplit[0]
					tanggalUpdateTime := tanggalUpdateSplit[1]
					
					tanggalUpdateDateSplit := strings.Split(tanggalUpdateDate, "-")
					dayupdate := tanggalUpdateDateSplit[2]
					monthUpdate := tanggalUpdateDateSplit[1]
					yearUpdate := tanggalUpdateDateSplit[0]
					yearUpdateSlice := yearUpdate[len(yearUpdate)-2:]

					tanggalUpdateTimeSplit := strings.Split(tanggalUpdateTime, ":")
					hourUpdate := tanggalUpdateTimeSplit[0]
					minuteUpdate := tanggalUpdateTimeSplit[1]

					jMasterProdukResponse.TanggalUpdate = dayupdate + "." + monthUpdate + "." + yearUpdateSlice + " - " + hourUpdate + "." + minuteUpdate

					tanggalInputSplit := strings.Split(jMasterProdukResponse.TanggalInput, " ")
					tanggalInputDate := tanggalInputSplit[0]
					tanggalInputTime := tanggalInputSplit[1]
					
					tanggalInputDateSplit := strings.Split(tanggalInputDate, "-")
					dayInput := tanggalInputDateSplit[2]
					monthInput := tanggalInputDateSplit[1]
					yearInput := tanggalInputDateSplit[0]
					yearInputSlice := yearInput[len(yearInput)-2:]

					tanggalInputTimeSplit := strings.Split(tanggalInputTime, ":")
					hourinput := tanggalInputTimeSplit[0]
					minuteInput := tanggalInputTimeSplit[1]

					jMasterProdukResponse.TanggalInput = dayInput + "." + monthInput + "." + yearInputSlice + " - " + hourinput + "." + minuteInput

					query := fmt.Sprintf("SELECT IFNULL(SUM(total_produk), 0) total_produk FROM db_master_product_stok WHERE produk_id = '%s'", jMasterProdukResponse.ProdukId)
					if err := db.QueryRow(query).Scan(&jMasterProdukResponse.TotalStokProduk); err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
						dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}

					jMasterProdukResponses = append(jMasterProdukResponses, jMasterProdukResponse)

					if err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
						dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
				}
				// ---------- end of query get menu ----------

				dataLogMasterProduk(jMasterProdukResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				errorMessage = "Method undifined!"
				dataLogMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
		}
	}
}

func dataLogMasterProduk(jMasterProdukResponses []JMasterProdukResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "MENU", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnMasterProduk(jMasterProdukResponses, username, errorCode, errorMessage, logData, totalRecords, totalPage, c)
}

func returnMasterProduk(jMasterProdukResponses []JMasterProdukResponse, username string, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"UsernameLogin":   username,
			"Result": jMasterProdukResponses, 
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
}
