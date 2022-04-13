package main

import (
	"fmt"
	"robFoodDD/dd"
	checksignal "robFoodDD/signal"
	"time"
)

func main() {
	timePoints := []string{
		"06:30",
		"08:30",
		"10:30",
		"12:30",
		"14:30",
		"16:30",
		"18:30",
		"20:30",
		"22:30",
	}
	var cstSh, _ = time.LoadLocation("Asia/Shanghai")
	var multiReserveTime []dd.ReserveTime
	today := time.Now().Format("2006-01-02")
	for i := 0; i < len(timePoints)-1; i++ {
		t1s := today + " " + timePoints[i] + ":00"
		t1, _ := time.ParseInLocation("2006-01-02 15:04:05", t1s, cstSh)

		t2s := today + " " + timePoints[i+1] + ":00"
		t2, _ := time.ParseInLocation("2006-01-02 15:04:05", t2s, cstSh)

		if t1.Before(time.Now()) {
			continue
		}

		multiReserveTime = append(multiReserveTime, dd.ReserveTime{
			StartTimestamp: int(t1.Unix()),
			EndTimestamp:   int(t2.Unix()),
			SelectMsg:      "今天 " + timePoints[i] + "-" + timePoints[i+1],
		})
	}
	fmt.Println(checksignal.ToJson(multiReserveTime))
}
