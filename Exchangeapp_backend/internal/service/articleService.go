package service

import (
	"encoding/json"
	"errors"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

// var cacheKey = "articles"

type ArticleService struct {
	articles *repository.ArticleRepository
	redis *redis.Client
}

func NewArticleService(articles *repository.ArticleRepository, redisClient *redis.Client) *ArticleService {
	return &ArticleService{articles: articles, redis: redisClient}
}

func (s *ArticleService) Create(title string, content string, preview string) (*models.Article, error) {
	article := &models.Article{
		Title:   title,
		Content: content,
		Preview: preview,
	}
	if err := s.articles.Create(article); err != nil {
		return nil, err
	}

	s.invalidateArticleCache()

	return article, nil
}

func (s *ArticleService) Update(id string, title string, content string, preview string) (*models.Article, error) {
	article, err := s.articles.FindByID(id)
	if err != nil {
		return nil, err
	}

	article.Title = title
	article.Content = content
	article.Preview = preview

	if err := s.articles.Update(article); err != nil {
		return nil, err
	}

	s.invalidateArticleCache()

	return article, nil
}

func (s *ArticleService) Delete(id string) error {
	article, err := s.articles.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.articles.Delete(article); err != nil {
		return err
	}

	s.invalidateArticleCache()

	return nil
}

func (s *ArticleService) GetByID(id string) (*models.Article,error) {
	article, err := s.articles.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	} else if err != nil{
		return nil, err
	}
	return  article, err
}

func (s *ArticleService) ListPage(page int, pageSize int) ([]models.Article, int64, error) {
	cacheKey := articleListPageCacheKey(page, pageSize)
	cacheData, err := s.redis.Get(cacheKey).Result()

	var payload struct {
		List []models.Article `json:"list"`
		Total int64 `json:"total"`
	}

	if err == redis.Nil {	// 缓存未命中
		// 查询数据库
		articles, total, err := s.articles.FindPage(page, pageSize)
		if err != nil {
			return nil, 0, err
		}

		payload.List = articles
		payload.Total = total

		payloadJson, err := json.Marshal(payload)
		if err != nil {
			return articles, total, nil
		}

		// 存入缓存
		_ = s.redis.Set(cacheKey, payloadJson, 10*time.Minute).Err()

		return articles, total, nil
	} else if err != nil {	// redis查询出错
		return nil, 0, err
	} else {	// 缓存命中
		if err := json.Unmarshal([]byte(cacheData), &payload); err != nil {
			return nil, 0, err
		}

		articles := payload.List
		total := payload.Total

		return  articles, total, nil
	}
}

func (s *ArticleService) Like(id string) error {
	if err := s.redis.Incr(articleLikeKey(id)).Err(); err != nil {
		return err
	}

	return  nil
}

func (s *ArticleService) GetLikes(id string) (string, error) {
	likes, err := s.redis.Get(articleLikeKey(id)).Result(); 
	if err == redis.Nil {
		likes = "0"
	} else if err != nil {
		return "", err
	}
	
	return likes, nil
}

// 清理文章redis缓存
func (s *ArticleService) invalidateArticleCache() {
	keys, err := s.redis.Keys("articles:list:page:*").Result()
	if err == nil && len(keys) > 0 {
		_ = s.redis.Del(keys...).Err()
	}
}

func articleLikeKey(id string) string {
	return "article:" + id + ":like"
}

func articleListPageCacheKey(page int, pageSize int) string {
	return fmt.Sprintf("articles:list:page:%d:size:%d", page, pageSize)
}