package session

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// GeoInfo структура для хранения информации о геолокации
type GeoInfo struct {
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
}

// GeoCache структура для хранения кэша GeoIP
type GeoCache struct {
	mu         sync.RWMutex
	cache      map[string]cachedGeoInfo
	ttl        time.Duration // Время жизни кэша
	lastUpdate map[string]time.Time
}

// cachedGeoInfo структура для хранения информации о геолокации и времени кэширования
type cachedGeoInfo struct {
	info GeoInfo
}

// Инициализация кэша
var geoCache = GeoCache{
	cache:      make(map[string]cachedGeoInfo),
	ttl:        10 * time.Minute, // Установим TTL для кэша 10 минут
	lastUpdate: make(map[string]time.Time),
}

// Функция для извлечения браузера из User-Agent
func parseBrowser(userAgent string) string {
	browserRegex := map[string]*regexp.Regexp{
		"Chrome":            regexp.MustCompile(`(?i)Chrome`),
		"Firefox":           regexp.MustCompile(`(?i)Firefox`),
		"Safari":            regexp.MustCompile(`(?i)Safari`),
		"Edge":              regexp.MustCompile(`(?i)Edg`),
		"Opera":             regexp.MustCompile(`(?i)Opera|OPR`),
		"Internet Explorer": regexp.MustCompile(`(?i)MSIE|Trident`),
	}

	for name, regex := range browserRegex {
		if regex.MatchString(userAgent) {
			return name
		}
	}

	return "Unknown"
}

// Функция для извлечения операционной системы из User-Agent
func parseOperatingSystem(userAgent string) string {
	osRegex := map[string]*regexp.Regexp{
		"Windows": regexp.MustCompile(`(?i)Windows`),
		"Mac OS":  regexp.MustCompile(`(?i)Macintosh|Mac OS X`),
		"Linux":   regexp.MustCompile(`(?i)Linux`),
		"Android": regexp.MustCompile(`(?i)Android`),
		"iOS":     regexp.MustCompile(`(?i)iPhone|iPad`),
	}

	for name, regex := range osRegex {
		if regex.MatchString(userAgent) {
			return name
		}
	}

	return "Unknown"
}

// Функция для извлечения браузера и операционной системы из User-Agent
func parseUserAgent(userAgent string) (string, string) {
	browser := parseBrowser(userAgent)
	os := parseOperatingSystem(userAgent)
	return browser, os
}

// Функция для получения IP-адреса клиента
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return strings.Split(ip, ",")[0] // Если есть список IP, берем первый
}

// Функция для получения информации о городе и местоположении по IP-адресу
func getGeoInfo(ip string) (string, string, error) {
	// Проверка на наличие кэшированной информации
	geoCache.mu.RLock()
	if cachedInfo, found := geoCache.cache[ip]; found {
		if time.Since(geoCache.lastUpdate[ip]) < geoCache.ttl { // Проверка времени кэширования
			geoCache.mu.RUnlock()
			return cachedInfo.info.City, formatLocation(cachedInfo.info.Latitude, cachedInfo.info.Longitude), nil
		}
	}
	geoCache.mu.RUnlock()

	// Если не нашли в кэше, делаем запрос к API
	resp, err := http.Get("http://ip-api.com/json/" + ip)
	if err != nil {
		return "", "", err
	}
	// Создаем отложенный вызов Close(), проверяя возможную ошибку
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			log.Printf("Error closing body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var errorMsg struct{ Message string }
		err := json.NewDecoder(resp.Body).Decode(&errorMsg) // Чтение ошибки из ответа
		if err != nil {                                     // Проверка, возникла ли ошибка при декодировании
			return "", "", err // Возвращаем ошибку
		}
		return "", "", fmt.Errorf("error from server: %s", errorMsg.Message) // Возвращаем сообщение об ошибке
	}

	var result GeoInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	// Кэшируем результат
	geoCache.mu.Lock()
	geoCache.cache[ip] = cachedGeoInfo{info: result}
	geoCache.lastUpdate[ip] = time.Now() // Обновляем время кэширования
	geoCache.mu.Unlock()

	return result.City, formatLocation(result.Latitude, result.Longitude), nil
}

// Форматирование местоположения
func formatLocation(lat, long float64) string {
	return fmt.Sprintf("%.6f, %.6f", lat, long)
}
