package internal

import (
	"fmt"
	"net/http"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
)

func (app *JMcCannBackPackApp) render(w http.ResponseWriter, page string, status int, data any) {
	ts, ok := app.TemplateCache.Cache[page]
	if !ok {
		err := &appError.ServerError{
			Handler: "render",
			Err:     fmt.Errorf("%w: %s", appError.ServerErrTemplateMissing, page),
		}
		app.ErrorLog.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := ts.ExecuteTemplate(w, "base", data); err != nil {
		app.ErrorLog.Println(&appError.ServerError{
			Handler: "render",
			Err:     fmt.Errorf("%w: %s", appError.ServerErrTemplateRender, err),
		})
		return
	}
}

func (app *JMcCannBackPackApp) serverError(w http.ResponseWriter, err error) {
	app.ErrorLog.Println(&appError.ServerError{
		Handler: "serverError",
		Err:     fmt.Errorf("%w: %s", appError.ServerErrInternal, err),
	})
	app.render(w, "error.templ.html", http.StatusInternalServerError, nil)
}

func (app *JMcCannBackPackApp) clientError(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	switch status {
	case http.StatusNotFound:
		app.render(w, "notfound.templ.html", status, nil)
	case http.StatusForbidden:
		app.render(w, "error.templ.html", status, &appError.ServerError{
			Handler: "clientError",
			Err:     appError.ServerErrForbidden,
		})
	case http.StatusUnauthorized:
		app.render(w, "error.templ.html", status, &appError.ServerError{
			Handler: "clientError",
			Err:     appError.ServerErrUnauthorized,
		})
	case http.StatusBadRequest:
		app.render(w, "error.templ.html", status, &appError.ServerError{
			Handler: "clientError",
			Err:     appError.ServerErrBadRequest,
		})
	}
}

type Item interface {
	Id() string
}

func getItem[T Item](key string, item []T) *T {
	for _, t := range item {
		if t.Id() == key {
			return &t
		}
	}
	return nil
}
