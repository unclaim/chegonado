package domain

import (
	"context"
)

// GamificationService — интерфейс для бизнес-логики геймификации.
type GamificationService interface {
	HandleUserRegistered(event any)
}

// GamificationRepository — интерфейс для доступа к данным.
type GamificationRepository interface {
	AddExperience(ctx context.Context, userID int64, amount int) error
}
