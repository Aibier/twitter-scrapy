package main

import (
	"twitter-scrapy/configs"
	"twitter-scrapy/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	//run database
	configs.ConnectDB()

	//routes
	routes.UserRoute(router)

	err := router.Run("localhost:8000")
	if err != nil {
		return 
	}
}
