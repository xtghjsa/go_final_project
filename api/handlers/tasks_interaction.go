package handlers

import (
	"encoding/json"
	"log"
	"main/api/repeat_date"
	"net/http"
	"strconv"
	"time"
)

// TaskInteractionHandler - Взаимодействие с задачами
func (d *TaskManager) TaskInteractionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET": // Получение задачи по id
		d.GetTask(w, r)
	case "POST": // Добавление задачи
		d.AddTask(w, r)
	case "DELETE": // Удаление задачи
		d.DeleteTask(w, r)
	case "PUT": // Обновление параметров задачи
		d.UpdateTask(w, r)
	default:
		http.Error(w, "недоступный метод", http.StatusMethodNotAllowed)
		log.Println("недоступный метод")
	}
}

func (d *TaskManager) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		_, err := w.Write(ErrorResponse("id не указан"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		return
	}
	row := d.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?", id)

	var taskId, date, title, comment, repeat string

	err := row.Scan(&taskId, &date, &title, &comment, &repeat)
	if err != nil {
		_, err = w.Write(ErrorResponse("Задача не найдена"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
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
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка возвращения ответа", err)
	}
}

func (d *TaskManager) AddTask(w http.ResponseWriter, r *http.Request) {

	var task Task

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при декодировании запроса", err)
		return
	}
	if task.Title == "" {
		_, err = w.Write(ErrorResponse("Не указано название задачи"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("Не указано название задачи")
		return
	}
	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}
	dateTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write(ErrorResponse("Некорректный формат даты"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
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
			task.Date, err = repeat_date.NextDate(time.Now(), time.Now().Format("20060102"), task.Repeat)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write(ErrorResponse("Некорректный формат повторения"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Println("Ошибка возвращения ответа", err)
				}
				log.Println("Некорректный формат повторения", err)
				return
			}
		}
	}
	res, err := d.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
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
	_, err = w.Write(jsonId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка возвращения ответа", err)
	}

}

func (d *TaskManager) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		_, err := w.Write(ErrorResponse("id не указан"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("id не указан")
		return
	}
	res, err := d.DB.Exec("DELETE FROM scheduler WHERE id=?", id)
	if err != nil {
		_, err = w.Write(ErrorResponse("Ошибка при удалении записи из базы данных"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("Ошибка при удалении записи из базы данных", err)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		_, err = w.Write(ErrorResponse("Задача не найдена"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("Задача не найдена", err)
		return
	}

	jsonResponse, err := json.Marshal(map[string]string{})
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

func (d *TaskManager) UpdateTask(w http.ResponseWriter, r *http.Request) {

	var task TaskR

	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		_, err = w.Write(ErrorResponse("Ошибка при декодировании запроса"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("Ошибка при декодировании запроса", err)
		return
	}
	if task.Title == "" {
		_, err = w.Write(ErrorResponse("Не указано название задачи"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("Не указано название задачи")
		return
	}
	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}
	dateTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		_, err = w.Write(ErrorResponse("Некорректный формат даты"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("Некорректный формат даты", err)
		return
	}

	if dateTime.Before(time.Now()) {
		if task.Repeat == "" {
			task.Date = time.Now().Format("20060102")
		} else if task.Repeat != "" {
			task.Date, err = repeat_date.NextDate(time.Now(), time.Now().Format("20060102"), task.Repeat)
			if err != nil {
				_, err = w.Write(ErrorResponse("Некорректный формат повторения"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Println("Ошибка возвращения ответа", err)
				}
				log.Println("Некорректный формат повторения", err)
				return
			}
		}
	}

	_, err = d.DB.Exec("UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ?  WHERE id = ?", task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при записи изменений в базу данных", err)
		return
	}
	row := d.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?", task.ID)

	var taskId, date, title, comment, repeat string

	err = row.Scan(&taskId, &date, &title, &comment, &repeat)
	if err != nil {
		_, err = w.Write(ErrorResponse("Задача не найдена"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		log.Println("Задача не найдена", err)
		return
	}

	if taskId == "" {
		_, err = w.Write(ErrorResponse("Задача не найдена"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		return
	}
	jsonResponse, err := json.Marshal(map[string]string{})
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

// MarkTaskAsDoneHandler - Отметка задачи как завершенной с последующим повторением или удалением в зависимости от параметра repeat
func (d *TaskManager) MarkTaskAsDoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		_, err := w.Write(ErrorResponse("id не указан"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка возвращения ответа", err)
		}
		return
	}

	var taskId, date, title, comment, repeat string

	err := d.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?", id).Scan(&taskId, &date, &title, &comment, &repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println("Задача не найдена")
		return
	}
	if repeat == "" {
		_, err = d.DB.Exec("DELETE FROM scheduler WHERE id=?", id)
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
		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Println("Ошибка возвращения ответа", err)
		}

	}
	if repeat != "" {
		renewDate, err := repeat_date.NextDate(time.Now(), date, repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Ошибка при получении даты для повторения", err)
			return
		}
		_, err = d.DB.Exec("UPDATE scheduler SET date = ? WHERE id = ?", renewDate, id)
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
		_, err = w.Write(jsonResponse)
		if err != nil {
			log.Println("Ошибка возвращения ответа", err)
		}

	}
}
