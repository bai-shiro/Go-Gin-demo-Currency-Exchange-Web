package repository

import (
	"exchangeapp/internal/models"

	"gorm.io/gorm"
)

//
type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(article *models.Article) error {
	return  r.db.Create(article).Error
}

func (r *ArticleRepository) Update(article *models.Article) error {
	return r.db.Save(article).Error
}

func (r *ArticleRepository) Delete(article *models.Article) error {
	return r.db.Delete(article).Error
}

func (r *ArticleRepository) FindAll() ([]models.Article, error) {
	var articles []models.Article
	if err := r.db.Find(&articles).Error; err != nil {
		return nil, err
	}
	return  articles, nil
}

func (r *ArticleRepository) FindByID(id string) (*models.Article, error) {
	var article models.Article
	if err := r.db.Where("id=?", id).First(&article).Error; err != nil{
		return nil, err
	}
	return &article, nil
}