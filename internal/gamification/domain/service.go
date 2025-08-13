package domain

import (
	"context"
	"fmt"

	"github.com/unclaim/chegonado/internal/auth"
)

// gamificationService реализует интерфейс GamificationService.
type gamificationService struct {
	repo GamificationRepository
}

// NewGamificationService создаёт новый сервис геймификации.
func NewGamificationService(repo GamificationRepository) GamificationService {
	return &gamificationService{repo: repo}
}

// HandleUserRegistered — обработчик события регистрации пользователя.
// Вызывается шиной событий.
func (s *gamificationService) HandleUserRegistered(event any) {
	ctx := context.Background() // Используем фоновый контекст для асинхронной операции.
	userEvent, ok := event.(auth.UserRegisteredEvent)
	if !ok {
		fmt.Println("[Gamification] Получено некорректное событие.")
		return
	}

	fmt.Printf("[Gamification] Получено событие регистрации для пользователя: %v\n", userEvent.UserID)

	// Выдача начальных очков опыта (XP).
	err := s.repo.AddExperience(ctx, userEvent.UserID, 100)
	if err != nil {
		fmt.Printf("[Gamification] Ошибка при выдаче XP пользователю %v: %v\n", userEvent.UserID, err)
		return
	}

	fmt.Printf("[Gamification] Пользователю %v успешно выдано 100 XP.\n", userEvent.UserID)
}
