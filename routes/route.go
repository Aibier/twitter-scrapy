package routes

import (
	"github.com/Aibier/twitter-scrapy/controllers"

	"github.com/gin-gonic/gin"
)

// TwitRoute ...
func TwitRoute(router *gin.Engine) {
	router.GET("/sync", controllers.SyncPosts())
	router.GET("/twits/:twitId", controllers.GetTwitPost())
	router.DELETE("/twits/:twitId", controllers.DeleteTwitPost())
	router.GET("/twits", controllers.GetAllTwits())
}
