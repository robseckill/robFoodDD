package signal

import (
	"fmt"
	"github.com/spf13/cast"
	"os"
	"strings"
	"time"
)

func Exist(t time.Time) {
	for {
		if cast.ToInt64(strings.TrimPrefix(time.Now().Format("1504"), "0")) >
			cast.ToInt64(strings.TrimPrefix(t.Format("1504"), "0")) {
			fmt.Printf("已到退出时间 %s，执行退出……\n", t.Format("2006-01-02 15:04"))
			os.Exit(1)
		}
		time.Sleep(time.Second)
	}
}
