package main

import (
	"github.com/Aibier/twitter-scrapy/configs"
	"github.com/Aibier/twitter-scrapy/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	//run database
	configs.ConnectDB()
	//routes
	routes.TwitRoute(router)

	err := router.Run("localhost:8000")
	if err != nil {
		return 
	}
}
