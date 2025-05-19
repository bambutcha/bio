package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"mime/quotedprintable"
	"net/http"
	"net/smtp"
	"os"
	"time"
	"mime"
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
		// Разрешаем запросы с вашего домена
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

    // Простая Валидация
    if form.Name == "" || form.Email == "" || form.Message == "" {
        sendJSONResponse(w, false, "Пожалуйста, заполните все обязательные поля", http.StatusBadRequest)
        return
    }

    // Получение конфигурации SMTP
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

    // Формирование письма с правильными заголовками
    subject := "Новое сообщение с сайта"
    if form.Subject != "" {
        subject = form.Subject
    }

    // Формируем MIME-совместимое письмо
    var msg bytes.Buffer
    msg.WriteString("From: " + smtpUser + "\r\n")
    msg.WriteString("To: " + toEmail + "\r\n")
    msg.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", subject) + "\r\n")
    msg.WriteString("Reply-To: " + form.Email + "\r\n")
    msg.WriteString("MIME-Version: 1.0\r\n")
    msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
    msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
    msg.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
    msg.WriteString("\r\n") // Пустая строка между заголовками и телом

    // Кодируем тело письма
    qp := quotedprintable.NewWriter(&msg)
    qp.Write([]byte(fmt.Sprintf(
        "Имя: %s\nEmail: %s\n\n%s",
        form.Name,
        form.Email,
        form.Message,
    )))
    qp.Close()

    // Отправка через TLS
    tlsConfig := &tls.Config{
        ServerName: smtpServer,
    }

    conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, tlsConfig)
    if err != nil {
        log.Printf("TLS connection error: %v", err)
        sendJSONResponse(w, false, "Ошибка подключения к серверу", http.StatusInternalServerError)
        return
    }
    defer conn.Close()

    client, err := smtp.NewClient(conn, smtpServer)
    if err != nil {
        log.Printf("SMTP client error: %v", err)
        sendJSONResponse(w, false, "Ошибка создания SMTP клиента", http.StatusInternalServerError)
        return
    }
    defer client.Close()

    // Аутентификация
    auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpServer)
    if err := client.Auth(auth); err != nil {
        log.Printf("SMTP auth error: %v", err)
        sendJSONResponse(w, false, "Ошибка аутентификации", http.StatusInternalServerError)
        return
    }

    // Установка отправителя и получателя
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

    // Отправка данных
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
