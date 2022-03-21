package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	helpers "twitter-scrapy/helper"

	"twitter-scrapy/configs"
	"twitter-scrapy/models"
	"twitter-scrapy/responses"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var twitPostCollection *mongo.Collection = configs.GetCollection(configs.DB, "posts")

const PerPAGE = "100"

func SyncPosts() gin.HandlerFunc {
	return func(c *gin.Context) {
		cxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		twitURL, _ := url.Parse(os.Getenv("TWITTER_URL"))
		searchPath := &url.URL{Path: os.Getenv("SEARCH_PATH")}
		searchPathURL := twitURL.ResolveReference(searchPath)
		chanMessages := make(chan string)
		 var coinsStr string
		coinsStr = os.Getenv("COINS")
		if os.Getenv("COINS") == "" {
			coinsStr = "btc,eth,sol,xtz"
		}
		coins := strings.Split(coinsStr, ",")
		for _, coin := range coins {
			coin := coin
			go func(jobName string) {
				err = saveTwitsIntoDB(cxt, twitterClient, searchPathURL, coin)
				fmt.Println(err)
				chanMessages <- coin + " is synced!"
			}("Syncing btc..... \n")
		}
		for i := range chanMessages {
			log.Println(i)
		}
		defer cancel()
		c.JSON(http.StatusCreated, responses.TwitPostResponse{Status: http.StatusCreated, Message: "success"})
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

func saveTwitsIntoDB(cxt context.Context, twitterClient helpers.HTTPClient, path *url.URL, keyword string) error {
	searchPathQ := path.Query()
	searchPathQ.Set("query", keyword)
	searchPathQ.Set("tweet.fields", "source,created_at")
	searchPathQ.Set("max_results", "100")
	path.RawQuery = searchPathQ.Encode()
	responseData, responseStatusCode, err := twitterClient.MakeRequest(cxt, "GET", path, nil)
	if responseStatusCode != 200 {
		var errResponse responses.ErrorResponse
		_ = json.Unmarshal(responseData, &errResponse)
		log.Printf("reponse %v", string(responseData))
		return fmt.Errorf("bad Request")
	}
	var searchResponse responses.RecentSearchAPIResponse
	err = json.Unmarshal(responseData, &searchResponse)
	if err != nil {
		log.Printf("Unmarshal err: %v", err)
		return err
	}
	log.Printf("PerPAGE: %d, total retrieved: %d", PerPAGE, len(searchResponse.Data))
	_, err = twitPostCollection.InsertMany(cxt, searchResponse.Data)
	if err != nil {
		log.Printf("failed to save into DB %v", err)
		return err
	}
	if searchResponse.Meta.NextToken != "" {
		if err = saveTwitsIntoDB(cxt, twitterClient, path,keyword); err != nil {
			return err
		}
	}
	return nil
}
