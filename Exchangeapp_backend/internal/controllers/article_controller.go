package controllers

import (
	"errors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleController struct {
	articles *service.ArticleService
}

func NewArticleController(articles *service.ArticleService) *ArticleController {
	return &ArticleController{articles: articles}
}

func (c *ArticleController) Create(ctx *gin.Context) {
	var article models.Article

	if err := ctx.ShouldBindJSON(&article); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error" : err.Error()})
		return
	}

	if err := c.articles.Create(&article); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, article)
}

func (c *ArticleController) List(ctx *gin.Context) {
	articles, err := c.articles.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, articles)
}

func (c *ArticleController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	article, err := c.articles.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error" : err.Error()})
		return
	} else if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, article)
}