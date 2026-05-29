package internal

import (
	"fmt"
	"net/http"
	"time"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
)

type OperativesListData struct {
	model.PageHeader
	Operatives []model.Operative
}

type OperativeViewData struct {
	model.PageHeader
	Operative   *model.Operative
	CurrentRate *model.OperativeRate
}

type OperativeFormData struct {
	model.PageHeader
	*FormFields
}

type OperativeRateFormData struct {
	model.PageHeader
	*FormFields
	Operative   *model.Operative
	CurrentRate *model.OperativeRate
	History     []model.OperativeRate
}

type JobOperativeFormData struct {
	model.PageHeader
	*FormFields
	Session    *model.JobSession
	Operatives []model.Operative
}

func (app *JMcCannBackPackApp) OperativesList(w http.ResponseWriter, r *http.Request) {
	operatives, err := app.DB.GetOperatives()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativesList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Operatives")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativesList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := OperativesListData{PageHeader: *pageHeader}
	if operatives != nil {
		data.Operatives = *operatives
	}

	app.render(w, "operativeslist.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) OperativeCreate(w http.ResponseWriter, r *http.Request) {
	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Operative")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := OperativeFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}

	app.render(w, "operativecreate.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) OperativeCreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Operative")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := OperativeFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}
	data.Values["name"] = r.FormValue("name")
	data.Values["email"] = r.FormValue("email")
	data.Values["password"] = r.FormValue("password")
	data.Values["phone"] = r.FormValue("phone")
	data.Values["trade"] = r.FormValue("trade")
	data.Values["rate"] = r.FormValue("rate")

	data.ValidateFormFields(FFName | FFEmail | FFPassword | FFTrade | FFRate)
	if data.HasErrors() {
		app.render(w, "operativecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	userID, err := app.DB.NewUser(
		data.Values["name"],
		data.Values["email"],
		data.Values["password"],
		model.UserRoleOperative,
	)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	var phonePtr, tradePtr *string
	if data.Values["phone"] != "" {
		p := data.Values["phone"]
		phonePtr = &p
	}
	if data.Values["trade"] != "" {
		t := data.Values["trade"]
		tradePtr = &t
	}

	operativeID, err := app.DB.NewOperative(userID, data.Values["name"], data.Values["email"], phonePtr, tradePtr)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	var rateAmount int
	if _, err := fmt.Sscanf(data.Values["rate"], "%d", &rateAmount); err != nil || rateAmount <= 0 {
		data.Errors["rate"] = "invalid rate amount"
		app.render(w, "operativecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	rate := model.NewMoney("GBP", uint64(rateAmount))
	_, err = app.DB.NewOperativeRate(operativeID, rate, time.Now())
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/operatives/"+operativeID.String(), http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) OperativeView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	operative, err := app.DB.GetOperative(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	currentRate, err := app.DB.GetOperativeRate(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), operative.Name)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := OperativeViewData{
		PageHeader:  *pageHeader,
		Operative:   operative,
		CurrentRate: currentRate,
	}

	app.render(w, "operativeview.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) OperativeUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	trade := r.FormValue("trade")

	var phonePtr, tradePtr *string
	if phone != "" {
		phonePtr = &phone
	}
	if trade != "" {
		tradePtr = &trade
	}

	_, err = app.DB.UpdateOperative(id, name, email, phonePtr, tradePtr)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/operatives/"+idStr, http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) OperativeDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	_, err = app.DB.DeleteOperative(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeDelete",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/operatives", http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) OperativeRateCreate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeRateCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	var rateAmount int
	if _, err := fmt.Sscanf(r.FormValue("rate"), "%d", &rateAmount); err != nil || rateAmount <= 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	currentRate, err := app.DB.GetOperativeRate(id)
	if err == nil && currentRate != nil {
		app.DB.UpdateOperativeRate(&currentRate.ID, time.Now())
	}

	rate := model.NewMoney("GBP", uint64(rateAmount))
	_, err = app.DB.NewOperativeRate(id, rate, time.Now())
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeRateCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/operatives/"+idStr+"/rates", http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) OperativeRateHistory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	operative, err := app.DB.GetOperative(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	currentRate, _ := app.DB.GetOperativeRate(id)

	history, err := app.DB.GetOperativesRates(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeRateHistory",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), operative.Name+" — Rates")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "OperativeRateHistory",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := OperativeRateFormData{
		PageHeader:  *pageHeader,
		FormFields:  NewFormFields(),
		Operative:   operative,
		CurrentRate: currentRate,
	}
	if history != nil {
		data.History = *history
	}

	app.render(w, "operativeratehistory.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) JobOperativeCreate(w http.ResponseWriter, r *http.Request) {
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

	operatives, err := app.DB.GetOperatives()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Add Operative")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := JobOperativeFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
		Session:    session,
	}
	if operatives != nil {
		data.Operatives = *operatives
	}

	app.render(w, "joboperativecreate.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) JobOperativeCreatePost(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Add Operative")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	session, err := app.DB.GetJobSession(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrNotFound, err),
		})
		return
	}

	operatives, err := app.DB.GetOperatives()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := JobOperativeFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
		Session:    session,
	}
	if operatives != nil {
		data.Operatives = *operatives
	}

	data.Values["operative_id"] = r.FormValue("operative_id")
	data.Values["arrival_time"] = r.FormValue("arrival_time")

	data.ValidateFormFields(FFOperativeID | FFArrivalTime)
	if data.HasErrors() {
		app.render(w, "joboperativecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	operativeID, err := parseUUID(data.Values["operative_id"])
	if err != nil {
		data.Errors["operative_id"] = "invalid operative"
		app.render(w, "joboperativecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	arrivalTime, err := time.Parse("2006-01-02T15:04", data.Values["arrival_time"])
	if err != nil {
		data.Errors["arrival_time"] = "invalid time format"
		app.render(w, "joboperativecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	rate, err := app.DB.GetOperativeRate(operativeID)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	_, err = app.DB.NewJobOperative(&session.JobID, id, operativeID, arrivalTime, rate.RatePerHour)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/sessions/"+idStr, http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) JobOperativeUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	opIDStr := r.PathValue("opId")

	opID, err := parseUUID(opIDStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	departureTimeStr := r.FormValue("departure_time")
	departureTime, err := time.Parse("2006-01-02T15:04", departureTimeStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	_, err = app.DB.UpdateJobOperative(opID, &departureTime, nil)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "JobOperativeUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/sessions/"+idStr, http.StatusSeeOther)
}
