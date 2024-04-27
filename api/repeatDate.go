package api

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Проверка на наличие даты повторения
	if repeat == "" {
		return "", errors.New("повтор не может быть пустым")
	}
	// Проверка на корректность формата даты
	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		log.Println("неверный формат даты", err)
	}
	// Проверка на корректность года в дате
	var yearCheckString string
	for k, v := range date {
		if k == 0 || k == 1 || k == 2 || k == 3 {
			yearCheckString += string(v)
		}
	}
	yearCheckInt, err := strconv.Atoi(yearCheckString)
	if err != nil {
		return "", err
	}
	yearMin := 0001
	if yearCheckInt < yearMin {
		return "", errors.New("год в дате не может быть меньше 1")
	}
	// Проверка на корректность месяца в дате
	var monthCheckString string
	for k, v := range date {
		if k == 4 || k == 5 {
			monthCheckString += string(v)
		}
	}
	monthCheckInt, err := strconv.Atoi(monthCheckString)
	if err != nil {
		return "", err
	}
	monthMax := 12
	monthMin := 01
	if monthCheckInt > monthMax || monthCheckInt < monthMin {
		return "", errors.New("месяц в дате не может быть больше 12 или меньше 1")
	}
	// Проверка на корректность дня в дате
	var dayCheckString string
	for k, v := range date {
		if k == 6 || k == 7 {
			dayCheckString += string(v)
		}
	}
	dayCheckInt, err := strconv.Atoi(dayCheckString)
	if err != nil {
		return "", err
	}
	dayMax := 31
	dayMin := 01
	if dayCheckInt > dayMax || dayCheckInt < dayMin {
		return "", errors.New("дни в дате не может быть больше 31 или меньше 1")
	}

	rules := strings.Split(repeat, " ")
	// Проверка на корректность формата задаваемого повтора
	if rules[0] != "d" && rules[0] != "w" && rules[0] != "m" && rules[0] != "y" {
		return "", errors.New("неправильный формат повтора")
	}
	// Проверка на наличие выбранных дней
	if rules[0] == "d" && len(rules) == 1 {
		return "", errors.New("количество выбранных дней не может быть пустым")
	}
	switch rules[0] {
	// дни
	case "d":
		days, err := strconv.Atoi(rules[1])
		if err != nil {
			return "", err
		}
		if days > 400 || days < 1 {
			return "", errors.New("дней не может быть больше 400 и меньше 1")
		}
		var caseDResult string
		if parsedDate.After(now) {
			caseDResult = parsedDate.AddDate(0, 0, days).Format("20060102")
		}
		for i := parsedDate; i.Before(now); i = i.AddDate(0, 0, days) {
			caseDResult = i.AddDate(0, 0, days).Format("20060102")
		}
		return caseDResult, nil
		// дни недели
	case "w":
		// Проверка на наличие выбранных дней недели
		if len(rules) == 1 {
			return "", errors.New("количество выбранных дней недели не может быть пустым")
		}

		chosenWeekdaysStr := strings.Split(rules[1], ",")
		chosenWeekdaysInt := make([]int, 0, 7)
		weekdaysMap := map[int]time.Weekday{1: time.Monday, 2: time.Tuesday, 3: time.Wednesday, 4: time.Thursday, 5: time.Friday, 6: time.Saturday, 7: time.Sunday}
		for _, v := range chosenWeekdaysStr {
			weekday, err := strconv.Atoi(v)
			if err != nil {
				return "", err
			}
			if weekday > 7 || weekday < 1 {
				return "", errors.New("день недели не может быть больше 7 или меньше 1")
			}
			chosenWeekdaysInt = append(chosenWeekdaysInt, weekday)
		}

		nowUnix := now.Unix()
		var day int64 = 86400
		for i := 0; i < 999; i++ {
			nowUnix += day
			for _, v := range chosenWeekdaysInt {
				caseWResult := time.Unix(nowUnix, 0)
				if caseWResult.Weekday() == weekdaysMap[v] {
					return caseWResult.Format("20060102"), nil
				}
			}
		}
		return "", err

		// дни месяца
	case "m":
		return "", errors.New("not implemented, WIP")

		// год
	case "y":
		var caseYResult string
		if parsedDate.After(now) {
			caseYResult = parsedDate.AddDate(1, 0, 0).Format("20060102")
		}
		for i := parsedDate; i.Before(now); i = i.AddDate(1, 0, 0) {
			caseYResult = i.AddDate(1, 0, 0).Format("20060102")
		}
		return caseYResult, nil
	}
	return "", errors.New("неправильный формат повтора")
}
