// Package app is the basic layout of your application
package app

import "net/http"

// App holds any data that need to be persistent bewteen connections
type App struct {
	greeting string
}

// Hello is a example handler
func (a *App) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(a.greeting))
}
