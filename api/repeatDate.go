package api

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Check year correctness in date
	if repeat == "" {
		return "", errors.New("repeat can't be empty")
	}
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
		return "", errors.New("year is out of range")
	}
	// Check month correctness in date
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
		return "", errors.New("month is out of range")
	}
	// Check day correctness in date
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
		return "", errors.New("day is out of range")
	}
	// Check date correctness
	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		log.Println("wrong date format", err)
	}
	rules := strings.Split(repeat, " ")
	// Check repeat symbol correctness
	if rules[0] != "d" && rules[0] != "w" && rules[0] != "m" && rules[0] != "y" {
		return "", errors.New("wrong repeat symbol")
	}
	if rules[0] == "d" && len(rules) == 1 {
		return "", errors.New("days can't be empty")
	}
	switch rules[0] {
	// day
	case "d":
		if len(rules) == 1 {
			return "", errors.New("days amount can't be empty")
		}
		days, err := strconv.Atoi(rules[1])
		if err != nil {
			log.Println("wrong days format or days are not specified", err)
		}
		if days > 400 || days <= 0 {
			return "", errors.New("days can't be more than 400 and less than 0")
		}
		var caseDResult string
		if parsedDate.After(now) {
			caseDResult = parsedDate.AddDate(0, 0, days).Format("20060102")
		}
		for i := parsedDate; i.Before(now); i = i.AddDate(0, 0, days) {
			caseDResult = i.AddDate(0, 0, days).Format("20060102")
		}
		return caseDResult, nil
		// week
	case "w":
		return "", errors.New("not implemented, WIP")
		// month
	case "m":
		return "", errors.New("not implemented, WIP")
		// year
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
	return "", errors.New("wrong repeat format")
}
