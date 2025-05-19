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
	"strings"
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

func encodeSubject(subject string) string {
	var encoded bytes.Buffer
	w := quotedprintable.NewWriter(&encoded)
	w.Write([]byte(subject))
	w.Close()
	
	return fmt.Sprintf("=?UTF-8?Q?%s?=", 
		strings.ReplaceAll(encoded.String(), " ", "_"))
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
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&form); err != nil {
		log.Printf("Error decoding request: %v", err)
		sendJSONResponse(w, false, "Некорректный формат запроса", http.StatusBadRequest)
		return
	}

	// Простая валидация
	if form.Name == "" || form.Email == "" || form.Message == "" {
		sendJSONResponse(w, false, "Пожалуйста, заполните все обязательные поля", http.StatusBadRequest)
		return
	}

	// Получаем настройки SMTP из переменных окружения
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

	// Формируем сообщение
	subject := form.Subject
	if subject == "" {
		subject = "Новое сообщение с сайта"
	}

	mailBody := fmt.Sprintf(
		"Имя: %s\nEmail: %s\n\n%s",
		form.Name,
		form.Email,
		form.Message,
	)

	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("From: %s\r\n", smtpUser))
	body.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	body.WriteString(fmt.Sprintf("Reply-To: %s\r\n", form.Email))
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	body.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	body.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	body.WriteString("\r\n")

	qpWriter := quotedprintable.NewWriter(&body)
	qpWriter.Write([]byte(mailBody))
	qpWriter.Close()

	message := body.Bytes()

	// Настраиваем SMTP аутентификацию
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpServer)

	// Отправляем email
	var err error
	if smtpPort == "465" {
		// Для портов SSL (465) - Яндекс использует этот порт
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         smtpServer,
		}
		conn, dialErr := tls.Dial("tcp", fmt.Sprintf("%s:%s", smtpServer, smtpPort), tlsConfig)
		if dialErr != nil {
			log.Printf("Failed to dial SMTP server: %v", dialErr)
			sendJSONResponse(w, false, "Не удалось подключиться к SMTP серверу", http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		
		client, clientErr := smtp.NewClient(conn, smtpServer)
		if clientErr != nil {
			log.Printf("Failed to create SMTP client: %v", clientErr)
			sendJSONResponse(w, false, "Не удалось создать SMTP клиент", http.StatusInternalServerError)
			return
		}
		defer client.Close()
		
		if authErr := client.Auth(auth); authErr != nil {
			log.Printf("SMTP authentication failed: %v", authErr)
			sendJSONResponse(w, false, "Ошибка аутентификации SMTP", http.StatusInternalServerError)
			return
		}
		
		if fromErr := client.Mail(smtpUser); fromErr != nil {
			log.Printf("Failed to set sender: %v", fromErr)
			sendJSONResponse(w, false, "Ошибка установки отправителя", http.StatusInternalServerError)
			return
		}
		
		if rcptErr := client.Rcpt(toEmail); rcptErr != nil {
			log.Printf("Failed to set recipient: %v", rcptErr)
			sendJSONResponse(w, false, "Ошибка установки получателя", http.StatusInternalServerError)
			return
		}
		
		wc, dataErr := client.Data()
		if dataErr != nil {
			log.Printf("Failed to start data command: %v", dataErr)
			sendJSONResponse(w, false, "Ошибка запуска команды данных", http.StatusInternalServerError)
			return
		}
		defer wc.Close()
		
		_, err = wc.Write(message)
	} else {
		// Для портов TLS (587, 25, etc.)
		err = smtp.SendMail(
			fmt.Sprintf("%s:%s", smtpServer, smtpPort),
			auth,
			smtpUser,
			[]string{toEmail},
			message,
		)
	}

	if err != nil {
		log.Printf("Failed to send email: %v", err)
		sendJSONResponse(w, false, "Не удалось отправить email", http.StatusInternalServerError)
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
