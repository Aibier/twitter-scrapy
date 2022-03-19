package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	helpers "twitter-scrapy/helper"

	"twitter-scrapy/configs"
	"twitter-scrapy/models"
	"twitter-scrapy/responses"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var twitPostCollection *mongo.Collection = configs.GetCollection(configs.DB, "posts")
var validate = validator.New()

func CreateTwitPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var twit models.TwitPost
		defer cancel()

		//validate the request body
		if err := c.BindJSON(&twit); err != nil {
			c.JSON(http.StatusBadRequest, responses.TwitPostResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//use the validator library to validate required fields
		if validationErr := validate.Struct(&twit); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.TwitPostResponse{Status: http.StatusBadRequest, Message: "error", Data: map[string]interface{}{"data": validationErr.Error()}})
			return
		}

		newTwit := models.TwitPost{
			Id:       primitive.NewObjectID(),
			CreatedAt: time.Now(),
			Title:    twit.Title,
		}

		result, err := twitPostCollection.InsertOne(ctx, newTwit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.TwitPostResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusCreated, responses.TwitPostResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}
const PerPAGE = "100"
func SyncPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		cxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var posts []models.TwitPost
		config := helpers.HttPConfig{
			RequestTimeout: 30,
			SSLEnabled:     false,
			Username:       os.Getenv("USERNAME"),
			Password:       os.Getenv("PASSWORD"),
		}
		twitterClient, err := helpers.NewHTTPClient(config)
		if err != nil {
			log.Printf("err %v", err)
		}
		baseURL, _ := url.Parse("https://api.twitter.com/2/tweets/search/recent?query=nyc")
		responseData, responseStatusCode, err := twitterClient.MakeRequest(cxt, "GET", baseURL, nil)
		defer cancel()
		if responseStatusCode != 200 {
			var errResponse responses.ErrorResponse
			_ = json.Unmarshal(responseData, &errResponse)
			log.Printf("reponse %v", string(responseData))
			_ = json.Unmarshal(responseData, &posts)
			c.JSON(http.StatusBadRequest, errResponse)
			return
		}
		_ = json.Unmarshal(responseData, &posts)
		log.Printf("PerPAGE: %d, total retrieved: %d", PerPAGE, len(posts))
		result, err := twitPostCollection.InsertMany(cxt, []interface{}{posts})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.TwitPostResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusCreated, responses.TwitPostResponse{Status: http.StatusCreated, Message: "success", Data: map[string]interface{}{"data": result}})
	}
}

func GetTwitPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		postId := c.Param("postId")
		var post models.TwitPost
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(postId)

		err := twitPostCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&post)

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.TwitPostResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		c.JSON(http.StatusOK, responses.TwitPostResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": post}})
	}
}

func DeleteTwitPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		postId := c.Param("postId")
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(postId)

		result, err := twitPostCollection.DeleteOne(ctx, bson.M{"id": objId})

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.TwitPostResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		if result.DeletedCount < 1 {
			c.JSON(http.StatusNotFound,
				responses.TwitPostResponse{Status: http.StatusNotFound, Message: "error", Data: map[string]interface{}{"data": "Post with specified ID not found!"}},
			)
			return
		}

		c.JSON(http.StatusOK,
			responses.TwitPostResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": "Post successfully deleted!"}},
		)
	}
}

func GetAllTwits() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var posts []models.TwitPost
		defer cancel()

		results, err := twitPostCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.TwitPostResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			return
		}

		//reading from the db in an optimal way
		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleTwitPost models.TwitPost
			if err = results.Decode(&singleTwitPost); err != nil {
				c.JSON(http.StatusInternalServerError, responses.TwitPostResponse{Status: http.StatusInternalServerError, Message: "error", Data: map[string]interface{}{"data": err.Error()}})
			}

			posts = append(posts, singleTwitPost)
		}

		c.JSON(http.StatusOK,
			responses.TwitPostResponse{Status: http.StatusOK, Message: "success", Data: map[string]interface{}{"data": posts}},
		)
	}
}
