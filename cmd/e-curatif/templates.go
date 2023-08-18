package main

import (
        "html/template"
        "path/filepath"
        "time"

        "e-curatif/internal/data"
)

// templateData type acts as the holding structure for any dynamic data that
// will be passed to HTML templates.
type templateData struct {
        Source *data.Source
        Sources []*data.Source

        Info *data.Info
        Infos []*data.Info

        Form any
}

// @ tables source and info, columns "Created" and "Updated" have
// timestamp(UTC)
// SELECT NOW()::timestamp;
// 2023-02-10 19:28:53.116296
// |________| needed
func humanDate(t time.Time) string {
	return t.Format("02/01/2006")
}

// template.FuncMap() is initializa and stocked in a global variable. It
// facilitates the use humanDate function.
var functions = template.FuncMap{
        "humanDate": humanDate,
}

// newTemplateCache() uses filepath.Glob() function to get a slice of all
// filepaths that match the path string.
// Exemple: [./ui/html/base.tmpl.html ./ui/html/pages/index.tmpl.html ...]
func newTemplateCache() (map[string]*template.Template, error) {
        cache := map[string]*template.Template{}

        pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
        if err != nil {
                return nil, err
        }

        for _, page := range pages {
                name := filepath.Base(page)

                // Hardcoded path string below are added maually simply because
                // they aren't inside ".../html/pages/*.tmpl.html"
                files := []string{
                        "./ui/html/base.tmpl.html", 
                        "./ui/html/partials/nav.tmpl.html",
                        page,
                }

                // Parse the files into a template set.
                ts, err := template.ParseFiles(files...)
                if err != nil {
                        return nil, err
                }

                // Template set added to the map, using the name of the page
                // like ('home.tmpl.html') as the key.
                cache[name] = ts
        }

        return cache, nil
}
