package db

import (
	"database/sql"
	"time"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
	"github.com/google/uuid"
)

func (as *AppStorage) NewResource(name string, resourceType model.ResourceType, unitOfMeasure *string) (*uuid.UUID, error) {
	if name == "" {
		return nil, &appError.DBError{
			Context: "NewResource",
			Values:  []string{"name"},
			Action:  "insert",
			Table:   "resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into resource (id, name, resource_type, unit_of_measure) values (?, ?, ?, ?)`,
		idString, name, string(resourceType), unitOfMeasure,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewResource",
			Values:  []string{"id", "name", "resource_type", "unit_of_measure"},
			Action:  "insert",
			Table:   "resource",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetResource(id *uuid.UUID) (*model.Resource, error) {
	if id == nil {
		return nil, &appError.DBError{
			Context: "GetResource",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(
		`select id, name, resource_type, unit_of_measure from resource where id = ?`,
		id.String(),
	)
	var r model.Resource
	var idStr, resourceTypeStr string
	var unitOfMeasure sql.NullString
	err := row.Scan(&idStr, &r.Name, &resourceTypeStr, &unitOfMeasure)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetResource",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "resource",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetResource",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "resource",
			Err:     appError.DBErrQuery,
		}
	}
	r.ID, _ = uuid.Parse(idStr)
	r.ResourceType = model.ResourceType(resourceTypeStr)
	if unitOfMeasure.Valid {
		r.UnitOfMeasure = &unitOfMeasure.String
	}
	return &r, nil
}

func (as *AppStorage) GetResourceType(id *uuid.UUID) (*model.ResourceType, error) {
	if id == nil {
		return nil, &appError.DBError{
			Context: "GetResourceType",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(`select resource_type from resource where id = ?`, id.String())
	var resourceTypeStr string
	err := row.Scan(&resourceTypeStr)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetResourceType",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "resource",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetResourceType",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "resource",
			Err:     appError.DBErrQuery,
		}
	}
	rt := model.ResourceType(resourceTypeStr)
	return &rt, nil
}

func (as *AppStorage) GetResources() (*[]model.Resource, error) {
	rows, err := as.DB.Query(
		`select id, name, resource_type, unit_of_measure from resource`,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetResources",
			Values:  []string{},
			Action:  "query",
			Table:   "resource",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var resources []model.Resource
	for rows.Next() {
		var r model.Resource
		var idStr, resourceTypeStr string
		var unitOfMeasure sql.NullString
		if err := rows.Scan(&idStr, &r.Name, &resourceTypeStr, &unitOfMeasure); err != nil {
			return nil, &appError.DBError{
				Context: "GetResources",
				Values:  []string{},
				Action:  "scan",
				Table:   "resource",
				Err:     appError.DBErrQuery,
			}
		}
		r.ID, _ = uuid.Parse(idStr)
		r.ResourceType = model.ResourceType(resourceTypeStr)
		if unitOfMeasure.Valid {
			r.UnitOfMeasure = &unitOfMeasure.String
		}
		resources = append(resources, r)
	}
	return &resources, nil
}

func (as *AppStorage) UpdateResource(id *uuid.UUID, name string, unitOfMeasure *string) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateResource",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(
		`update resource set name = ?, unit_of_measure = ? where id = ?`,
		name, unitOfMeasure, id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateResource",
			Values:  []string{"name", "unit_of_measure"},
			Action:  "update",
			Table:   "resource",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) DeleteResource(id *uuid.UUID) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "DeleteResource",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(`delete from resource where id = ?`, id.String())
	if err != nil {
		return 0, &appError.DBError{
			Context: "DeleteResource",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "resource",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) NewResourceRate(resourceID *uuid.UUID, rate model.Money, costUnit model.CostUnit, effectiveFrom time.Time) (*uuid.UUID, error) {
	if resourceID == nil {
		return nil, &appError.DBError{
			Context: "NewResourceRate",
			Values:  []string{"resourceID"},
			Action:  "insert",
			Table:   "resource_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into resource_rate (id, resource_id, rate, cost_unit, effective_from) values (?, ?, ?, ?, ?)`,
		idString, resourceID.String(), rate.Serialize(), string(costUnit), effectiveFrom.Unix(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewResourceRate",
			Values:  []string{"id", "resource_id", "rate", "cost_unit", "effective_from"},
			Action:  "insert",
			Table:   "resource_rate",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetResourceRate(resourceID *uuid.UUID) (*model.ResourceRate, error) {
	if resourceID == nil {
		return nil, &appError.DBError{
			Context: "GetResourceRate",
			Values:  []string{"resourceID"},
			Action:  "query",
			Table:   "resource_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(
		`select id, resource_id, rate, cost_unit, effective_from, effective_to
		 from resource_rate
		 where resource_id = ? and (effective_to is null or effective_to > ?)
		 order by effective_from desc limit 1`,
		resourceID.String(), time.Now().Unix(),
	)
	var r model.ResourceRate
	var idStr, resourceIDStr, rateStr, costUnitStr string
	var effectiveFrom float64
	var effectiveTo sql.NullFloat64
	err := row.Scan(&idStr, &resourceIDStr, &rateStr, &costUnitStr, &effectiveFrom, &effectiveTo)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetResourceRate",
			Values:  []string{"resourceID"},
			Action:  "query",
			Table:   "resource_rate",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetResourceRate",
			Values:  []string{"resourceID"},
			Action:  "query",
			Table:   "resource_rate",
			Err:     appError.DBErrQuery,
		}
	}
	r.ID, _ = uuid.Parse(idStr)
	r.ResourceID, _ = uuid.Parse(resourceIDStr)
	r.EffectiveFrom = time.Unix(int64(effectiveFrom), 0)
	m, err := model.MoneyFromString(rateStr)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetResourceRate",
			Values:  []string{"rate"},
			Action:  "deserialize",
			Table:   "resource_rate",
			Err:     appError.DBErrSerialization,
		}
	}
	r.Rate = *m
	r.CostUnit = model.CostUnit(costUnitStr)
	if effectiveTo.Valid {
		t := time.Unix(int64(effectiveTo.Float64), 0)
		r.EffectiveTo = &t
	}
	return &r, nil
}

func (as *AppStorage) GetResourcesRates(resourceID *uuid.UUID) (*[]model.ResourceRate, error) {
	if resourceID == nil {
		return nil, &appError.DBError{
			Context: "GetResourcesRates",
			Values:  []string{"resourceID"},
			Action:  "query",
			Table:   "resource_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	rows, err := as.DB.Query(
		`select id, resource_id, rate, cost_unit, effective_from, effective_to
		 from resource_rate where resource_id = ? order by effective_from desc`,
		resourceID.String(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetResourcesRates",
			Values:  []string{"resourceID"},
			Action:  "query",
			Table:   "resource_rate",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var rates []model.ResourceRate
	for rows.Next() {
		var r model.ResourceRate
		var idStr, resourceIDStr, rateStr, costUnitStr string
		var effectiveFrom float64
		var effectiveTo sql.NullFloat64
		if err := rows.Scan(&idStr, &resourceIDStr, &rateStr, &costUnitStr, &effectiveFrom, &effectiveTo); err != nil {
			return nil, &appError.DBError{
				Context: "GetResourcesRates",
				Values:  []string{},
				Action:  "scan",
				Table:   "resource_rate",
				Err:     appError.DBErrQuery,
			}
		}
		r.ID, _ = uuid.Parse(idStr)
		r.ResourceID, _ = uuid.Parse(resourceIDStr)
		r.EffectiveFrom = time.Unix(int64(effectiveFrom), 0)
		m, err := model.MoneyFromString(rateStr)
		if err != nil {
			return nil, &appError.DBError{
				Context: "GetResourcesRates",
				Values:  []string{"rate"},
				Action:  "deserialize",
				Table:   "resource_rate",
				Err:     appError.DBErrSerialization,
			}
		}
		r.Rate = *m
		r.CostUnit = model.CostUnit(costUnitStr)
		if effectiveTo.Valid {
			t := time.Unix(int64(effectiveTo.Float64), 0)
			r.EffectiveTo = &t
		}
		rates = append(rates, r)
	}
	return &rates, nil
}

func (as *AppStorage) UpdateResourceRate(id *uuid.UUID, effectiveTo time.Time) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateResourceRate",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "resource_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(
		`update resource_rate set effective_to = ? where id = ?`,
		effectiveTo.Unix(), id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateResourceRate",
			Values:  []string{"effective_to"},
			Action:  "update",
			Table:   "resource_rate",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) DeleteResourceRate(id *uuid.UUID) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "DeleteResourceRate",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "resource_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(`delete from resource_rate where id = ?`, id.String())
	if err != nil {
		return 0, &appError.DBError{
			Context: "DeleteResourceRate",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "resource_rate",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}
