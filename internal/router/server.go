package router

import (
	"net/http"
	"time"
	"fmt"
)

func NewHTTPServer(app *App, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Config.App.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}