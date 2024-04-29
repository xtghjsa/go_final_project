package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func errorResponse(errorText string) []byte {
	jsonErr, err := json.Marshal(map[string]string{"error": errorText})
	if err != nil {
		log.Println(err)
	}
	return jsonErr
}

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

// StartServer Запускает сервер
func StartServer() {

	Port := "7540"
	webDir := "./web"
	TODO_PORT, exists := os.LookupEnv("TODO_PORT")
	if exists && TODO_PORT != "" {
		Port = TODO_PORT
	}
	// При наличии файла .env c заданным значением переменной TODO_PORT сервер запустится на указанном порту если переменная не задана, то будет использован стандартный порт 7540
	fmt.Printf("Сервер запущен на localhost:%s", Port)
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/tasks", showTasksHandler)
	http.HandleFunc("/api/task", taskManagerHandler)
	http.HandleFunc("/api/task/done", markTaskAsDoneHandler)
	err := http.ListenAndServe(":"+TODO_PORT, nil)
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

// api/task
func taskManagerHandler(w http.ResponseWriter, r *http.Request) {
	var dbPath = "./database/scheduler.db"
	var TODO_DBFILE, exists = os.LookupEnv("TODO_DBFILE")
	if exists && TODO_DBFILE != "" {
		dbPath = TODO_DBFILE
	}
	switch r.Method {
	case "GET":
		id := r.URL.Query().Get("id")
		if id == "" {
			w.Write(errorResponse("id не указан"))
			return
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?", id)

		var taskId, date, title, comment, repeat string

		err = row.Scan(&taskId, &date, &title, &comment, &repeat)
		task := map[string]string{
			"id":      taskId,
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		}
		if taskId == "" {
			w.Write(errorResponse("Задача не найдена"))
			return
		}
		jsonResponse, err := json.Marshal(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)

	case "POST":

		var task Task

		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			w.Write(errorResponse("Ошибка при декодировании запроса"))
			log.Println("Ошибка при декодировании запроса")
			return
		}
		if task.Title == "" {
			w.Write(errorResponse("Не указано название задачи"))
			log.Println("Не указано название задачи")
			return
		}
		if task.Date == "" {
			task.Date = time.Now().Format("20060102")
		}
		dateTime, err := time.Parse("20060102", task.Date)
		if err != nil {
			w.Write(errorResponse("Некорректный формат даты"))
			log.Println("Некорректный формат даты")
			return
		}
		if dateTime.Format("20060102") == time.Now().Format("20060102") {
			dateTime = time.Now()
		}
		if dateTime.Before(time.Now()) {
			if task.Repeat == "" {
				task.Date = time.Now().Format("20060102")
			} else if task.Repeat != "" {
				task.Date, err = NextDate(time.Now(), time.Now().Format("20060102"), task.Repeat)
				if err != nil {
					w.Write(errorResponse("Некорректный формат повторения"))
					log.Println("Некорректный формат повторения")
					return
				}
			}
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных")
			return
		}
		defer db.Close()
		stmt, err := db.Prepare("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при создании записи в базе данных")
			return
		}
		res, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в базу данных")
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при получении id записи в базе данных")
			return
		}
		idStr := strconv.Itoa(int(id))
		jsonId, err := json.Marshal(map[string]string{"id": idStr})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи id в json")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(jsonId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при возвращении ответа")
			return
		}

	case "DELETE":
		id := r.URL.Query().Get("id")
		if id == "" {
			w.Write(errorResponse("id не указан"))
			return
		}

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			w.Write(errorResponse("Ошибка при открытии базы данных"))
			return
		}
		defer db.Close()

		row := db.QueryRow("SELECT id FROM scheduler WHERE id=?", id)

		var taskId string

		err = row.Scan(&taskId)
		if err != nil {
			w.Write(errorResponse("Задача не найдена"))
			return
		}

		stmt, err := db.Prepare("DELETE FROM scheduler WHERE id=?")
		if err != nil {
			w.Write(errorResponse("Ошибка при подготовке запроса"))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(id)
		if err != nil {
			w.Write(errorResponse("Ошибка при удалении записи из базы данных"))
			log.Println(err)
			return
		}

		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			w.Write(errorResponse("Ошибка при записи в json"))
			log.Println(err)
			return
		}

		w.Write(jsonResponse)

	case "PUT":
		var task TaskR

		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			w.Write(errorResponse("Ошибка при декодировании запроса"))
			log.Println("Ошибка при декодировании запроса")
			return
		}
		if task.Title == "" {
			w.Write(errorResponse("Не указано название задачи"))
			log.Println("Не указано название задачи")
			return
		}
		if task.Date == "" {
			task.Date = time.Now().Format("20060102")
		}
		dateTime, err := time.Parse("20060102", task.Date)
		if err != nil {
			w.Write(errorResponse("Некорректный формат даты"))
			log.Println("Некорректный формат даты")
			return
		}

		if dateTime.Before(time.Now()) {
			if task.Repeat == "" {
				task.Date = time.Now().Format("20060102")
			} else if task.Repeat != "" {
				task.Date, err = NextDate(time.Now(), time.Now().Format("20060102"), task.Repeat)
				if err != nil {
					w.Write(errorResponse("Некорректный формат повторения"))
					log.Println("Некорректный формат повторения")
					return
				}
			}
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных")
			return
		}
		defer db.Close()
		stmt, err := db.Prepare("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ?  WHERE id = ?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при создании изменений в базе данных")
			return
		}
		_, err = stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat, task.ID)
		if err != nil {
			w.Write(errorResponse("Ошибка при записи изменений в базу данных"))
			log.Println("Ошибка при записи изменений в базу данных")
			return
		}
		row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?", task.ID)

		var taskId, date, title, comment, repeat string

		err = row.Scan(&taskId, &date, &title, &comment, &repeat)

		if taskId == "" {
			w.Write(errorResponse("Задача не найдена"))
			return
		}
		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)

	default:
		http.Error(w, "неизвестный метод", http.StatusMethodNotAllowed)
	}
}

// api/tasks
func showTasksHandler(w http.ResponseWriter, r *http.Request) {
	var dbPath = "./database/scheduler.db"
	var TODO_DBFILE, exists = os.LookupEnv("TODO_DBFILE")
	if exists && TODO_DBFILE != "" {
		dbPath = TODO_DBFILE
	}
	search := r.URL.Query().Get("search")
	if search != "" {
		searchTime, err := time.Parse("02.01.2006", search)
		if err != nil {
			log.Println("Запрос не является датой")
		}
		searchTimeString := searchTime.Format("20060102")
		search = "%" + search + "%"
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
		}
		defer db.Close()

		rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? OR title LIKE ? OR comment LIKE ? LIMIT 30", searchTimeString, search, search)
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
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	} else {
		db, err := sql.Open("sqlite", dbPath)
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
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// api/task/done
func markTaskAsDoneHandler(w http.ResponseWriter, r *http.Request) {
	var dbPath = "./database/scheduler.db"
	var TODO_DBFILE, exists = os.LookupEnv("TODO_DBFILE")
	if exists && TODO_DBFILE != "" {
		dbPath = TODO_DBFILE
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		w.Write(errorResponse("id не указан"))
		return
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при открытии базы данных")
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?")
	if err != nil {
		w.Write(errorResponse("Ошибка при получении id записи в базе данных"))
	}

	var taskId, date, title, comment, repeat string

	err = stmt.QueryRow(id).Scan(&taskId, &date, &title, &comment, &repeat)
	if err != nil {
		w.Write(errorResponse("Задача не найдена"))
		return
	}
	if repeat == "" {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "не задан id", http.StatusBadRequest)
			return
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		stmt, err := db.Prepare("DELETE FROM scheduler WHERE id=?")
		if err != nil {
			w.Write(errorResponse("Ошибка при подготовке запроса"))
			return
		}
		_, err = stmt.Exec(id)
		if err != nil {
			w.Write(errorResponse("Ошибка при удалении записи из базы данных"))
			log.Println(err)
			return
		}
		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			w.Write(errorResponse("Ошибка при записи в json"))
			log.Println(err)
			return
		}
		w.Write(jsonResponse)

	}
	if repeat != "" {
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()
		stmt, err = db.Prepare("UPDATE scheduler SET date = ? WHERE id = ?")
		if err != nil {
			w.Write(errorResponse("Ошибка при подготовке запроса"))
			return
		}
		renewDate, err := NextDate(time.Now(), date, repeat)
		if err != nil {
			w.Write(errorResponse("Ошибка при получении даты для повторения"))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = stmt.Exec(renewDate, id)
		if err != nil {
			w.Write(errorResponse("Ошибка при записи изменений в базу данных"))
			log.Println(err)
			return
		}
		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			w.Write(errorResponse("Ошибка при записи в json"))
			log.Println(err)
			return
		}
		w.Write(jsonResponse)

	}
}
