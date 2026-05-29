package internal

import (
	"fmt"
	"net/http"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
)

type SitesListData struct {
	model.PageHeader
	Sites []model.Site
}

type SiteViewData struct {
	model.PageHeader
	Site *model.Site
	Jobs []model.Job
}

type SiteFormData struct {
	model.PageHeader
	*FormFields
}

func (app *JMcCannBackPackApp) SitesList(w http.ResponseWriter, r *http.Request) {
	sites, err := app.DB.GetSites()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SitesList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Sites")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SitesList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := SitesListData{PageHeader: *pageHeader}
	if sites != nil {
		data.Sites = *sites
	}

	app.render(w, "siteslist.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) SiteCreate(w http.ResponseWriter, r *http.Request) {
	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Site")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SiteCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := SiteFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}

	app.render(w, "sitecreate.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) SiteCreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SiteCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Site")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SiteCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := SiteFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}
	data.Values["name"] = r.FormValue("name")
	data.Values["address"] = r.FormValue("address")

	data.ValidateFormFields(FFName | FFAddress)
	if data.HasErrors() {
		app.render(w, "sitecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	siteID, err := app.DB.NewSite(data.Values["name"], data.Values["address"])
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SiteCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/sites/"+siteID.String(), http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) SiteView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	site, err := app.DB.GetSite(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	jobs, err := app.DB.GetJobs()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SiteView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), site.Name)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SiteView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := SiteViewData{
		PageHeader: *pageHeader,
		Site:       site,
	}
	if jobs != nil {
		for _, j := range *jobs {
			if j.SiteID == site.ID {
				data.Jobs = append(data.Jobs, j)
			}
		}
	}

	app.render(w, "siteview.templ.html", http.StatusOK, data)
}
