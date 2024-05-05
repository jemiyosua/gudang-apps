package helper

import (
	"strconv"
	"time"
)

func GetDate(Format string) string {
	start := time.Now()

	Second := start.Second()
	SecondState := ""
	if Second < 10 {
		SecondState = "less"
	} else {
		SecondState = "more"
	}
	Minute := start.Minute()
	MinuteState := ""
	if Minute < 10 {
		MinuteState = "less"
	} else {
		MinuteState = "more"
	}
	Hour := start.Hour()
	HourState := ""
	if Hour < 10 {
		HourState = "less"
	} else {
		HourState = "more"
	}

	Day := start.Day()
	DayState := ""
	if Day < 10 {
		DayState = "less"
	} else {
		DayState = "more"
	}
	Month := start.Month()
	Year := start.Year()

	SecondString := strconv.Itoa(Second)
	vSecondString := ""
	if SecondState == "less" {
		vSecondString = "0" + SecondString
	} else {
		vSecondString = SecondString
	}
	MinuteString := strconv.Itoa(Minute)
	vMinuteString := ""
	if MinuteState == "less" {
		vMinuteString = "0" + MinuteString
	} else {
		vMinuteString = MinuteString
	}

	HourString := strconv.Itoa(Hour)
	vHourString := ""
	if HourState == "less" {
		vHourString = "0" + HourString
	} else {
		vHourString = HourString
	}

	DayString := strconv.Itoa(Day)
	vDayString := ""
	if DayState == "less" {
		vDayString = "0" + DayString
	} else {
		vDayString = DayString
	}
	MonthString := GetMonthInt(Month.String())
	YearString := strconv.Itoa(Year)
	Year2Digit := YearString[2:4]

	Date := ""
	if Format == "Ymd" {
		Date = YearString + MonthString + vDayString
	} else if Format == "dmYhis" {
		Date = vDayString + MonthString + YearString + vHourString + vMinuteString + vSecondString
	} else if Format == "m/d/Y" {
		Date = MonthString + "/" + DayString + "/" + YearString
	} else if Format == "d/m/Y" {
		Date = DayString + "/" + MonthString + "/" + YearString
	} else if Format == "d-m-Y h:i:s" {
		Date = vDayString + "-" + MonthString + "-" + YearString + " " + vHourString + ":" + vMinuteString + ":" + vSecondString
	} else if Format == "ymdhis" {
		Date = Year2Digit + MonthString + vDayString + vHourString + vMinuteString + vSecondString
	} else if Format == "Ymdhis" {
		Date = YearString + MonthString + vDayString + vHourString + vMinuteString + vSecondString
	} else if Format == "Y-m-d H:i:s" {
		Date = YearString + "-" + MonthString + "-" + vDayString + " " + vHourString + ":" + vMinuteString + "-" + vSecondString
	} else if Format == "m" {
		Date = MonthString
	} else if Format == "Y-m-d" {
		Date = YearString + "-" + MonthString + "-" + vDayString
	} else if Format == "h:i:s" {
		Date = vHourString + ":" + vMinuteString + ":" + vSecondString
	} else if Format == "d/m/Y h:i:s" {
		Date = vDayString + "/" + MonthString + "/" + YearString + " " + vHourString + ":" + vMinuteString + ":" + vSecondString
	} else if Format == "d-m-Y" {
		Date = vDayString + "-" + MonthString + "-" + YearString
	}

	return Date
}

func GetMonthInt(Month string) string {
	MonthInt := ""
	if Month == "January" {
		MonthInt = "01"
	} else if Month == "February" {
		MonthInt = "02"
	} else if Month == "March" {
		MonthInt = "03"
	} else if Month == "April" {
		MonthInt = "04"
	} else if Month == "May" {
		MonthInt = "05"
	} else if Month == "June" {
		MonthInt = "06"
	} else if Month == "July" {
		MonthInt = "07"
	} else if Month == "August" {
		MonthInt = "08"
	} else if Month == "September" {
		MonthInt = "09"
	} else if Month == "October" {
		MonthInt = "10"
	} else if Month == "November" {
		MonthInt = "11"
	} else if Month == "December" {
		MonthInt = "12"
	}

	return MonthInt
}

func GetMonthString(MonthInt string) (string, string) {
	MonthStringId := ""
	MonthStringEn := ""

	if MonthInt == "01" {
		MonthStringId = "Januari"
		MonthStringEn = "January"
	} else if MonthInt == "02" {
		MonthStringId = "Februari"
		MonthStringEn = "February"
	} else if MonthInt == "03" {
		MonthStringId = "Maret"
		MonthStringEn = "March"
	} else if MonthInt == "04" {
		MonthStringId = "April"
		MonthStringEn = "Apr"
	} else if MonthInt == "05" {
		MonthStringId = "Mei"
		MonthStringEn = "May"
	} else if MonthInt == "06" {
		MonthStringId = "Juni"
		MonthStringEn = "June"
	} else if MonthInt == "07" {
		MonthStringId = "Juli"
		MonthStringEn = "July"
	} else if MonthInt == "08" {
		MonthStringId = "Agustus"
		MonthStringEn = "August"
	} else if MonthInt == "09" {
		MonthStringId = "September"
		MonthStringEn = "September"
	} else if MonthInt == "10" {
		MonthStringId = "Oktober"
		MonthStringEn = "October"
	} else if MonthInt == "11" {
		MonthStringId = "November"
		MonthStringEn = "November"
	} else if MonthInt == "12" {
		MonthStringId = "Desember"
		MonthStringEn = "December"
	}

	return MonthStringId, MonthStringEn
}