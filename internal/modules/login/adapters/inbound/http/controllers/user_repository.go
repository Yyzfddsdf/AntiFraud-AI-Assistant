package controllers

import (
	"context"
	"fmt"

	"antifraud/internal/modules/login/domain/models"
	"antifraud/internal/platform/database"

	"gorm.io/gorm"
)

// UserRepository 定义认证应用服务依赖的最小用户仓储端口。
type UserRepository interface {
	FindByID(ctx context.Context, userID interface{}) (models.User, error)
	FindByPhone(ctx context.Context, phone string) (models.User, error)
	FindByAccount(ctx context.Context, field string, value string) (models.User, error)
	ExistsByIdentity(ctx context.Context, email string, username string, phone string) (bool, error)
	Create(ctx context.Context, user *models.User) error
	DeleteByID(ctx context.Context, userID interface{}) error
	UpdateRole(ctx context.Context, userID interface{}, role string) error
	List(ctx context.Context, query string) ([]models.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func newGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func defaultUserRepository() UserRepository {
	return newGormUserRepository(database.DB)
}

func (r *gormUserRepository) FindByID(ctx context.Context, userID interface{}) (models.User, error) {
	if r == nil || r.db == nil {
		return models.User{}, fmt.Errorf("main db is not initialized")
	}
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *gormUserRepository) FindByPhone(ctx context.Context, phone string) (models.User, error) {
	if r == nil || r.db == nil {
		return models.User{}, fmt.Errorf("main db is not initialized")
	}
	var user models.User
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *gormUserRepository) FindByAccount(ctx context.Context, field string, value string) (models.User, error) {
	if r == nil || r.db == nil {
		return models.User{}, fmt.Errorf("main db is not initialized")
	}
	var user models.User
	if err := r.db.WithContext(ctx).Where(field+" = ?", value).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *gormUserRepository) ExistsByIdentity(ctx context.Context, email string, username string, phone string) (bool, error) {
	if r == nil || r.db == nil {
		return false, fmt.Errorf("main db is not initialized")
	}
	var existingUser models.User
	err := r.db.WithContext(ctx).
		Where("email = ? OR username = ? OR phone = ?", email, username, phone).
		First(&existingUser).Error
	if err == nil {
		return true, nil
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return false, err
}

func (r *gormUserRepository) Create(ctx context.Context, user *models.User) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("main db is not initialized")
	}
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormUserRepository) DeleteByID(ctx context.Context, userID interface{}) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("main db is not initialized")
	}
	return r.db.WithContext(ctx).Unscoped().Where("id = ?", userID).Delete(&models.User{}).Error
}

func (r *gormUserRepository) UpdateRole(ctx context.Context, userID interface{}, role string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("main db is not initialized")
	}
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("role", role).Error
}

func (r *gormUserRepository) List(ctx context.Context, query string) ([]models.User, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("main db is not initialized")
	}
	var users []models.User
	db := r.db.WithContext(ctx).Model(&models.User{})
	if query != "" {
		db = db.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
