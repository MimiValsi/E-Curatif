package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
        "os"
        "io"

	"e-curatif/internal/data"
	"e-curatif/internal/validator"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Calls openDB function via application struct, acquires a communication with
// PSQL. If it can't have a connection with the DB then log an error.
// TODO: Add web error page with a 500 status code
func (app *application) dbConn(ctx context.Context) *pgxpool.Conn {
        conn, err := app.DB.Acquire(ctx)
        if err != nil {
                app.errorLog.Fatalln("Couldn't connect to DB")
                return nil
        }

        return conn
}

// ######### 
// Home page
// #########

// Retrieve a slice of all active Info related to their Source
func (app *application) home(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        s, err := app.source.GetAllActive(conn)
        if err != nil {
                app.serverError(w, err)
                return
        }

        // newTemplateData helper is called to get a templateData struct
        // containing the required data.
        data := app.newTemplateData(r)
        data.Sources = s

        // Pass the data to the render() helper so it can be displayed
        app.render(w, http.StatusOK, "home.tmpl.html", data)
}

// ###############
// Source Handlers
// ###############

// Struct that represent the form data and validator.
// It capitalizad so it can be exported and be read by html/template package
// when rendering the template.
type sourceCreateForm struct {
        Name string

        validator.Validator
}
// sourceView() handler checks in the URL string the parameter "id", converts it
// to a integer and check if exists. If yes then fetch the data to be displayed.
func (app *application) sourceView(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        key := chi.URLParam(r, "id")
        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        src, err := app.source.Data(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w, err)
                }

                return
        }

        infos, err := app.info.List(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w, err)
                }
                
                return
        }

        data := app.newTemplateData(r)
        data.Source = src
        data.Infos = infos

        app.render(w, http.StatusOK, "sourceView.tmpl.html", data)
}

// Generate the sourceCreate page to the user, once field filled and submitted,
// a POST form is sent by sourceCreatePost handler. It checks the URL ("name")
// and attempt to send it to the DB.
// If succeeds then it redirects the user to sourceView page.
func (app *application) sourceCreate(w http.ResponseWriter, r *http.Request) {
        data := app.newTemplateData(r)
        data.Form = sourceCreateForm{}

        app.render(w, http.StatusOK, "sourceCreate.tmpl.html", data)
}

func (app *application) sourceCreatePost(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        err := r.ParseForm()
        if err != nil {
                app.clientError(w, http.StatusBadRequest)
        }

        form := sourceCreateForm{
                Name: r.PostForm.Get("name"),
        }

        emptyField := "Ce champ ne doit pas être vide"

        form.CheckField(validator.NotBlank(form.Name), 
                "name", emptyField)

        if !form.Valid() {
                data := app.newTemplateData(r)
                data.Form = form
                app.render(w, http.StatusUnprocessableEntity,
                        "sourceCreate.tmpl.html", data)
                return
        }

        id, err := app.source.Insert(form.Name, conn)
        if err != nil {
                app.serverError(w, err)
                return
        }

        http.Redirect(w, r, fmt.Sprintf("/source/view/%d", id), 
                http.StatusSeeOther)
}

// To delete a Source, only a POST form is necessary.
// Same idea as sourceView, but deletes the source choosen.
func (app *application) sourceDeletePost(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        key := chi.URLParam(r, "id")

        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        err = app.source.Delete(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w, err)
                }

                return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
}

// To update a Source and to simplify the user, first we fetch the data from the
// Source id, display it in the field before beeing modified.
func (app *application) sourceUpdate(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        key := chi.URLParam(r, "id")
        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        src, err := app.source.Data(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w, err)
                }

                return
        }

        data := app.newTemplateData(r)
        data.Source = src

        app.render(w, http.StatusOK, "sourceUpdate.tmpl.html", data)
}

func (app *application) sourceUpdatePost(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        err := r.ParseForm()
        if err != nil {
                app.clientError(w, http.StatusBadRequest)
        }

        key := chi.URLParam(r, "id")
        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        form := sourceCreateForm{
                Name: r.PostForm.Get("name"),
        }

        emptyField := "Ce champ ne doit pas être vide"

        form.CheckField(validator.NotBlank(form.Name), "name", emptyField)

        if !form.Valid() {
                data := app.newTemplateData(r)
                data.Form = form
                
                app.render(w, http.StatusUnprocessableEntity, 
                        "sourceUpdate.tmpl.html", data)
                return
        }

        app.source.Name = form.Name

        err = app.source.Update(id, conn)
        if err != nil {
                app.serverError(w, err)
                return
        }

        http.Redirect(w, r, fmt.Sprintf("/source/view/%d", id), 
                http.StatusSeeOther)
}


// #############
// Info handlers
// #############

// Same thing as sourceCreateForm struct.
type infoCreateForm struct {
	ID       int
	Agent    string
	Material string
	Priority string
	Target   string
	Detail   string
	Created  string
	Updated  string
	Status   string
	Event    string
	Rte      string
	Estimate string
	Brips    string
	Ais      string
	Oups     string
	Ameps    string
	Doneby   string

	validator.Validator
}

func (app *application) infoCreate(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        key := chi.URLParam(r, "id")
        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        src, err := app.source.Data(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w, err)
                }

                return
        }

        data := app.newTemplateData(r)
        data.Form = infoCreateForm{}
        data.Source = src

        app.render(w, http.StatusOK, "infoCreate.tmpl.html", data)
}

// Starts connection with DB, read URL and fetch for the Source id.
// Issue a request form with r.ParseForm function and retrieves every data in
// the fields prompt by the user.
// Some data are priority which are controled with "emptyField" helper, after
// that send it DB.
func (app *application) infoCreatePost(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        err := r.ParseForm()
        if err != nil {
                app.clientError(w, http.StatusBadRequest)
        }

        key := chi.URLParam(r, "id")
        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

	form := infoCreateForm{
		Agent:    r.PostForm.Get("agent"),
		Material: r.PostForm.Get("material"),
		Detail:   r.PostForm.Get("detail"),
		Event:    r.PostForm.Get("event"),
		Priority: r.PostForm.Get("priority"),
		Oups:     r.PostForm.Get("oups"),
		Ameps:    r.PostForm.Get("ameps"),
		Brips:    r.PostForm.Get("brips"),
		Rte:      r.PostForm.Get("rte"),
		Ais:      r.PostForm.Get("ais"),
		Estimate: r.PostForm.Get("estimate"),
		Target:   r.PostForm.Get("target"),
		Status:   r.PostForm.Get("status"),
		Doneby:   r.PostForm.Get("doneby"),
	}
        
        emptyField := "Ce champ ne doit pas être vide"

        form.CheckField(validator.NotBlank(form.Agent),
                "agent", emptyField)

        form.CheckField(validator.NotBlank(form.Material),
                "material", emptyField)

        form.CheckField(validator.NotBlank(form.Detail),
                "detail", emptyField)

        form.CheckField(validator.NotBlank(form.Event),
                "event", emptyField)

        form.CheckField(validator.NotBlank(form.Priority),
                "priority",emptyField)

        form.CheckField(validator.NotBlank(form.Status),
                "status", emptyField)

        if !form.Valid() {
                data := app.newTemplateData(r)
                data.Form = form
                app.render(w, http.StatusUnprocessableEntity, 
                        "infoCreate.tmpl.html", data)
                return
        }

	app.info.Agent = form.Agent
	app.info.Material = form.Material
	app.info.Detail = form.Detail
	app.info.Event = form.Event
	app.info.Oups = form.Oups
	app.info.Ameps = form.Ameps
	app.info.Brips = form.Brips
	app.info.Rte = form.Rte
	app.info.Ais = form.Ais
	app.info.Estimate = form.Estimate
	app.info.Target = form.Target
	app.info.Status = form.Status
	app.info.Doneby = form.Doneby
	app.info.Priority, err = strconv.Atoi(form.Priority)
	if err != nil {
		app.notFound(w)
		return
	}

        _, err = app.info.Insert(id, conn)
        if err != nil {
                app.serverError(w, err)
                return
        }

        http.Redirect(w, r, fmt.Sprintf("/source/%d/info/create", id), 
                http.StatusSeeOther)
}

// Creates DB connection, read the URL and fetch "id", retrieves data from DB
// with the specified "id".
func (app *application) infoView(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        key := chi.URLParam(r, "id")
        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        info, err := app.info.Data(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w,err)
                }
        
                return
        }

        data := app.newTemplateData(r)
        data.Info = info

        app. render(w, http.StatusOK, "infoView.tmpl.html", data)
}

// Same thing as sourceDeletePost but for a info.
func (app *application) infoDeletePost(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        sKey := chi.URLParam(r, "sid")
        iKey := chi.URLParam(r, "id")

        id, err := strconv.Atoi(iKey)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }       
        
        sID, err := strconv.Atoi(sKey)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        err = app.info.Delete(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w, err)
                }

                return
        }

        http.Redirect(w, r, fmt.Sprintf("/source/view/%d", sID), 
                http.StatusSeeOther)
}

// Same thing as sourceUpdate.
// Creates connexion to DB and fetch the data that already exists so it can be
// displayed for the user before beeing updated.
func (app *application) infoUpdate(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        key := chi.URLParam(r, "id")
        id, err := strconv.Atoi(key)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

        info, err := app.info.Data(id, conn)
        if err != nil {
                if errors.Is(err, data.ErrNoRows) {
                        app.notFound(w)
                } else {
                        app.serverError(w, err)
                }

                return
        }

        data := app.newTemplateData(r)
        data.Info = info

        app.render(w, http.StatusOK, "infoUpdate.tmpl.html", data)
}

// same thing as sourceUpdatePost.
func (app *application) infoUpdatePost(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        err := r.ParseForm()
        if err != nil {
                app.clientError(w, http.StatusBadRequest)
        }

        sKey := chi.URLParam(r, "sid")
        sID, err := strconv.Atoi(sKey)
        if err != nil || sID < 1 {
                app.notFound(w)
                return
        }

        iKey := chi.URLParam(r, "id")
        id, err := strconv.Atoi(iKey)
        if err != nil || id < 1 {
                app.notFound(w)
                return
        }

	form := infoCreateForm{
		Agent:    r.PostForm.Get("agent"),
		Material: r.PostForm.Get("material"),
		Detail:   r.PostForm.Get("detail"),
		Event:    r.PostForm.Get("event"),
		Priority: r.PostForm.Get("priority"),
		Oups:     r.PostForm.Get("oups"),
		Ameps:    r.PostForm.Get("ameps"),
		Brips:    r.PostForm.Get("brips"),
		Rte:      r.PostForm.Get("rte"),
		Ais:      r.PostForm.Get("ais"),
		Estimate: r.PostForm.Get("estimate"),
		Target:   r.PostForm.Get("target"),
		Status:   r.PostForm.Get("status"),
		Doneby:   r.PostForm.Get("doneby"),
	}

	app.info.Agent = form.Agent
	app.info.Material = form.Material
	app.info.Detail = form.Detail
	app.info.Event = form.Event
	app.info.Oups = form.Oups
	app.info.Ameps = form.Ameps
	app.info.Brips = form.Brips
	app.info.Rte = form.Rte
	app.info.Ais = form.Ais
	app.info.Estimate = form.Estimate
	app.info.Target = form.Target
	app.info.Status = form.Status
	app.info.Doneby = form.Doneby
	app.info.Priority, err = strconv.Atoi(form.Priority)
	if err != nil {
		app.notFound(w)
		return
	}

        err = app.info.Update(id, conn)
        if err != nil {
                app.serverError(w, err)
                return
        }

        http.Redirect(w, r, fmt.Sprintf("/source/%d/info/view/%d", sID, id),
                http.StatusSeeOther)
}

func (app *application) importCSV(w http.ResponseWriter, r *http.Request) {
        data := app.newTemplateData(r)
        app.render(w, http.StatusOK, "importCSV.tmpl.html", data)
}

func (app *application) importCSVPost(w http.ResponseWriter, r *http.Request) {
        conn := app.dbConn(r.Context())
        defer conn.Release()

        // Max size: 1MB
        r.ParseMultipartForm(1_000_000)

        file, handler, err := r.FormFile("inpt")
        if err != nil {
                app.errorLog.Println("Error retrieving the file")
                app.errorLog.Println(err)
                return
        }

        defer file.Close()

        app.infoLog.Printf("Uploaded File: %+v\n", handler.Filename)
        app.infoLog.Printf("File size: %+v\n", handler.Size)
        app.infoLog.Printf("MIME Header: %+v\n", handler.Header)

	dst, err := os.Create("csvFiles/" + handler.Filename)
	if err != nil {
		app.errorLog.Println(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
	defer dst.Close()

        // Copie the file and transfert it to the system.
	if _, err := io.Copy(dst, file); err != nil {
		app.errorLog.Println(w, err.Error(),
			http.StatusInternalServerError)
		return
	}

        // Run the extension verification et file encoding, if it's valid, the
        // data will be transfert to DB.
	app.csv.Verify("csvFiles/" + handler.Filename)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
