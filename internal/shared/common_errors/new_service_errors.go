// Сделать во всех сервисах единый подход к возврату ошибок; необходимо везде реализовать
// чёткое разграничения ошибок необходимо чётко разделить ошибки которые возникли в самом сервисе и
// ошибки которые пришли из репозитария; сделать все ошибки
// более информативные чтобы по ним было легче отлаживать приложение улучшить и показать полный обновлённый код

// Использовать струтктуру которая находится по
// /internal/shared/common_errors/new_service_errors.go
package common_errors

import (
	"fmt"
)

// ServiceError — это универсальная структура для ошибок уровня сервиса.
// Она содержит сообщение для пользователя (или для логирования) и может
// оборачивать исходную ошибку, пришедшую, например, из репозитория.
type ServiceError struct {
	Msg string
	Err error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

// Unwrap позволяет errors.Unwrap получить исходную ошибку.
func (e *ServiceError) Unwrap() error {
	return e.Err
}

// NewServiceError создает новую ошибку сервиса с заданным сообщением.
// Используется для ошибок, которые возникают на уровне бизнес-логики.
func NewServiceError(msg string) error {
	return &ServiceError{Msg: msg}
}

// WrapServiceError оборачивает исходную ошибку `err` в ServiceError
// с новым сообщением `msg`. Полезно, когда нужно добавить контекст к
// ошибке, пришедшей из другого слоя (например, из репозитория).
func WrapServiceError(msg string, err error) error {
	return &ServiceError{Msg: msg, Err: err}
}
