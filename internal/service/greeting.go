package service

import (
	"fmt"
	"test/pkg/token"
)

// Greeting is the domain response for hello endpoints.
type Greeting struct {
	Message string `json:"message"`
	Version string `json:"version"`
}

// GreetingService defines greeting-related business logic.
type GreetingService interface {
	Hello() Greeting
	HelloName(name string) Greeting
	MyToken() string
}

type greetingService struct {
	version   string
	secretKey string
}

// NewGreetingService constructs a GreetingService.
func NewGreetingService(version string, secretKey string) GreetingService {
	return &greetingService{version: version, secretKey: secretKey}
}

func (s *greetingService) Hello() Greeting {
	return Greeting{
		Message: "Hello, World!",
		Version: s.version,
	}
}

func (s *greetingService) HelloName(name string) Greeting {
	if name == "" {
		name = "World"
	}
	return Greeting{
		Message: fmt.Sprintf("Hello, %s!", name),
		Version: s.version,
	}
}

func (s *greetingService) MyToken() string {
	token, err := token.CreateToken(s.secretKey)
	if err != nil {
		return ""
	}
	return token
}
