package internal

import (
	"fmt"
	"net/http"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/auth"
	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
)

type QSDashboardData struct {
	model.PageHeader
	Jobs       []model.Job
	Sites      []model.Site
	Operatives []model.Operative
	Resources  []model.Resource
}

type OperativeDashboardData struct {
	model.PageHeader
	Operative *model.Operative
	Jobs      []model.Job
}

func (app *JMcCannBackPackApp) Dashboard(w http.ResponseWriter, r *http.Request) {
	role, ok := auth.UserRoleFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	switch model.UserRole(role) {
	case model.UserRoleQS:
		app.QSDashboard(w, r)
	case model.UserRoleOperative:
		app.OperativeDashboard(w, r)
	default:
		app.clientError(w, http.StatusForbidden)
	}
}

func (app *JMcCannBackPackApp) QSDashboard(w http.ResponseWriter, r *http.Request) {
	jobs, err := app.DB.GetJobs()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "QSDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	sites, err := app.DB.GetSites()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "QSDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	operatives, err := app.DB.GetOperatives()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "QSDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	resources, err := app.DB.GetResources()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "QSDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Jobs")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "QSDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	// pageHeader.BackURL = "/dashboard"

	data := QSDashboardData{
		PageHeader: *pageHeader,
	}
	if jobs != nil {
		data.Jobs = *jobs
	}
	if sites != nil {
		data.Sites = *sites
	}
	if operatives != nil {
		data.Operatives = *operatives
	}
	if resources != nil {
		data.Resources = *resources
	}

	app.render(w, "qsdashboard.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) OperativeDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	operative, err := app.DB.GetOperativeByUserID(userID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrNotFound, err),
		})
		return
	}

	jobs, err := app.DB.GetJobs()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Jobs")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeDashboard",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	// pageHeader.BackURL = "/dashboard"

	data := OperativeDashboardData{
		PageHeader: *pageHeader,
		Operative:  operative,
	}
	if jobs != nil {
		data.Jobs = *jobs
	}

	app.render(w, "operativedashboard.templ.html", http.StatusOK, data)
}
