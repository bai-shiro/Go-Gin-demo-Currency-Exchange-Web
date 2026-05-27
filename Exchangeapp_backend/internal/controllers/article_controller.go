package controllers

import (
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/dto"
	"exchangeapp/internal/models"
	"exchangeapp/internal/response"
	"exchangeapp/internal/service"
	"strconv"

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

	response.Success(ctx, gin.H{"message": "successfully deleted the article"})
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

func (c *ArticleController) ListPage(ctx *gin.Context) {
	page := parseIntWithDefault(ctx.Query("page"), 1)
	pageSize := parseIntWithDefault(ctx.Query("page_size"), 10)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	articles, total, err := c.articles.ListPage(page, pageSize)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, response.Page{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		List:     toArticleResponses(articles),
	})
}

func (c *ArticleController) Like(ctx *gin.Context) {
	articleID := ctx.Param("id")

	userIDValue, ok := ctx.Get("userID")
	if !ok {
		response.Error(ctx, apperrors.ErrUnauthorized)
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok || userID == 0 {
		response.Error(ctx, apperrors.ErrUnauthorized)
		return
	}

	liked, likes, err := c.articles.Like(articleID, userID)
	if err != nil {
		response.Error(ctx, apperrors.ErrInternal)
		return
	}

	response.Success(ctx, gin.H{
		"liked": liked,
		"likes": likes,
	})
}

func (c *ArticleController) GetLikes(ctx *gin.Context) {
	articleID := ctx.Param("id")

	likes, err := c.articles.GetLikes(articleID)
	if err != nil {
		response.Error(ctx, apperrors.ErrInternal)
		return
	}

	response.Success(ctx, gin.H{"likes": likes})
}

func toArticleResponse(article *models.Article) dto.ArticleResponse {
	return dto.ArticleResponse{
		ID:      article.ID,
		Title:   article.Title,
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

func parseIntWithDefault(s string, defaultVal int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
