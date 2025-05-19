package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"mime/quotedprintable"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

// ContactForm представляет данные с формы контактов
type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

// Response представляет структуру JSON-ответа
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// enableCORS добавляет заголовки CORS
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с домена
		w.Header().Set("Access-Control-Allow-Origin", "https://bambutcha.github.io")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обработка preflight запросов
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// pingHandler отвечает на запрос для проверки работоспособности сервера
func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// contactHandler обрабатывает запросы с формы контактов
func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var form ContactForm
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		log.Printf("Error decoding request: %v", err)
		sendJSONResponse(w, false, "Некорректный формат запроса", http.StatusBadRequest)
		return
	}

	// Простая валидация
	if form.Name == "" || form.Email == "" || form.Message == "" {
		sendJSONResponse(w, false, "Пожалуйста, заполните все обязательные поля", http.StatusBadRequest)
		return
	}

	if form.Email == "yagadanaga@yandex.ru" || form.Email == "yagadanaga@ya.ru" {
		sendJSONResponse(w, false, "Имя почты не должно совпадать с именем почты автора", http.StatusBadRequest)
		return
	}

	// Получаем настройки SMTP
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	toEmail := os.Getenv("TO_EMAIL")

	if smtpServer == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || toEmail == "" {
		log.Println("Missing SMTP configuration")
		sendJSONResponse(w, false, "Ошибка конфигурации сервера", http.StatusInternalServerError)
		return
	}

	// Формируем тему письма
	subject := "Новое сообщение с сайта"
	if form.Subject != "" {
		subject = form.Subject
	}

	// Создаем буфер для письма
	var msg bytes.Buffer
	
	// Заголовки письма
	msg.WriteString("From: " + smtpUser + "\r\n")
    msg.WriteString("To: " + toEmail + "\r\n")
    msg.WriteString("Subject: " + "[Bambutcha] " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n")
    msg.WriteString("Reply-To: " + form.Email + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString("\r\n") // Пустая строка перед телом

	// Тело письма в quoted-printable
	qp := quotedprintable.NewWriter(&msg)
	qp.Write([]byte(fmt.Sprintf(
		"Имя: %s\nEmail: %s\n\n%s",
		form.Name,
		form.Email,
		form.Message,
	)))
	qp.Close()

	// Настройка TLS
	tlsConfig := &tls.Config{
		ServerName: smtpServer,
	}

	// Подключаемся к SMTP серверу
	conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, tlsConfig)
	if err != nil {
		log.Printf("TLS connection error: %v", err)
		sendJSONResponse(w, false, "Не удалось подключиться к SMTP серверу", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Создаем SMTP клиент
	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		log.Printf("SMTP client error: %v", err)
		sendJSONResponse(w, false, "Не удалось создать SMTP клиент", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Аутентификация
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpServer)
	if err := client.Auth(auth); err != nil {
		log.Printf("SMTP auth error: %v", err)
		sendJSONResponse(w, false, "Ошибка аутентификации SMTP", http.StatusInternalServerError)
		return
	}

	// Устанавливаем отправителя и получателя
	if err := client.Mail(smtpUser); err != nil {
		log.Printf("Mail from error: %v", err)
		sendJSONResponse(w, false, "Ошибка установки отправителя", http.StatusInternalServerError)
		return
	}

	if err := client.Rcpt(toEmail); err != nil {
		log.Printf("Rcpt to error: %v", err)
		sendJSONResponse(w, false, "Ошибка установки получателя", http.StatusInternalServerError)
		return
	}

	// Отправляем данные письма
	wc, err := client.Data()
	if err != nil {
		log.Printf("Data command error: %v", err)
		sendJSONResponse(w, false, "Ошибка подготовки данных", http.StatusInternalServerError)
		return
	}
	defer wc.Close()

	if _, err := wc.Write(msg.Bytes()); err != nil {
		log.Printf("Write error: %v", err)
		sendJSONResponse(w, false, "Ошибка отправки письма", http.StatusInternalServerError)
		return
	}

	// Успешный ответ
	sendJSONResponse(w, true, "Сообщение успешно отправлено", http.StatusOK)
}

// sendJSONResponse отправляет форматированный JSON ответ
func sendJSONResponse(w http.ResponseWriter, success bool, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: success,
		Message: message,
	})
}

func main() {
	// Получаем порт из переменной окружения или используем 8080 по умолчанию
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Регистрируем обработчики запросов
	http.HandleFunc("/ping", enableCORS(pingHandler))
	http.HandleFunc("/api/contact", enableCORS(contactHandler))

	// Запускаем сервер
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
