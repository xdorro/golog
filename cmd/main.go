package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golog"
)

func main() {
	golog.SetFileOutPutLog("logs.log")

	hook, err := golog.NewAMQPHook()
	if err != nil {
		log.Fatalln(err)
	}
	golog.AddHook(hook)

	go func() {
		for {
			golog.Infof("Hello %s", time.Now())
			time.Sleep(1 * time.Second)
		}
	}()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)
	<-signChan
	log.Println("Shutting down")
	_ = hook.Close()
}
