package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/apognu/gocal"
)

type StringSlice map[string][]string
type StringStringSlice map[string]StringSlice

type TodayData struct {
	Calendar StringStringSlice
}

type PeopleHrCalendar struct {
	Url           string
	LastRetrieved time.Time
	CacheTimeout  time.Duration
	Data          []byte
}
type PeopleHr struct {
	Calendar *PeopleHrCalendar
}

func NewPeopleHr(calendarUrl string) *PeopleHr {
	return &PeopleHr{
		Calendar: &PeopleHrCalendar{
			Url:          calendarUrl,
			CacheTimeout: time.Hour, // Cache for 1 hour
		},
	}
}

func (obj *PeopleHr) EnsureData() {
	obj.Calendar.EnsureData()
}

func (obj *PeopleHrCalendar) EnsureData() {
	// Errors are logged out result in none of the data being changed, therefore stale data will be used.
	if obj.LastRetrieved.Add(obj.CacheTimeout).Before(time.Now()) {
		resp, err := http.Get(obj.Url)
		if err != nil {
			log.Printf("Error fetching PeopleHr calendar data: %s", err)
		} else {
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading PeopleHr calendar data: %s", err)
			} else {
				obj.Data = data
				obj.LastRetrieved = time.Now()
			}
		}
	}
}

func (obj *PeopleHr) Today() (*TodayData, error) {
	// We want 5 days-worth of calendar entries.
	obj.EnsureData()

	var err error
	res := TodayData{}

	res.Calendar, err = obj.TodayCalendar()
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (obj *PeopleHr) TodayCalendar() (StringStringSlice, error) {
	var err error

	res := make(StringStringSlice, 0)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 59, time.UTC)

	days := 0
	for {
		// Skip weekends.
		dayOfWeek := start.Weekday().String()
		if dayOfWeek != "Saturday" && dayOfWeek != "Sunday" {
			res[start.Format("2006-01-02")], err = obj.TodayCalendarDay(&start, &end)
			days += 1
		}

		if days >= 5 {
			break
		}
		start = start.Add(24 * time.Hour)
		end = end.Add(24 * time.Hour)
	}

	return res, err
}

func (obj *PeopleHr) TodayCalendarDay(start *time.Time, end *time.Time) (StringSlice, error) {
	cal := gocal.NewParser(bytes.NewReader(obj.Calendar.Data))
	cal.Start = start
	cal.End = end
	err := cal.Parse()

	if err != nil {
		return nil, err
	}

	res := make(StringSlice, 0)
	for _, evt := range cal.Events {
		summaryParts := strings.Split(evt.Summary, " - ")
		key := summaryParts[len(summaryParts)-1]
		val := evt.Summary[:len(evt.Summary)-len(key)-len(" - ")]

		if key == "Other Events" {
			descParts := strings.Split(evt.Description, "\\n")
			key = descParts[0]
		}

		if _, ok := res[key]; !ok {
			res[key] = make([]string, 0)
		}
		res[key] = append(res[key], val)
	}

	return res, nil
}
