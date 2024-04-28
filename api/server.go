package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// EnvInit Загружает переменные из файла .env
func EnvInit() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("файл .env не найден")
		_, err = os.Create(".env")
		if err != nil {
			log.Println("не удалось создать файл .env")
		}
	}
}

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// StartServer Запускает сервер
func StartServer() {

	defaultPort := "7540"
	webDir := "./web"
	// При наличии файла .env c заданным значением переменной TODO_PORT сервер запустится на указанном порту
	TODO_PORT, exists := os.LookupEnv("TODO_PORT")
	if exists && TODO_PORT != "" {
		fmt.Printf("Сервер запущен на localhost:%s", TODO_PORT+" по порту указанному в .env")
		http.Handle("/", http.FileServer(http.Dir(webDir)))
		http.HandleFunc("/api/nextdate", nextDateHandler)
		http.HandleFunc("/api/tasks", tasksHandler)
		http.HandleFunc("/api/task", taskManagerHandler)
		err := http.ListenAndServe(":"+TODO_PORT, nil)
		if err != nil {
			panic(err)
		}
	}
	// Если переменная не задана, по умолчанию будет использован стандартный порт 7540
	fmt.Printf("Сервер запущен на localhost:%s", defaultPort+" по умолчанию")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task", taskManagerHandler)
	err := http.ListenAndServe(":7540", nil)
	if err != nil {
		panic(err)
	}

}
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParameter := r.URL.Query().Get("now")
	dateParameter := r.URL.Query().Get("date")
	repeatParameter := r.URL.Query().Get("repeat")
	now, err := time.Parse("20060102", nowParameter)
	if err != nil {
		http.Error(w, "некорректный формат now", http.StatusBadRequest)
		return
	}
	next, err := NextDate(now, dateParameter, repeatParameter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = fmt.Fprintf(w, next)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func taskManagerHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":

	case "POST":

		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		rules := strings.Split(task.Repeat, " ")
		if rules[0] != "d" && rules[0] != "w" && rules[0] != "m" && rules[0] != "y" {
			http.Error(w, "неправильный формат повтора", http.StatusBadRequest)
			log.Println("неправильный формат повтора")
			return
		}
		if rules[0] != "y" && len(rules) == 1 {
			http.Error(w, "выбранные даты не могут быть пустыми", http.StatusBadRequest)
			log.Println("выбранные даты не могут быть пустыми")
			return
		}
		if task.Title == "" {
			http.Error(w, "не задано название задачи", http.StatusBadRequest)
			log.Println("не задано название задачи")
			return
		}
		if task.Date == "" {
			http.Error(w, "дате не была задана, подставленная сегодняшняя", http.StatusBadRequest)
			task.Date = time.Now().Format("20060102")
			log.Println("дате не была задана, подставленная сегодняшняя")
			return
		}
		timeTime, err := time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if timeTime.Before(time.Now()) {
			if task.Repeat == "" {
				task.Date = time.Now().Format("20060102")
			}
			task.Date, err = NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		_, err = time.Parse("20060102", task.Date)
		if err != nil {
			http.Error(w, "некорректный формат даты", http.StatusBadRequest)
			log.Println("некорректный формат даты")
			return
		}
		db, err := sql.Open("sqlite", "../go_final_project/database/scheduler.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer db.Close()
		stmt, err := db.Prepare("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		res, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(map[string]int64{"id": id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonResponse)

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "не задан id", http.StatusBadRequest)
			return
		}
		db, err := sql.Open("sqlite", "../go_final_project/database/scheduler.db")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		stmt, err := db.Prepare("DELETE FROM scheduler WHERE id=?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "неизвестный метод", http.StatusMethodNotAllowed)
	}
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite", "../go_final_project/database/scheduler.db")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 30 ")
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
	if err := rows.Err(); err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	type TaskResponse struct {
		Tasks []Task `json:"tasks"`
	}
	tasksResponse := TaskResponse{Tasks: []Task{}}
	for _, task := range tasksMap {
		tasksResponse.Tasks = append(tasksResponse.Tasks, Task{
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
