package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func main() {
	router := gin.Default()
	router.Use(CORS())

	apiVersion := "/api/v1/"

	router.POST(apiVersion+"Login", Login)
	router.POST(apiVersion+"Menu", Menu)
	router.POST(apiVersion+"MasterProduk", MasterProduk)
	router.POST(apiVersion+"TransaksiBeli", TransaksiBeli)
	router.POST(apiVersion+"TransaksiJual", TransaksiJual)
	router.POST(apiVersion+"TransaksiJualDetail", TransaksiJualDetail)
	router.POST(apiVersion+"Category", Category)
	router.POST(apiVersion+"ProductPrice", ProductPrice)
	router.POST(apiVersion+"ProductStock", ProductStock)
	router.POST(apiVersion+"ScanProduct", ScanProduct)
	router.POST(apiVersion+"MasterProdukEach", MasterProdukEach)

	PORT := os.Getenv("PORT")

	router.Run(":" + PORT)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Signature, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
