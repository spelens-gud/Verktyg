package main

import (
	"net/http"
	"time"

	"github.com/spelens-gud/Verktyg.git/kits/kserver"
)

func main() {
	kserver.Run(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(time.Second)
		writer.WriteHeader(200)
	}), kserver.Config{
		Addr:                     ":8888",
		ReadTimeoutSeconds:       0,
		ReadHeaderTimeoutSeconds: 0,
		WriteTimeoutSeconds:      0,
		IdleTimeoutSeconds:       0,
		CloseWaitSeconds:         0,
		GracefulRestart:          true,
	})
}
