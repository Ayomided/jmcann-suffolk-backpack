package internal

import (
	"fmt"
	"net/http"
	"time"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/auth"
	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
)

type JobsListData struct {
	model.PageHeader
	Jobs []model.Job
}

type JobViewData struct {
	model.PageHeader
	Job        *model.Job
	Site       *model.Site
	Sessions   []model.JobSession
	Operatives []model.Operative
	IsQS       bool
}

type JobFormData struct {
	model.PageHeader
	*FormFields
	Sites []model.Site
}

type SessionViewData struct {
	model.PageHeader
	Session      *model.JobSession
	Job          *model.Job
	Operatives   []model.JobOperative
	Resources    []model.JobResource
	IsQS         bool
	SessionTotal *model.Money
}

type SessionFormData struct {
	model.PageHeader
	*FormFields
	Jobs []model.Job
}

type JobResourceFormData struct {
	model.PageHeader
	*FormFields
	Session   *model.JobSession
	Resources []model.Resource
}

func (app *JMcCannBackPackApp) JobsList(w http.ResponseWriter, r *http.Request) {
	jobs, err := app.DB.GetJobs()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobsList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Jobs")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobsList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := JobsListData{PageHeader: *pageHeader}
	if jobs != nil {
		data.Jobs = *jobs
	}

	app.render(w, "jobs.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) JobCreate(w http.ResponseWriter, r *http.Request) {
	sites, err := app.DB.GetSites()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Job")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := JobFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}
	if sites != nil {
		data.Sites = *sites
	}

	app.render(w, "jobcreate.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) JobCreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Job")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := JobFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}
	data.Values["reference"] = r.FormValue("reference")
	data.Values["name"] = r.FormValue("name")
	data.Values["site_id"] = r.FormValue("site_id")
	data.Values["start_datetime"] = r.FormValue("start_datetime")
	data.Values["expected_headcount"] = r.FormValue("expected_headcount")

	data.ValidateFormFields(FFReference | FFJobName | FFSiteID | FFStartDatetime | FFHeadcount)
	if data.HasErrors() {
		sites, _ := app.DB.GetSites()
		if sites != nil {
			data.Sites = *sites
		}
		app.render(w, "jobcreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	siteID, err := parseUUID(data.Values["site_id"])
	if err != nil {
		data.Errors["site_id"] = "invalid site"
		app.render(w, "jobcreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	startDatetime, err := time.Parse("2006-01-02", data.Values["start_datetime"])
	if err != nil {
		data.Errors["start_datetime"] = "invalid date format"
		app.render(w, "jobcreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	createdBy, err := parseUUID(userID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrInvalidID, err),
		})
		return
	}

	headcount := uint32(2)
	if data.Values["expected_headcount"] != "" {
		var h int
		if _, err := fmt.Sscanf(data.Values["expected_headcount"], "%d", &h); err == nil && h > 0 {
			headcount = uint32(h)
		}
	}

	status := model.JobStatusActive
	jobID, err := app.DB.NewJob(
		data.Values["reference"],
		data.Values["name"],
		siteID,
		createdBy,
		&status,
		startDatetime,
		&model.JobOperativeRequirement{ExpectedHeadcount: headcount},
		nil,
	)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/jobs/"+jobID.String(), http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) JobView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	job, err := app.DB.GetJob(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	site, err := app.DB.GetSite(&job.SiteID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	sessions, err := app.DB.GetJobSessions(&job.ID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	operatives, err := app.DB.GetOperativesWithRates()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), job.Name)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	role, ok := auth.UserRoleFromContext(r.Context())
	data := JobViewData{
		PageHeader: *pageHeader,
		Job:        job,
		Site:       site,
		IsQS:       ok && model.UserRole(role) == model.UserRoleQS,
	}
	if sessions != nil {
		data.Sessions = *sessions
	}
	if operatives != nil {
		data.Operatives = *operatives
	}

	app.render(w, "jobview.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) JobUpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobUpdateStatus",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	status := model.JobStatus(r.FormValue("status"))
	_, err = app.DB.UpdateJob(id, &status, nil, nil)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobUpdateStatus",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/jobs/"+idStr, http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) JobDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	_, err = app.DB.DeleteJob(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobDelete",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/jobs", http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) JobCalculateCost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	job, err := app.DB.GetJob(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if job.Status != model.JobStatusSubmitted && job.Status != model.JobStatusCompleted {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	sessions, err := app.DB.GetJobSessions(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCalculateCost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	grandTotal := model.NewMoney("GBP", 0)

	if sessions != nil {
		for _, session := range *sessions {
			sessionID := session.ID

			operatives, err := app.DB.GetJobOperatives(&sessionID)
			if err != nil {
				app.serverError(w, &appError.ServerError{
					Handler: "JobCalculateCost",
					Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
				})
				return
			}
			if operatives != nil {
				for _, op := range *operatives {
					cost := op.Calculate()
					if cost != nil {
						result, err := grandTotal.Add(*cost)
						if err != nil {
							app.serverError(w, &appError.ServerError{
								Handler: "JobCalculateCost",
								Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
							})
							return
						}
						grandTotal = result
						app.DB.UpdateJobOperative(&op.ID, op.DepartureTime, cost)
					}
				}
			}

			resources, err := app.DB.GetJobResources(&sessionID)
			if err != nil {
				app.serverError(w, &appError.ServerError{
					Handler: "JobCalculateCost",
					Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
				})
				return
			}
			if resources != nil {
				for _, res := range *resources {
					cost := res.Calculate()
					result, err := grandTotal.Add(cost)
					if err != nil {
						app.serverError(w, &appError.ServerError{
							Handler: "JobCalculateCost",
							Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
						})
						return
					}
					grandTotal = result
					app.DB.UpdateJobResource(&res.ID, &cost)
				}
			}
		}
	}

	status := model.JobStatusApproved
	_, err = app.DB.UpdateJob(id, &status, nil, &grandTotal)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobCalculateCost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/jobs/"+idStr, http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) SessionCreate(w http.ResponseWriter, r *http.Request) {
	jobs, err := app.DB.GetJobs()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Session")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := SessionFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}
	if jobs != nil {
		data.Jobs = *jobs
	}

	app.render(w, "sessioncreate.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) SessionCreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Session")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := SessionFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}
	data.Values["job_id"] = r.FormValue("job_id")
	data.Values["start_time"] = r.FormValue("start_time")
	data.Values["notes"] = r.FormValue("notes")

	data.ValidateFormFields(FFJobID | FFStartTime)
	if data.HasErrors() {
		jobs, _ := app.DB.GetJobs()
		if jobs != nil {
			data.Jobs = *jobs
		}
		app.render(w, "sessioncreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	jobID, err := parseUUID(data.Values["job_id"])
	if err != nil {
		data.Errors["job_id"] = "invalid job"
		app.render(w, "sessioncreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	startTime, err := time.Parse("2006-01-02T15:04", data.Values["start_time"])
	if err != nil {
		data.Errors["start_time"] = "invalid time format"
		app.render(w, "sessioncreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	var notesPtr *string
	if data.Values["notes"] != "" {
		n := data.Values["notes"]
		notesPtr = &n
	}

	sessionID, err := app.DB.NewJobSession(jobID, startTime, notesPtr)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/sessions/"+sessionID.String(), http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) SessionView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	session, err := app.DB.GetJobSession(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	job, err := app.DB.GetJob(&session.JobID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	operatives, err := app.DB.GetJobOperatives(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	resources, err := app.DB.GetJobResources(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), job.Name)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	role, ok := auth.UserRoleFromContext(r.Context())
	data := SessionViewData{
		PageHeader: *pageHeader,
		Session:    session,
		Job:        job,
		IsQS:       ok && model.UserRole(role) == model.UserRoleQS,
	}
	if operatives != nil {
		data.Operatives = *operatives
	}
	if resources != nil {
		data.Resources = *resources
	}

	app.render(w, "sessionview.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) SessionSubmit(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	operative, err := app.DB.GetOperativeByUserID(userID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionSubmit",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrNotFound, err),
		})
		return
	}

	now := time.Now()
	_, err = app.DB.UpdateJobSession(id, &now, &now, &operative.ID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionSubmit",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	session, err := app.DB.GetJobSession(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "SessionSubmit",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/jobs/"+session.JobID.String(), http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) JobResourceCreate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	session, err := app.DB.GetJobSession(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	resources, err := app.DB.GetResources()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Add Resource")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := JobResourceFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
		Session:    session,
	}
	if resources != nil {
		data.Resources = *resources
	}

	app.render(w, "jobresourcecreate.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) JobResourceCreatePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Add Resource")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	session, err := app.DB.GetJobSession(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrNotFound, err),
		})
		return
	}

	resources, err := app.DB.GetResources()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := JobResourceFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
		Session:    session,
	}
	if resources != nil {
		data.Resources = *resources
	}

	data.Values["resource_id"] = r.FormValue("resource_id")
	data.Values["quantity"] = r.FormValue("quantity")
	data.Values["duration_hours"] = r.FormValue("duration_hours")
	data.Values["arrival_time"] = r.FormValue("arrival_time")

	data.ValidateFormFields(FFResourceID)
	if data.HasErrors() {
		app.render(w, "jobresourcecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	resourceID, err := parseUUID(data.Values["resource_id"])
	if err != nil {
		data.Errors["resource_id"] = "invalid resource"
		app.render(w, "jobresourcecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	resource, err := app.DB.GetResource(resourceID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrNotFound, err),
		})
		return
	}

	if (resource.ResourceType == model.ResourceTypeMechanical || resource.ResourceType == model.ResourceTypeTool) && data.Values["duration_hours"] == "" {
		data.Errors["duration_hours"] = "duration is required"
		if resource.ResourceType == model.ResourceTypeMechanical && data.Values["arrival_time"] == "" {
			data.Errors["arrival_time"] = "arrival time is required"
		}
		app.render(w, "jobresourcecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	rate, err := app.DB.GetResourceRate(resourceID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	var quantity, duration *float64
	var arrivalTime *time.Time
	var calculatedCost model.Money

	if data.Values["quantity"] != "" {
		var q float64
		if _, err := fmt.Sscanf(data.Values["quantity"], "%f", &q); err == nil {
			quantity = &q
			calculatedCost = rate.Rate.Multiply(*quantity)
		}
	}

	if data.Values["duration_hours"] != "" {
		var d float64
		if _, err := fmt.Sscanf(data.Values["duration_hours"], "%f", &d); err == nil {
			duration = &d
			calculatedCost = rate.Rate.Multiply(*duration)
		}
	}

	if data.Values["arrival_time"] != "" {
		t, err := time.Parse("2006-01-02T15:04", data.Values["arrival_time"])
		if err == nil {
			arrivalTime = &t
		}
	}

	_, err = app.DB.NewJobResource(
		&session.JobID,
		id,
		resourceID,
		quantity,
		duration,
		arrivalTime,
		rate.Rate,
		&calculatedCost,
	)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/sessions/"+idStr, http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) JobResourceUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	resIDStr := r.PathValue("resId")

	resID, err := parseUUID(resIDStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	_, err = app.DB.UpdateJobResource(resID, nil)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobResourceUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/sessions/"+idStr, http.StatusSeeOther)
}
