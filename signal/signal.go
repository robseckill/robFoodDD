package signal

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func Notify(path string) {
	c := make(chan os.Signal)
	signal.Notify(c, watchSignal()...)
	fmt.Println()
	go func(p string) {
		for {
			sign := <-c
			switch sign {
			default:
				signal.Reset(sign)

				go func() {
					defer func() {
						if e := recover(); e != nil {
							log.Println("err", e)
						}
					}()
				}()

				Del(p)
				os.Exit(0)
			}
		}
	}(path)
}

func watchSignal() (sign []os.Signal) {
	sign = append(sign,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGILL,
		syscall.SIGABRT,
		syscall.SIGFPE,
		syscall.SIGKILL,
		syscall.SIGSEGV,
		//syscall.SIGPIPE,
		syscall.SIGALRM,
		syscall.SIGTERM)
	return
}

func Del(ex string) {
	os.Remove(ex)
}

func GetFileName() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return ex
}
