package controllers

import (
	"encoding/json"
	"errors"
	"exchangeapp/internal/global"
	"exchangeapp/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var cacheKey = "articles"

func CreateArticle(ctx *gin.Context){
	var article models.Article

	if err := ctx.ShouldBindJSON(&article); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error" : err.Error()})
		return
	}

	if err := global.Db.AutoMigrate(&article); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	if err := global.Db.Create(&article).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	if err := global.RedisDB.Del(cacheKey).Err(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, article)
}

func GetArticles(ctx *gin.Context) {

	cacheData, err := global.RedisDB.Get(cacheKey).Result()

	if err == redis.Nil {	// 缓存未命中
		var articles []models.Article

		if err := global.Db.Find(&articles).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
			return
		}

		articlesJson, err := json.Marshal(articles)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
			return
		}

		// 存入缓存
		if err := global.RedisDB.Set(cacheKey, articlesJson, 10*time.Minute).Err(); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, articles)

	} else if err != nil {	// 报错
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	} else {	// 缓存命中
		var articles []models.Article
		if err := json.Unmarshal([]byte(cacheData), &articles); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, articles)
	}
}

func GetArticlesByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var article models.Article

	if err := global.Db.Where("id=?", id).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"Error" : err.Error()})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, article)
}