package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/auth"
	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
	"github.com/google/uuid"
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

func (app *JMcCannBackPackApp) pageHeaderFromContext(context context.Context, pageTitle string) (*model.PageHeader, error) {
	userId, ok := auth.UserIDFromContext(context)
	if !ok {
		return nil, errors.New("no user in cotext")
	}
	user, err := app.DB.GetUserById(userId)
	if err != nil {
		return nil, fmt.Errorf("no user found with id: %s. %s", userId, err)
	}

	if user.Role == model.UserRoleOperative {
		userTrade, err := app.DB.GetOperativeTradeByUserID(user.ID.String())
		if err != nil || userTrade == nil {
			return nil, fmt.Errorf("no user found with id: %s. %s", userId, err)
		}

		return &model.PageHeader{
			Title:    pageTitle,
			UserName: user.Name,
			UserRole: *userTrade,
		}, nil

	}

	return &model.PageHeader{
		Title:    pageTitle,
		UserName: user.Name,
		UserRole: "Admin",
	}, nil
}

func parseUUID(s string) (*uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
