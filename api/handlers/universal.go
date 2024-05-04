package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
)

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskR struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskManager struct {
	DB     *sql.DB
	search string
}

// ErrorResponse - Форма для ответа на ошибку в json
func ErrorResponse(errorText string) []byte {
	jsonErr, err := json.Marshal(map[string]string{"error": errorText})
	if err != nil {
		log.Println("Ошибка при записи в json", err)
	}
	return jsonErr
}
