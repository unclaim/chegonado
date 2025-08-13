package eventbus

import (
	"fmt"
	"reflect"
)

// Event - интерфейс, который должны реализовывать все события.
// Он просто служит для типизации.
type Event interface{}

// EventHandler - тип функции, которая обрабатывает событие.
// Она принимает Event в качестве аргумента.
type EventHandler func(event Event)

// EventBus - простая реализация шины событий "в памяти".
// Она хранит мапу обработчиков для каждого типа события.
type EventBus struct {
	handlers map[reflect.Type][]EventHandler
}

// NewEventBus создает и возвращает новый экземпляр EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[reflect.Type][]EventHandler),
	}
}

// Subscribe регистрирует обработчик для определённого типа события.
func (eb *EventBus) Subscribe(eventType Event, handler EventHandler) {
	// Получаем тип события с помощью рефлексии.
	eventTypeReflect := reflect.TypeOf(eventType)

	// Добавляем обработчик в список для этого типа события.
	eb.handlers[eventTypeReflect] = append(eb.handlers[eventTypeReflect], handler)

	fmt.Printf("Обработчик успешно подписан на событие типа: %s\n", eventTypeReflect)
}

// Publish отправляет событие. Все зарегистрированные обработчики для этого типа события будут вызваны.
func (eb *EventBus) Publish(event Event) {
	// Получаем тип события.
	eventTypeReflect := reflect.TypeOf(event)

	// Проверяем, есть ли обработчики для этого типа.
	if handlers, ok := eb.handlers[eventTypeReflect]; ok {
		// Если есть, то запускаем каждый обработчик в отдельной горутине,
		// чтобы они выполнялись асинхронно и не блокировали друг друга.
		for _, handler := range handlers {
			go handler(event)
		}
	}
}
