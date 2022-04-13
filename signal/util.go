package signal

import (
	"encoding/json"
	"github.com/spf13/cast"
	"math/rand"
	"time"
)

func ToJson(data interface{}) string {
	s, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(s)
}

func RandSleep(st time.Duration) {
	var a float64
	a = 1.0 + cast.ToFloat64(rand.Intn(10))/100.0
	s := cast.ToInt64(cast.ToFloat64(st.Milliseconds()) * a)
	time.Sleep(time.Duration(s) * time.Millisecond)
}
