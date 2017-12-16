package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/derailed/hangman2/internal/cors"
	"github.com/derailed/hangman2/internal/game"
	"github.com/go-kit/kit/log"
)

const port = ":9095"

const localDic = "http://localhost:9094"

func main() {
	dicUrl := flag.String("url", localDic, "Dictionary Host URL")

	flag.Parse()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	svc := game.NewService(*dicUrl)
	svc = game.NewLoggingService(svc, logger)

	mux := http.NewServeMux()
	mux.Handle("/game/v1/", game.MakeHandler(svc, logger))
	http.Handle("/", cors.AccessControl(mux))

	errs := make(chan error, 2)
	go func() {
		logger.Log("msg", "HTTP", "addr", port)
		errs <- http.ListenAndServe(port, nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}
