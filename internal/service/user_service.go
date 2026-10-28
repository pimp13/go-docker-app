package service

import (
	"errors"
	"fmt"
	"strings"
	"test/internal/dto"
	"test/internal/model"
	"test/pkg/token"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	CreateUser(req dto.UserCreateRequest, meta map[string]any) (*dto.UserCreateResponse, error)

	GetAllUsers() ([]model.User, error)

	SecretKey() string

	GetUserByID(id string) (*model.User, error)
}

type userService struct {
	db        *gorm.DB
	secretKey string
}

func NewUserService(db *gorm.DB, secretKey string) UserService {
	return &userService{db: db, secretKey: secretKey}
}

func (s *userService) SecretKey() string {
	return s.secretKey
}

func (s *userService) CreateUser(
	req dto.UserCreateRequest,
	meta map[string]any,
) (*dto.UserCreateResponse, error) {
	// نرمال‌سازی
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Username = strings.ToLower(strings.TrimSpace(req.Username))

	// Hash token using SHA-256
	tokenStr, exists := meta["token"]
	if exists {
		meta["token"] = tokenStr.(string)
		meta["token_hash"] = token.HashToken(tokenStr.(string))
	} else {
		return nil, errors.New("token is required")
	}

	// بررسی تکراری بودن ایمیل
	var count int64
	if err := s.db.Model(&model.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("email already exists")
	}

	// بررسی تکراری بودن یوزرنیم (اگر ارسال شده باشد)
	if req.Username != "" {
		count = 0
		if err := s.db.Model(&model.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("username already exists")
		}
	}

	// ساخت یوزر
	user := model.User{
		ID:        uuid.New(),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		Phone:     strings.TrimSpace(req.Phone),
		Role:      model.UserRoleUser,
		Status:    model.UserStatusActive,
		Meta:      meta,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &dto.UserCreateResponse{
		Message: "User created successfully!",
		UserID:  user.ID.String(),
	}, nil
}

func (s *userService) GetAllUsers() ([]model.User, error) {
	var users []model.User
	if err := s.db.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %v", err)
	}
	return users, nil
}

func (s *userService) GetUserByID(id string) (*model.User, error) {
	var user model.User
	if err := s.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}
	return &user, nil
}
