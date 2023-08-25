package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
)

// serverError() helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error responder to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())

	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError)
}

// clientError() helper sends a specific status code and corresponding
// description to the user.
// Exemple: 400 "Bad Request" when there's a problem with te request that the
// user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound() uses clientError() method to generate an error as response to the
// user. It's a convinience wrapper around clientError
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// render() retrieves the appropriate template set from the cache based on the
// page name (exemple: 'home.tmpl.html'). If no entry exists in the cache with
// the provided name, then create a new error and call the serverError() helper
// method.
func (app *application) render(w http.ResponseWriter, status int, page string,
	data *templateData) {

	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

// newTemplateData() returns a pointer to templateData struct already
// initialized.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{}
}
