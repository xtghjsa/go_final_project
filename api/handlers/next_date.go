package handlers

import (
	"fmt"
	"log"
	"main/api/next_date"
	"net/http"
	"time"
)

// NextDateHandler - Получение даты для повторения задачи
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParameter := r.URL.Query().Get("now")
	dateParameter := r.URL.Query().Get("date")
	repeatParameter := r.URL.Query().Get("repeat")
	now, err := time.Parse("20060102", nowParameter)
	if err != nil {
		http.Error(w, "некорректный формат now", http.StatusBadRequest)
		log.Println("некорректный формат now", err)
		return
	}
	next, err := next_date.RepeatDate(now, dateParameter, repeatParameter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Ошибка при получении даты для повторения", err)
		return
	}
	_, err = fmt.Fprintf(w, next)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при выводе даты", err)
		return
	}
}
