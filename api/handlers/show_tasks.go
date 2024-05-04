package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// ShowTasksHandler - Получение списка задач. Поиск задач по дате/заголовку/комментарию
func (d *TaskManager) ShowTasksHandler(w http.ResponseWriter, r *http.Request) {
	searchRequest := r.URL.Query().Get("search")
	search := &TaskManager{DB: d.DB, search: searchRequest}
	if searchRequest != "" {
		search.taskSearch(w)
	} else {
		d.showAllTasks(w)
	}
}

func (d *TaskManager) taskSearch(w http.ResponseWriter) {
	searchTime, _ := time.Parse("02.01.2006", d.search)
	searchTimeString := searchTime.Format("20060102")
	d.search = "%" + d.search + "%"

	rows, err := d.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? OR title LIKE ? OR comment LIKE ? LIMIT 30", searchTimeString, d.search, d.search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	tasksMap := make(map[string]map[string]string)
	for rows.Next() {
		var id, date, title, comment, repeat string
		err := rows.Scan(&id, &date, &title, &comment, &repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		task := map[string]string{
			"id":      id,
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		}
		tasksMap[id] = task
	}
	if err := rows.Err(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	type TaskResponse struct {
		Tasks []TaskR `json:"tasks"`
	}
	tasksResponse := TaskResponse{Tasks: []TaskR{}}
	for _, task := range tasksMap {
		tasksResponse.Tasks = append(tasksResponse.Tasks, TaskR{
			ID:      task["id"],
			Date:    task["date"],
			Title:   task["title"],
			Comment: task["comment"],
			Repeat:  task["repeat"],
		})
	}
	jsonResponse, err := json.Marshal(tasksResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при записи в json", err)
		return
	}

	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка возвращения ответа", err)
	}
}

func (d *TaskManager) showAllTasks(w http.ResponseWriter) {
	rows, err := d.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 30 ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	tasksMap := make(map[string]map[string]string)
	for rows.Next() {
		var id, date, title, comment, repeat string
		err := rows.Scan(&id, &date, &title, &comment, &repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		task := map[string]string{
			"id":      id,
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		}
		tasksMap[id] = task
	}
	if err := rows.Err(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	type TaskResponse struct {
		Tasks []TaskR `json:"tasks"`
	}
	tasksResponse := TaskResponse{Tasks: []TaskR{}}
	for _, task := range tasksMap {
		tasksResponse.Tasks = append(tasksResponse.Tasks, TaskR{
			ID:      task["id"],
			Date:    task["date"],
			Title:   task["title"],
			Comment: task["comment"],
			Repeat:  task["repeat"],
		})
	}
	jsonResponse, err := json.Marshal(tasksResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка возвращения ответа", err)
	}
}
