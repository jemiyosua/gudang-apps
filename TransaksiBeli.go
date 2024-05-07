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

type JTransaksiBeliRequest struct {
	Username string
	ParamKey string
	Method string
	Id string
	KodeProduk string
	NamaProduk string
	HargaProduk string
	StatusProduk string
	TransaksiBelitList []JTransaksiBeliListRequest
	Page        int
	RowPage     int
	OrderBy     string
	Order       string
}

type JTransaksiBeliListRequest struct {
	IdProduk string
	NamaProduk string
	KategoriProduk string
	HargaBeliProduk string
	HargaJualProduk string
	Unit string
	Qty string
	Isi string
	Total string
	StatusProduk string
}

type JTransaksiBeliResponse struct {
	Id string
	KodeProduk string
	NamaProduk string
	HargaProduk string
	StatusProduk string
	TanggalInput string
}

func TransaksiBeli(c *gin.Context) {
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

	reqBody := JTransaksiBeliRequest{}
	jTransaksiBeliResponse := JTransaksiBeliResponse{}
	jTransaksiBeliResponses := []JTransaksiBeliResponse{}

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
	logFile = logFile + "TransaksiBeli_" + dateNow + ".log"
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
		dataLogTransaksiBeli(jTransaksiBeliResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}

	IsJson := helper.IsJson(bodyString)
	if !IsJson {
		errorMessage = "Error, Body - invalid json data"
		dataLogTransaksiBeli(jTransaksiBeliResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
		return
	}
	// ------ end of body json validation ------

	// ------ Header Validation ------
	if helper.ValidateHeader(bodyString, c) {
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			errorMessage = "Error, Bind Json Data"
			dataLogTransaksiBeli(jTransaksiBeliResponses, reqBody.Username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
			return
		} else {
			username := reqBody.Username
			paramKey := reqBody.ParamKey
			method := reqBody.Method
			transaksiBeliList := reqBody.TransaksiBelitList
			id := reqBody.Id
			kodeProduk := reqBody.KodeProduk
			namaProduk := reqBody.NamaProduk
			hargaProduk := reqBody.HargaProduk
			statusProduk := reqBody.StatusProduk
			page := reqBody.Page
			rowPage := reqBody.RowPage

			errorCodeRole, errorMessageRole, role := helper.GetRole(username, c)
			if errorCodeRole == "1" {
				dataLogTransaksiBeli(jTransaksiBeliResponses, reqBody.Username, errorCodeRole, errorMessageRole, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
				dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
			// ------ end of Param Validation ------

			// ------ start check session paramkey ------
			checkAccessVal := helper.CheckSession(username, paramKey, c)
			if checkAccessVal != "1" {
				dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCodeSession, errorMessageSession, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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
				
				transactionId := "TRX_" + day + month + year + "_" + hour + minute + second

				sliceLength := len(transaksiBeliList)

				query1 := fmt.Sprintf("INSERT INTO db_transaksi_beli (transaksi_id, jumlah_transaksi, user_input, tgl_input) VALUES ('%s', '%d', '%s', NOW())", transactionId, sliceLength, username)
				if _, err = db.Exec(query1); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				}

				var wg sync.WaitGroup
				wg.Add(sliceLength)

				for i := 0; i < sliceLength; i++ {
					go func(i int) {
						defer wg.Done()

						try.This(func() {

							idProdukList := transaksiBeliList[i].IdProduk
							namaProdukList := transaksiBeliList[i].NamaProduk
							catProdukList := transaksiBeliList[i].KategoriProduk
							hargaBeliProduklist := transaksiBeliList[i].HargaBeliProduk
							hargaJualProdukList := transaksiBeliList[i].HargaJualProduk
							unitProduklist := transaksiBeliList[i].Unit
							qtyProdukList := transaksiBeliList[i].Qty
							isiProdukList := transaksiBeliList[i].Isi
							totalProdukList := transaksiBeliList[i].Total
							statusProdukList := transaksiBeliList[i].StatusProduk

							// cek tabel master produk
							cntKodeProdukDB := 0
							query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_product WHERE produk_id = '%s'", idProdukList)
							if err := db.QueryRow(query).Scan(&cntKodeProdukDB); err != nil {
								errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
								dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								// return
							}

							if cntKodeProdukDB == 0 {
								query := fmt.Sprintf("INSERT INTO db_master_product (produk_id, nama_produk, cat_produk, status, user_input, tgl_update, tgl_input) VALUES ('%s', '%s', '%s', '%s', '%s', NOW(), NOW())", idProdukList, namaProdukList, catProdukList, statusProdukList, username)
								if _, err = db.Exec(query); err != nil {
									errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
									dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
									// return
								}
							}

							// cek tabel master produk harga
							hargaBeliDB := ""
							query2 := fmt.Sprintf("SELECT IFNULL(harga_beli,0) harga_beli FROM db_master_product_harga WHERE produk_id = '%s'", transaksiBeliList[i].IdProduk)
							if err := db.QueryRow(query2).Scan(&hargaBeliDB); err != nil {
								errorMessage = fmt.Sprintf("Error running %q: %+v", query2, err)
								dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
								// return
							}

							if hargaBeliDB != hargaBeliProduklist {
								hargaBeliInt, _ := strconv.Atoi(hargaBeliProduklist)
								hargaJualInt, _ := strconv.Atoi(hargaJualProdukList)
								margin := hargaJualInt - hargaBeliInt
								query := fmt.Sprintf("INSERT INTO db_master_product_harga (produk_id, harga_beli, harga_jual, margin, status, user_input, tgl_input) VALUES ('%s', '%s', '%s', '%d', '1', '%s', NOW())", idProdukList, hargaBeliProduklist, hargaJualProdukList, margin, username)
								if _, err = db.Exec(query); err != nil {
									errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
									dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
									// return
								}
							}

							query1 := fmt.Sprintf("INSERT INTO db_transaksi_beli_detail (transaksi_id, produk_id, harga_beli, unit, qty, isi, total, user_input, tgl_input) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', NOW())", transactionId, idProdukList, hargaBeliProduklist, unitProduklist, qtyProdukList, isiProdukList, totalProdukList, username)
							if _, err = db.Exec(query1); err != nil {
								errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
								dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				// Log := fmt.Sprintf("INSERT NEW ITEM : %s at %s : %s %s by %s", kodeProduk, hour, minute, state, username)
				// helper.LogActivity(username, "MASTER-ITEM", ip, bodyString, method, Log, errorCode, role, c)
				dataLogTransaksiBeli(jTransaksiBeliResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "UPDATE" {

				if kodeProduk == "" {
					errorMessage += "Kode Produk can't null value"
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				queryUpdate := ""

				if kodeProduk != "" {
					queryUpdate += fmt.Sprintf(" , kode_produk = '%s' ", kodeProduk)
				}

				if namaProduk != "" {
					queryUpdate += fmt.Sprintf(" , nama_produk = '%s' ", namaProduk)
				}

				if hargaProduk != "" {
					queryUpdate += fmt.Sprintf(" , harga_produk = '%s' ", hargaProduk)
				}

				if statusProduk != "" {
					queryUpdate += fmt.Sprintf(" , status = '%s' ", statusProduk)
				}

				query := fmt.Sprintf("UPDATE db_master_item SET tgl_update = NOW() %s WHERE kode_produk = '%s'", queryUpdate, kodeProduk)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("UPDATE ITEM : %s at %s : %s %s by %s", kodeProduk, hour, minute, state, username)
				helper.LogActivity(username, "MASTER-ITEM", ip, bodyString, method, Log, errorCode, role, c)
				dataLogTransaksiBeli(jTransaksiBeliResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "DELETE" {

				if kodeProduk == "" {
					errorMessage += "Kode Produk can't null value"
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				query := fmt.Sprintf("DELETE FROM db_master_item WHERE kode_produk = '%s'", kodeProduk)
				if _, err = db.Exec(query); err != nil {
					errorMessage = fmt.Sprintf("Error running %q: %+v", query, err)
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				Log := fmt.Sprintf("DELETE ITEM : %s at %s : %s %s by %s", kodeProduk, hour, minute, state, username)
				helper.LogActivity(username, "MASTER-ITEM", ip, bodyString, method, Log, errorCode, role, c)
				dataLogTransaksiBeli(jTransaksiBeliResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)

			} else if method == "SELECT" {

				// ------ start Param Validation ------
				if page == 0 {
					errorMessage += "Page can't null or 0 value"
				}
	
				if rowPage == 0 {
					errorMessage += "Page can't null or 0 value"
				}

				if errorMessage != "" {
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				if kodeProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}
					
					queryWhere += fmt.Sprintf(" kode_produk = '%s' ", kodeProduk)
				}

				if namaProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}
					
					queryWhere += fmt.Sprintf(" nama_produk LIKE '%%%s%%' ", namaProduk)
				}

				if statusProduk != "" {
					if queryWhere != "" {
						queryWhere += " AND "
					}
					
					queryWhere += fmt.Sprintf(" status = '%s' ", statusProduk)
				}

				if queryWhere != "" {
					queryWhere = " WHERE " + queryWhere
				}

				totalRecords = 0
				totalPage = 0
				query := fmt.Sprintf("SELECT COUNT(1) AS cnt FROM db_master_item %s", queryWhere)
				if err := db.QueryRow(query).Scan(&totalRecords); err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
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

				query1 := fmt.Sprintf(`SELECT id, kode_produk, nama_produk, harga_produk, tgl_input FROM db_master_item %s %s`, queryWhere, queryLimit)
				fmt.Println(query1)
				rows, err := db.Query(query1)
				defer rows.Close()
				if err != nil {
					errorMessage = "Error running, " + err.Error()
					dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
					return
				}

				for rows.Next() {
					err = rows.Scan(
						&jTransaksiBeliResponse.Id,
						&jTransaksiBeliResponse.KodeProduk,
						&jTransaksiBeliResponse.NamaProduk,
						&jTransaksiBeliResponse.HargaProduk,
						&jTransaksiBeliResponse.TanggalInput,
					)

					jTransaksiBeliResponses = append(jTransaksiBeliResponses, jTransaksiBeliResponse)

					if err != nil {
						errorMessage = fmt.Sprintf("Error running %q: %+v", query1, err)
						dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
						return
					}
				}

				fmt.Println("OK")

				dataLogTransaksiBeli(jTransaksiBeliResponses, username, "0", errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			} else {
				errorMessage = "Method undifined!"
				dataLogTransaksiBeli(jTransaksiBeliResponses, username, errorCode, errorMessage, totalRecords, totalPage, method, path, ip, logData, allHeader, bodyJson, c)
				return
			}
		}
	}
}

func dataLogTransaksiBeli(jTransaksiBeliResponses []JTransaksiBeliResponse, username string, errorCode string, errorMessage string, totalRecords float64, totalPage float64, method string, path string, ip string, logData string, allHeader string, bodyJson string, c *gin.Context) {
	if errorCode != "0" {
		helper.SendLogError(username, "MASTER ITEM", errorMessage, bodyJson, "", errorCode, allHeader, method, path, ip, c)
	}
	returnTransaksiBeli(jTransaksiBeliResponses, errorCode, errorMessage, logData, totalRecords, totalPage, c)
	return
}

func returnTransaksiBeli(jTransaksiBeliResponses []JTransaksiBeliResponse, errorCode string, errorMessage string, logData string, totalRecords float64, totalPage float64, c *gin.Context) {

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
			"Result": jTransaksiBeliResponses, 
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
