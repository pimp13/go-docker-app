package funmod

import "gorm.io/gorm"

type Service interface {
	Funmod(param string) (string, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db}
}

func (s *service) Funmod(param string) (string, error) {
	return "Hello, " + param, nil
}
