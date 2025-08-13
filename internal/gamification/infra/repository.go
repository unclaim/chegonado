package infra

import (
	"context"
	"fmt"
)

// GamificationRepository — заглушка для репозитория.
type GamificationRepository struct{}

// NewGamificationRepository создаёт новый репозиторий.
func NewGamificationRepository() *GamificationRepository {
	return &GamificationRepository{}
}

// AddExperience — заглушка метода для добавления опыта.
func (r *GamificationRepository) AddExperience(ctx context.Context, userID int64, amount int) error {
	// В реальной жизни здесь будет код для работы с базой данных.
	fmt.Printf("[GamificationRepo] Добавление %d XP пользователю %v\n", amount, userID)
	return nil
}
