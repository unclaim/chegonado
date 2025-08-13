package httputils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RespondJSON makes the response with payload as json format
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Пишем ответ и проверяем на наличие ошибки
	if _, err := w.Write(body); err != nil {
		// Об обработке ошибки
		http.Error(w, "Ошибка при записи ответа", http.StatusInternalServerError)
		return
	}
}

func RenderJSONErr(w http.ResponseWriter, err string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	// Create the JSON error message
	response := fmt.Sprintf(`{"error":"%s"}`, err)

	// Write the response and check for errors
	if _, writeErr := w.Write([]byte(response)); writeErr != nil {
		// Handle the error (e.g., log it)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
