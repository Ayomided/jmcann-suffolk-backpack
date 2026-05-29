package internal

import (
	"fmt"
	"net/http"
	"time"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
)

type ResourcesListData struct {
	model.PageHeader
	Resources []model.Resource
}

type ResourceViewData struct {
	model.PageHeader
	Resource    *model.Resource
	CurrentRate *model.ResourceRate
}

type ResourceFormData struct {
	model.PageHeader
	*FormFields
}

type ResourceRateFormData struct {
	model.PageHeader
	*FormFields
	Resource    *model.Resource
	CurrentRate *model.ResourceRate
	History     []model.ResourceRate
}

func (app *JMcCannBackPackApp) ResourcesList(w http.ResponseWriter, r *http.Request) {
	resources, err := app.DB.GetResources()
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourcesList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Resources")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourcesList",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := ResourcesListData{PageHeader: *pageHeader}
	if resources != nil {
		data.Resources = *resources
	}

	app.render(w, "resourceslist.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) ResourceCreate(w http.ResponseWriter, r *http.Request) {
	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Resource")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := ResourceFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}

	app.render(w, "resourcecreate.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) ResourceCreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), "Create Resource")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := ResourceFormData{
		PageHeader: *pageHeader,
		FormFields: NewFormFields(),
	}
	data.Values["name"] = r.FormValue("name")
	data.Values["resource_type"] = r.FormValue("resource_type")
	data.Values["unit_of_measure"] = r.FormValue("unit_of_measure")
	data.Values["rate"] = r.FormValue("rate")
	data.Values["cost_unit"] = r.FormValue("cost_unit")

	data.ValidateFormFields(FFName | FFResourceType | FFRate | FFCostUnit)
	if data.HasErrors() {
		app.render(w, "resourcecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	var unitPtr *string
	if data.Values["unit_of_measure"] != "" {
		u := data.Values["unit_of_measure"]
		unitPtr = &u
	}

	resourceID, err := app.DB.NewResource(
		data.Values["name"],
		model.ResourceType(data.Values["resource_type"]),
		unitPtr,
	)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	var rateAmount int
	if _, err := fmt.Sscanf(data.Values["rate"], "%d", &rateAmount); err != nil || rateAmount <= 0 {
		data.Errors["rate"] = "invalid rate amount"
		app.render(w, "resourcecreate.templ.html", http.StatusUnprocessableEntity, data)
		return
	}

	rate := model.NewMoney("GBP", uint64(rateAmount))
	_, err = app.DB.NewResourceRate(resourceID, rate, model.CostUnit(data.Values["cost_unit"]), time.Now())
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceCreatePost",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/resources/"+resourceID.String(), http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) ResourceView(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	resource, err := app.DB.GetResource(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	currentRate, err := app.DB.GetResourceRate(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), resource.Name)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceView",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := ResourceViewData{
		PageHeader:  *pageHeader,
		Resource:    resource,
		CurrentRate: currentRate,
	}

	app.render(w, "resourceview.templ.html", http.StatusOK, data)
}

func (app *JMcCannBackPackApp) ResourceUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	name := r.FormValue("name")
	unitOfMeasure := r.FormValue("unit_of_measure")

	var unitPtr *string
	if unitOfMeasure != "" {
		unitPtr = &unitOfMeasure
	}

	_, err = app.DB.UpdateResource(id, name, unitPtr)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceUpdate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/resources/"+idStr, http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) ResourceDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	_, err = app.DB.DeleteResource(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceDelete",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/resources", http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) ResourceRateCreate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceRateCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrBadRequest, err),
		})
		return
	}

	var rateAmount int
	if _, err := fmt.Sscanf(r.FormValue("rate"), "%d", &rateAmount); err != nil || rateAmount <= 0 {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	currentRate, err := app.DB.GetResourceRate(id)
	if err == nil && currentRate != nil {
		app.DB.UpdateResourceRate(&currentRate.ID, time.Now())
	}

	rate := model.NewMoney("GBP", uint64(rateAmount))
	_, err = app.DB.NewResourceRate(id, rate, model.CostUnit(r.FormValue("cost_unit")), time.Now())
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceRateCreate",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	http.Redirect(w, r, "/resources/"+idStr+"/rates", http.StatusSeeOther)
}

func (app *JMcCannBackPackApp) ResourceRateHistory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := parseUUID(idStr)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	resource, err := app.DB.GetResource(id)
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	currentRate, _ := app.DB.GetResourceRate(id)

	history, err := app.DB.GetResourcesRates(id)
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceRateHistory",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	pageHeader, err := app.pageHeaderFromContext(r.Context(), resource.Name+" — Rates")
	if err != nil {
		app.serverError(w, &appError.ServerError{
			Handler: "ResourceRateHistory",
			Err:     fmt.Errorf("%w: %w", appError.ServerErrFailed, err),
		})
		return
	}

	data := ResourceRateFormData{
		PageHeader:  *pageHeader,
		FormFields:  NewFormFields(),
		Resource:    resource,
		CurrentRate: currentRate,
	}
	if history != nil {
		data.History = *history
	}

	app.render(w, "resourceratehistory.templ.html", http.StatusOK, data)
}
