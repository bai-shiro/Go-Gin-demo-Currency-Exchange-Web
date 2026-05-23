package controllers

import (
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/dto"
	"exchangeapp/internal/models"
	"exchangeapp/internal/response"
	"exchangeapp/internal/service"

	"github.com/gin-gonic/gin"
)

type ArticleController struct {
	articles *service.ArticleService
}

func NewArticleController(articles *service.ArticleService) *ArticleController {
	return &ArticleController{articles: articles}
}

func (c *ArticleController) Create(ctx *gin.Context) {
	var req dto.ArticleRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	article, err := c.articles.Create(req.Title, req.Content, req.Preview)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, toArticleResponse(article))
}

func (c *ArticleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var req dto.ArticleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, apperrors.ErrInvalidParams)
		return
	}

	article, err := c.articles.Update(id, req.Title, req.Content, req.Preview)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, toArticleResponse(article))
}

func (c *ArticleController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.articles.Delete(id); err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"message" : "successfully deleted the article"})
}

func (c *ArticleController) List(ctx *gin.Context) {
	articles, err := c.articles.List()
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, toArticleResponses(articles))
}

func (c *ArticleController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	article, err := c.articles.GetByID(id)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, toArticleResponse(article))
}

func (c *ArticleController) Like(ctx *gin.Context) {
	articleID := ctx.Param("id")

	if err := c.articles.Like(articleID); err != nil {
		response.Error(ctx, apperrors.ErrInternal)
		return
	}

	response.Success(ctx, gin.H{"message" : "successfully liked the article"})
}

func (c *ArticleController) GetLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likes, err := c.articles.GetLikes(articleID); 
	if err != nil {
		response.Error(ctx, apperrors.ErrInternal)
		return
	}

	response.Success(ctx, gin.H{"likes" : likes})
}

func toArticleResponse(article *models.Article) dto.ArticleResponse {
	return dto.ArticleResponse{
		ID: article.ID, 
		Title: article.Title, 
		Content: article.Content, 
		Preview: article.Preview, 
	}
}

func toArticleResponses(articles []models.Article) []dto.ArticleResponse {
	res := make([]dto.ArticleResponse, len(articles))
	for i := range articles {
		res[i] = toArticleResponse(&articles[i])
	}
	return res
}