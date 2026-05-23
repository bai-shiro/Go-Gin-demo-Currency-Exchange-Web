package service

import (
	"encoding/json"
	"errors"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var cacheKey = "articles"

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

	if err := s.redis.Del(cacheKey).Err(); err != nil {
		return nil, err
	}

	return article, nil
}

func (s *ArticleService) List() ([]models.Article, error) {
	cacheData, err := s.redis.Get(cacheKey).Result()

	if err == redis.Nil {	// 缓存未命中
		// 查询数据库
		articles, err := s.articles.FindAll()
		if err != nil {
			return nil, err
		}

		artilcesJson, err := json.Marshal(articles)
		if err != nil {
			return articles, err
		}

		// 存入缓存
		if err := s.redis.Set(cacheKey, artilcesJson, 10*time.Minute).Err(); err != nil {
			return articles, err
		}

		return articles, nil
	} else if err != nil {	// redis查询出错
		return nil, err
	} else {	// 缓存命中
		var articles []models.Article
		if err := json.Unmarshal([]byte(cacheData), &articles); err != nil {
			return nil, err
		}

		return  articles, nil
	}
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

func articleLikeKey(id string) string {
	return "article:" + id + ":like"
}