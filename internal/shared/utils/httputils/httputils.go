package httputils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Входной JSON-документ будет автоматически преобразован в объект данного типа
type Request struct {
	Message string `json:"message"`
	Number  int    `json:"number"`
}
type ResponseBody struct {
	Context context.Context `json:"context"`
	Request interface{}     `json:"body,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func Handler(ctx context.Context, request *Request) ([]byte, error) {
	// В логах функции будут напечатаны значения контекста вызова и тела запроса
	fmt.Println("context", ctx)
	fmt.Println("request", request)
	// Объект, содержащий тело ответа, преобразуется в массив байтов
	body, err := json.Marshal(&ResponseBody{
		Context: ctx,
		Request: request,
	})
	if err != nil {
		return nil, err
	}
	// Тело ответа необходимо вернуть в виде массива байтов
	return body, nil

}

func RespJSONError(w http.ResponseWriter, StatusCode int, err error, resp string) {
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(StatusCode)
	w.Header().Add("Content-Type", "application/json")
	body, err := json.Marshal(&ResponseBody{
		Error: resp,
	})
	if err != nil {
		err := fmt.Errorf("an error has occurred when parsing request: %v", err)
		fmt.Println(err)
	}
	// Пишем ответ и проверяем на наличие ошибки
	if _, err := w.Write(body); err != nil {
		// Об обработке ошибки
		http.Error(w, "Ошибка при записи ответа", http.StatusInternalServerError)
		return
	}
}
