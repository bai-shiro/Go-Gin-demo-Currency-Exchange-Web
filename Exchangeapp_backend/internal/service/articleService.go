package service

import (
	"encoding/json"
	"errors"
	"exchangeapp/internal/apperrors"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type articleListCachePayload struct {
	List  []models.Article `json:"list"`
	Total int64            `json:"total"`
}

type ArticleService struct {
	articles *repository.ArticleRepository
	redis    *redis.Client
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

	if err := s.invalidateArticleCache(); err != nil {
		log.Printf("failed to invalidate article cache: %v", err) // 文章创建成功了，缓存失效失败了，不应该影响正常流程，记录日志但不返回错误
	}

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

	if err := s.invalidateArticleCache(); err != nil {
		log.Printf("failed to invalidate article cache: %v", err)
	}

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

	if err := s.invalidateArticleCache(); err != nil {
		log.Printf("failed to invalidate article cache: %v", err)
	}

	// 删除文章相关的点赞缓存
	if err := s.redis.Del(articleLikedUsersKey(id), articleLikeKey(id)).Err(); err != nil {
		log.Printf("failed to delete article like cache: %v", err)
	}

	return nil
}

func (s *ArticleService) GetByID(id string) (*models.Article, error) {
	article, err := s.articles.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return article, err
}

func (s *ArticleService) ListPage(page int, pageSize int) ([]models.Article, int64, error) {
	cacheKey := articleListPageCacheKey(page, pageSize)
	cacheData, err := s.redis.Get(cacheKey).Result()

	if err == nil { // 缓存命中
		var payload articleListCachePayload
		// 反序列化缓存数据+坏缓存回源处理
		if err := json.Unmarshal([]byte(cacheData), &payload); err != nil {
			return s.listPageFromDB(page, pageSize, cacheKey)
		}

		articles := payload.List
		total := payload.Total

		return articles, total, nil
	}

	if err != redis.Nil { // redis查询出错
		return s.listPageFromDB(page, pageSize, cacheKey)
	}

	// 缓存未命中，查询数据库
	return s.listPageFromDB(page, pageSize, cacheKey)
}

func (s *ArticleService) Like(articleID string, userID uint) (bool, int64, error) {
	likedUserKey := articleLikedUsersKey(articleID)
	likeCountKey := articleLikeKey(articleID)
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	// 使用lua脚本保证redis的原子性
	const toggleLikeLuaScript = `
	local liked = redis.call("SADD", KEYS[1], ARGV[1])
	if liked == 0 then
		redis.call("SREM", KEYS[1], ARGV[1])
		local likes = redis.call("DECR", KEYS[2])
		if likes < 0 then
			likes = 0
			redis.call("SET", KEYS[2], 0)
		end
		return {0, likes}
	else
		local likes = redis.call("INCR", KEYS[2])
		return {1, likes}
	end
	`

	result, err := s.redis.Eval(toggleLikeLuaScript, []string{likedUserKey, likeCountKey}, userIDStr).Result()
	if err != nil {
		return false, 0, err
	}

	resArr, ok := result.([]interface{})
	if !ok || len(resArr) != 2 {
		return false, 0, fmt.Errorf("unexpected lua script result: %v", result)
	}

	likedInt, ok := resArr[0].(int64)
	if !ok {
		return false, 0, fmt.Errorf("unexpected liked value type: %T", resArr[0])
	}

	likes, ok := resArr[1].(int64)
	if !ok {
		return false, 0, fmt.Errorf("unexpected likes value type: %T", resArr[1])
	}

	return likedInt == 1, likes, nil

	// 上面是用lua脚本实现的原子操作，下面是分步实现的，可能会有并发问题，点赞数不准确；如果要用分步实现，需要加分布式锁保证原子性，性能可能会有影响；

	// liked, err := s.redis.SAdd(likedUserKey, username).Result() // 记录点赞过的用户，集合类型，自动去重,判断是否点过赞
	// if err != nil {
	// 	return false, 0, err
	// }

	// if liked == 0 { // 已经点过赞了，再次点赞是取消点赞
	// 	if err := s.redis.SRem(likedUserKey, username).Err(); err != nil {
	// 		return false, 0, err
	// 	}

	// 	likes, err := s.redis.Decr(likeCountKey).Result() // 点赞数-1
	// 	if err != nil {
	// 		return false, 0, err
	// 	}

	// 	return false, likes, nil
	// }

	// likes, err := s.redis.Incr(likeCountKey).Result() // 点赞数+1
	// if err != nil {
	// 	return false, 0, err
	// }
	// return true, likes, nil
}

func (s *ArticleService) GetLikes(id string) (int64, error) {
	likes, err := s.redis.Get(articleLikeKey(id)).Int64()
	if err == redis.Nil {
		likes = 0
	} else if err != nil {
		return 0, err
	}

	return likes, nil
}

// 清理文章redis缓存(分页列表)，目前没有做细粒度的缓存失效，直接把所有分页列表相关的key都删除掉；如果文章量很大，分页列表很多，这种做法可能会有性能问题，可以考虑更细粒度的缓存失效策略
func (s *ArticleService) invalidateArticleCache() error {
	const pattern = "articles:list:page:*"
	const batchSize = 100

	var cursor uint64
	var keys []string

	for {
		scanKeys, cursor, err := s.redis.Scan(cursor, pattern, batchSize).Result()
		if err != nil {
			return err
		}

		if len(scanKeys) > 0 {
			keys = append(keys, scanKeys...)
			if len(keys) >= batchSize { // 一次RTT删除过多key可能会有性能问题，分批删除；过少又会有过多RTT，权衡后定为100
				if err := s.redis.Del(keys...).Err(); err != nil {
					return err
				}
				keys = keys[:0] // 清空切片
			}
		}

		if cursor == 0 {
			break
		}
	}

	// 剩下的key也删除掉
	if len(keys) > 0 {
		if err := s.redis.Del(keys...).Err(); err != nil {
			return err
		}
	}

	return nil
}

// 查询数据库获取文章列表(分页)
func (s *ArticleService) listPageFromDB(page int, pageSize int, cacheKey string) ([]models.Article, int64, error) {
	articles, total, err := s.articles.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	payload := articleListCachePayload{
		List:  articles,
		Total: total,
	}

	payloadJson, err := json.Marshal(payload)
	if err == nil { // 存入缓存
		_ = s.redis.Set(cacheKey, payloadJson, 10*time.Minute).Err()
	}

	return articles, total, nil
}

func articleLikeKey(id string) string {
	return "article:" + id + ":like"
}

func articleLikedUsersKey(id string) string {
	return "article:" + id + ":liked_users"
}

func articleListPageCacheKey(page int, pageSize int) string {
	return fmt.Sprintf("articles:list:page:%d:size:%d", page, pageSize)
}
