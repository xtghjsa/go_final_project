package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Форма для ответа на ошибку в json
func errorResponse(errorText string) []byte {
	jsonErr, err := json.Marshal(map[string]string{"error": errorText})
	if err != nil {
		log.Println("Ошибка при записи в json", err)
	}
	return jsonErr
}

// Возвращает JWT-токен
func hashPassword(password string) string {
	secret := []byte(password)
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		log.Println("Ошибка при подписании токена", err)
	}
	log.Println(signedToken)
	return signedToken
}

// EnvInit Загружает переменные из файла .env и создает файл .env в корне проекта если он не существует
func EnvInit() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("файл .env не найден", err)
		log.Println("создаю файл .env")
		_, err = os.Create(".env")
		if err != nil {
			log.Println("не удалось создать файл .env", err)
		}
	}
}

// StartServer Запускает сервер
func StartServer() {

	Port := "7540"
	webDir := "./web"
	TODO_PORT := os.Getenv("TODO_PORT")
	if TODO_PORT != "" {
		Port = TODO_PORT
	}
	// При наличии файла .env c заданным значением переменной TODO_PORT сервер запустится на указанном порту если переменная не задана, то будет использован стандартный порт 7540
	fmt.Printf("Сервер запущен на localhost:%s\n", Port)
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/tasks", auth(showTasksHandler))
	http.HandleFunc("/api/task", auth(taskManagerHandler))
	http.HandleFunc("/api/task/done", auth(markTaskAsDoneHandler))
	http.HandleFunc("/api/signin", checkPasswordHandler)
	err := http.ListenAndServe(":"+Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Получение даты для повторения задачи
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParameter := r.URL.Query().Get("now")
	dateParameter := r.URL.Query().Get("date")
	repeatParameter := r.URL.Query().Get("repeat")
	now, err := time.Parse("20060102", nowParameter)
	if err != nil {
		http.Error(w, "некорректный формат now", http.StatusBadRequest)
		log.Println("некорректный формат now", err)
		return
	}
	next, err := NextDate(now, dateParameter, repeatParameter)
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

// api/task - Взаимодействие с задачами
func taskManagerHandler(w http.ResponseWriter, r *http.Request) {
	var dbPath string
	TODO_DBFILE := os.Getenv("TODO_DBFILE")
	if TODO_DBFILE != "" {
		dbPath = TODO_DBFILE
	} else {
		dbPath = "./scheduler.db"
	}
	switch r.Method {
	case "GET": // Получение задачи по id
		id := r.URL.Query().Get("id")
		if id == "" {
			w.Write(errorResponse("id не указан"))
			return
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
			return
		}
		defer db.Close()

		row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?", id)

		var taskId, date, title, comment, repeat string

		err = row.Scan(&taskId, &date, &title, &comment, &repeat)
		if err != nil {
			w.Write(errorResponse("Задача не найдена"))
			log.Println("Задача не найдена", err)
			return
		}
		task := map[string]string{
			"id":      taskId,
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		}
		if taskId == "" {
			http.Error(w, "Задача не найдена", http.StatusBadRequest)
			return
		}
		jsonResponse, err := json.Marshal(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в json", err)
			return
		}
		w.Write(jsonResponse)

	case "POST": // Добавление задачи

		var task Task

		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при декодировании запроса", err)
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
			w.WriteHeader(http.StatusBadRequest)
			w.Write(errorResponse("Некорректный формат даты"))
			log.Println("Некорректный формат даты", err)
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
					w.WriteHeader(http.StatusBadRequest)
					w.Write(errorResponse("Некорректный формат повторения"))
					log.Println("Некорректный формат повторения", err)
					return
				}
			}
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
			return
		}
		defer db.Close()
		stmt, err := db.Prepare("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при создании записи в базе данных", err)
			return
		}
		res, err := stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в базу данных", err)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при получении id записи в базе данных", err)
			return
		}
		idStr := strconv.Itoa(int(id))
		jsonId, err := json.Marshal(map[string]string{"id": idStr})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи id в json", err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(jsonId)

	case "DELETE": // Удаление задачи
		id := r.URL.Query().Get("id")
		if id == "" {
			w.Write(errorResponse("id не указан"))
			log.Println("id не указан")
			return
		}

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
			return
		}
		defer db.Close()

		row := db.QueryRow("SELECT id FROM scheduler WHERE id=?", id)

		var taskId string

		err = row.Scan(&taskId)
		if err != nil {
			w.Write(errorResponse("Задача не найдена"))
			log.Println("Задача не найдена", err)
			return
		}

		stmt, err := db.Prepare("DELETE FROM scheduler WHERE id=?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при подготовке запроса к базе данных", err)
			return
		}
		_, err = stmt.Exec(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при удалении записи из базы данных", err)
			return
		}

		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в json", err)
			return
		}
		w.Write(jsonResponse)

	case "PUT": // Обновление параметров задачи
		var task TaskR

		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			w.Write(errorResponse("Ошибка при декодировании запроса"))
			log.Println("Ошибка при декодировании запроса", err)
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
			log.Println("Некорректный формат даты", err)
			return
		}

		if dateTime.Before(time.Now()) {
			if task.Repeat == "" {
				task.Date = time.Now().Format("20060102")
			} else if task.Repeat != "" {
				task.Date, err = NextDate(time.Now(), time.Now().Format("20060102"), task.Repeat)
				if err != nil {
					w.Write(errorResponse("Некорректный формат повторения"))
					log.Println("Некорректный формат повторения", err)
					return
				}
			}
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
			return
		}
		defer db.Close()
		stmt, err := db.Prepare("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ?  WHERE id = ?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при создании изменений в базе данных", err)
			return
		}
		_, err = stmt.Exec(task.Date, task.Title, task.Comment, task.Repeat, task.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи изменений в базу данных", err)
			return
		}
		row := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?", task.ID)

		var taskId, date, title, comment, repeat string

		err = row.Scan(&taskId, &date, &title, &comment, &repeat)
		if err != nil {
			w.Write(errorResponse("Задача не найдена"))
			log.Println("Задача не найдена", err)
			return
		}

		if taskId == "" {
			w.Write(errorResponse("Задача не найдена"))
			return
		}
		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в json", err)
			return
		}
		w.Write(jsonResponse)
	default:
		http.Error(w, "недоступный метод", http.StatusMethodNotAllowed)
		log.Println("недоступный метод")
	}
}

// api/tasks - Получение списка задач. Поиск задач по дате/заголовку/комментарию
func showTasksHandler(w http.ResponseWriter, r *http.Request) {
	var dbPath string
	TODO_DBFILE := os.Getenv("TODO_DBFILE")
	if TODO_DBFILE != "" {
		dbPath = TODO_DBFILE
	} else {
		dbPath = "./scheduler.db"
	}
	search := r.URL.Query().Get("search")
	if search != "" {
		searchTime, err := time.Parse("02.01.2006", search)
		if err != nil {
			log.Println("Запрос не является датой", err)
		}
		searchTimeString := searchTime.Format("20060102")
		search = "%" + search + "%"
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
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
			log.Println("Ошибка при записи в json", err)
			return
		}

		w.Write(jsonResponse)
	} else {
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
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

		w.Write(jsonResponse)
	}
}

// api/task/done - Отметка задачи как завершенной с последующим повторением или удалением в зависимости от параметра repeat
func markTaskAsDoneHandler(w http.ResponseWriter, r *http.Request) {
	var dbPath string
	TODO_DBFILE := os.Getenv("TODO_DBFILE")
	if TODO_DBFILE != "" {
		dbPath = TODO_DBFILE
	} else {
		dbPath = "./scheduler.db"
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		w.Write(errorResponse("id не указан"))
		log.Println("id не указан")
		return
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при открытии базы данных", err)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Ошибка при подготовке запроса к базе данных", err)
	}

	var taskId, date, title, comment, repeat string

	err = stmt.QueryRow(id).Scan(&taskId, &date, &title, &comment, &repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Задача не найдена", err)
		return
	}
	if repeat == "" {
		id = r.URL.Query().Get("id")
		if id == "" {
			w.Write(errorResponse("не задан id"))
			log.Println("не задан id")
			return
		}
		db, err = sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
			return
		}
		defer db.Close()
		stmt, err = db.Prepare("DELETE FROM scheduler WHERE id=?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при подготовке запроса к базе данных", err)
			return
		}
		_, err = stmt.Exec(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при удалении записи из базы данных", err)
			return
		}
		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println(err)
			return
		}
		w.Write(jsonResponse)

	}
	if repeat != "" {
		db, err = sql.Open("sqlite", dbPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при открытии базы данных", err)
			return
		}
		defer db.Close()
		stmt, err = db.Prepare("UPDATE scheduler SET date = ? WHERE id = ?")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при подготовке запроса к базе данных", err)
			return
		}
		renewDate, err := NextDate(time.Now(), date, repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при получении даты для повторения", err)
			return
		}
		_, err = stmt.Exec(renewDate, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи изменений в базу данных", err)
			return
		}
		jsonResponse, err := json.Marshal(map[string]string{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в json", err)
			return
		}
		w.Write(jsonResponse)

	}
}

// api/signin - Проверка пароля, аутентификация
func checkPasswordHandler(w http.ResponseWriter, r *http.Request) {
	type pass struct {
		Password string `json:"password"`
	}
	var requested pass

	err := json.NewDecoder(r.Body).Decode(&requested)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при декодировании запроса", err)
		return
	}
	if requested.Password == "" {
		w.Write(errorResponse("Введите пароль"))
		return
	}
	savedPassword := os.Getenv("TODO_PASSWORD")
	if savedPassword == "" {
		w.Write(errorResponse("Пароль не задан"))
		return
	}
	if savedPassword == requested.Password {
		tokenResponse, err := json.Marshal(map[string]string{"token": hashPassword(requested.Password)})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при записи в json", err)
		}
		w.Write(tokenResponse)
	} else {
		w.Write(errorResponse("Неверный пароль"))

	}

}

// Проверка авторизации для всех основных запросов
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwt string
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}

			var valid bool
			valid = validateToken(jwt)
			if !valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}

// Проверка токена
func validateToken(token string) bool {
	pass := os.Getenv("TODO_PASSWORD")
	secret := []byte(pass)
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		log.Println("Ошибка при декодировании токена", err)
		return false
	}
	if jwtToken.Valid {
		log.Println("Токен валиден, доступ разрешен")
		return true
	} else {
		log.Println("Токен невалиден, доступ запрещен")
		return false
	}
}
