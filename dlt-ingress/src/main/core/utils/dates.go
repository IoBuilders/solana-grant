package utils

import "time"

const TimeOnlyWithoutSeconds = "15:04"

func SameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func TodayLastMinute(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.UTC)
}
