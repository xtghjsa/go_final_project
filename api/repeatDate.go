package api

import (
	"errors"
	"log"
	"sort"
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
		return "", errors.New("день в дате не может быть больше 31 или меньше 1")
	}

	// Проверка на корректность формата задаваемого повтора
	rules := strings.Split(repeat, " ")
	if rules[0] != "d" && rules[0] != "w" && rules[0] != "m" && rules[0] != "y" {
		return "", errors.New("неправильный формат повтора")
	}
	switch rules[0] {

	case "d": // дни
		// Проверка на наличие выбранных дней
		if len(rules) == 1 {
			return "", errors.New("количество выбранных дней не может быть пустым")
		}
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

	case "w": // дни недели
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

	case "m": // дни месяца
		if len(rules) > 3 || len(rules) == 1 {
			return "", errors.New("неверный формат повтора дней месяца")
		}
		chosenMonthsDaysStr := strings.Split(rules[1], ",")
		chosenMonthsDaysInt := make([]int, 0, 7)
		monthsMap := map[int]time.Month{1: time.January, 2: time.February, 3: time.March, 4: time.April, 5: time.May, 6: time.June, 7: time.July, 8: time.August, 9: time.September, 10: time.October, 11: time.November, 12: time.December}
		monthDayMax := 31
		monthDayMin := -2
		for _, v := range chosenMonthsDaysStr {
			monthDay, err := strconv.Atoi(v)
			if err != nil {
				return "", err
			}
			if monthDay > monthDayMax || monthDay < monthDayMin {
				return "", errors.New("выбранный день месяца не может быть больше 31 или меньше -2")
			}
			chosenMonthsDaysInt = append(chosenMonthsDaysInt, monthDay)
		}
		nowUnix := now.Unix()
		var day int64 = 86400
		dateUnix := parsedDate.Unix()
		var results []int64
		if len(rules) == 2 {
			if now.After(parsedDate) {
				for i := 0; i < 999; i++ {
					nowUnix += day
					for _, v := range chosenMonthsDaysInt {
						caseMResult := time.Unix(nowUnix, 0)
						if caseMResult.Day() == v && v > dayMin {
							results = append(results, caseMResult.Unix())
							return caseMResult.Format("20060102"), nil
						}
						if v < 0 {
							var preLastDay int64
							var lastDay int64
							preLastDay = time.Date(caseMResult.Year(), caseMResult.Month()+1, -1, 0, 0, 0, 0, time.UTC).Unix()
							lastDay = time.Date(caseMResult.Year(), caseMResult.Month()+1, 0, 0, 0, 0, 0, time.UTC).Unix()
							if v == -1 {
								results = append(results, lastDay)

							}
							if v == -2 {
								results = append(results, preLastDay)

							}

						}

					}
				}
			}
			for i := 0; i < 999; i++ {
				dateUnix += day
				for _, v := range chosenMonthsDaysInt {
					caseMResult := time.Unix(dateUnix, 0)
					if caseMResult.Day() == v && v > dayMin {
						results = append(results, caseMResult.Unix())
						return caseMResult.Format("20060102"), nil

					}
					if v < 0 {
						var preLastDay int64
						var lastDay int64
						preLastDay = time.Date(caseMResult.Year(), caseMResult.Month()+1, -1, 0, 0, 0, 0, time.UTC).Unix()
						lastDay = time.Date(caseMResult.Year(), caseMResult.Month()+1, 0, 0, 0, 0, 0, time.UTC).Unix()
						if v == -1 {
							results = append(results, lastDay)

						}
						if v == -2 {
							results = append(results, preLastDay)

						}

					}

				}
			}
			sort.Slice(results, func(i, j int) bool {
				return results[i] < results[j]
			})
			return time.Unix(results[0], 0).Format("20060102"), nil

		}

		if len(rules) > 2 {
			chosenMonthsStr := strings.Split(rules[2], ",")
			chosenMonthsInt := make([]int, 0, 7)
			for _, v := range chosenMonthsStr {
				month, err := strconv.Atoi(v)
				if err != nil {
					return "", err
				}
				if month > 12 || month < 1 {
					return "", errors.New("выбранный месяц не может быть больше 12 или меньше 1")
				}
				chosenMonthsInt = append(chosenMonthsInt, month)
			}
			for i := 0; i < 999; i++ {
				nowUnix += day
				for _, v := range chosenMonthsInt {
					for _, w := range chosenMonthsDaysInt {
						caseMResult := time.Unix(nowUnix, 0)
						if caseMResult.Day() == w && caseMResult.Month() == monthsMap[v] {
							return caseMResult.Format("20060102"), nil
						}

					}
				}
			}
		}

	case "y": // год
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
