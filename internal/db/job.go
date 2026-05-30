package db

import (
	"database/sql"
	"time"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
	"github.com/google/uuid"
)

func (as *AppStorage) NewJob(ref, name string,
	siteId, createdBy *uuid.UUID,
	status *model.JobStatus,
	startDateTime time.Time,
	jobOperativeReq *model.JobOperativeRequirement,
	jobResourceReq *model.JobResourceRequirement,
) (*uuid.UUID, error) {
	if status == nil {
		s := model.JobStatusDraft
		status = &s
	}

	if siteId == nil {
		return nil, &appError.DBError{
			Context: "NewJob",
			Values:  []string{"siteId"},
			Action:  "insert",
			Table:   "job",
			Err:     appError.DBErrBadArgument,
		}
	}

	if createdBy == nil {
		return nil, &appError.DBError{
			Context: "NewJob",
			Values:  []string{"createdBy"},
			Action:  "insert",
			Table:   "job",
			Err:     appError.DBErrBadArgument,
		}
	}

	tx, err := as.DB.Begin()
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewJob",
			Values:  []string{},
			Action:  "transaction",
			Table:   "job, job_operative_requirement, job_resource_requirement",
			Err:     appError.DBErrFatal,
		}
	}
	defer tx.Rollback()

	jobID, jobIDString := UUIDWithString()

	_, err = tx.Exec(
		`insert into job (id, reference, name, site_id, created_by, status, start_datetime)
		 values (?, ?, ?, ?, ?, ?, ?)`,
		jobIDString,
		ref,
		name,
		siteId.String(),
		createdBy.String(),
		string(*status),
		startDateTime.Unix(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewJob",
			Values:  []string{"id", "reference", "name", "site_id", "created_by", "status", "start_datetime"},
			Action:  "insert",
			Table:   "job",
			Err:     appError.DBErrInsert,
		}
	}

	if jobOperativeReq != nil {
		_, reqID := UUIDWithString()
		_, err = tx.Exec(
			`insert into job_operative_requirement (id, job_id, expected_headcount, notes)
			 values (?, ?, ?, ?)`,
			reqID,
			jobID,
			jobOperativeReq.ExpectedHeadcount,
			jobOperativeReq.Notes,
		)
		if err != nil {
			return nil, &appError.DBError{
				Context: "NewJob",
				Values:  []string{"id", "job_id", "expected_headcount", "notes"},
				Action:  "insert",
				Table:   "job_operative_requirement",
				Err:     appError.DBErrInsert,
			}
		}
	}

	if jobResourceReq != nil {
		if jobResourceReq.ResourceID == (uuid.UUID{}) {
			return nil, &appError.DBError{}
		}
		_, reqID := UUIDWithString()
		_, err = tx.Exec(
			`insert into job_resource_requirement (id, job_id, resource_id, expected_quantity, expected_duration_hours, notes)
			 values (?, ?, ?, ?, ?, ?)`,
			reqID,
			jobID,
			jobResourceReq.ResourceID.String(),
			jobResourceReq.ExpectedQuantity,
			jobResourceReq.ExpectedDurationHours,
			jobResourceReq.Notes,
		)
		if err != nil {
			return nil, &appError.DBError{
				Context: "NewJob",
				Values:  []string{"id", "job_id", "resource_id", "expected_quantity", "expected_duration_hours", "notes"},
				Action:  "insert",
				Table:   "job_resource_requirement",
				Err:     appError.DBErrInsert,
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, &appError.DBError{
			Context: "NewJob",
			Values:  []string{},
			Action:  "transaction",
			Table:   "job, job_operative_requirement, job_resource_requirement",
			Err:     appError.DBErrFatal,
		}
	}

	return &jobID, nil
}

func (as *AppStorage) GetJob(id *uuid.UUID) (*model.Job, error) {
	if id == nil {
		return nil, &appError.DBError{
			Context: "GetJob",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "job",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(
		`select id, reference, name, site_id, created_by, status, start_datetime, end_datetime, total_cost
		 from job where id = ?`,
		id.String(),
	)
	var j model.Job
	var idStr, siteIDStr, createdByStr, statusStr string
	var startDatetime float64
	var endDatetime sql.NullFloat64
	var totalCost sql.NullString
	err := row.Scan(
		&idStr, &j.Reference, &j.Name,
		&siteIDStr, &createdByStr, &statusStr,
		&startDatetime, &endDatetime, &totalCost,
	)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetJob",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "job",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetJob",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "job",
			Err:     appError.DBErrQuery,
		}
	}
	j.ID, _ = uuid.Parse(idStr)
	j.SiteID, _ = uuid.Parse(siteIDStr)
	j.CreatedBy, _ = uuid.Parse(createdByStr)
	j.Status = model.JobStatus(statusStr)
	j.StartDatetime = time.Unix(int64(startDatetime), 0)
	if endDatetime.Valid {
		t := time.Unix(int64(endDatetime.Float64), 0)
		j.EndDatetime = &t
	}
	if totalCost.Valid {
		m, err := model.MoneyFromString(totalCost.String)
		if err != nil {
			return nil, &appError.DBError{
				Context: "GetJob",
				Values:  []string{"total_cost"},
				Action:  "deserialize",
				Table:   "job",
				Err:     appError.DBErrSerialization,
			}
		}
		j.TotalCost = m
	}
	return &j, nil
}

func (as *AppStorage) GetJobs() (*[]model.Job, error) {
	rows, err := as.DB.Query(
		`select j.id, j.reference, j.name, j.site_id, j.created_by, j.status, j.start_datetime, j.end_datetime, j.total_cost, s.name
     from job j inner join site s where s.id = j.site_id`,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetJobs",
			Values:  []string{},
			Action:  "query",
			Table:   "job",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var jobs []model.Job
	for rows.Next() {
		var j model.Job
		var idStr, siteIDStr, createdByStr, statusStr string
		var startDatetime float64
		var endDatetime sql.NullFloat64
		var totalCost, siteName sql.NullString
		if err := rows.Scan(
			&idStr, &j.Reference, &j.Name,
			&siteIDStr, &createdByStr, &statusStr,
			&startDatetime, &endDatetime, &totalCost, &siteName,
		); err != nil {
			return nil, &appError.DBError{
				Context: "GetJobs",
				Values:  []string{},
				Action:  "scan",
				Table:   "job",
				Err:     appError.DBErrQuery,
			}
		}
		j.ID, _ = uuid.Parse(idStr)
		j.SiteID, _ = uuid.Parse(siteIDStr)
		j.CreatedBy, _ = uuid.Parse(createdByStr)
		j.Status = model.JobStatus(statusStr)
		j.StartDatetime = time.Unix(int64(startDatetime), 0)
		if endDatetime.Valid {
			t := time.Unix(int64(endDatetime.Float64), 0)
			j.EndDatetime = &t
		}
		if siteName.Valid {
			j.SiteName = &siteName.String
		}
		if totalCost.Valid {
			m, err := model.MoneyFromString(totalCost.String)
			if err != nil {
				return nil, &appError.DBError{
					Context: "GetJobs",
					Values:  []string{"total_cost"},
					Action:  "deserialize",
					Table:   "job",
					Err:     appError.DBErrSerialization,
				}
			}
			j.TotalCost = m
		}
		jobs = append(jobs, j)
	}
	return &jobs, nil
}

func (as *AppStorage) UpdateJob(id *uuid.UUID, status *model.JobStatus, endDatetime *time.Time, totalCost *model.Money) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateJob",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "job",
			Err:     appError.DBErrBadArgument,
		}
	}
	var endTs any
	if endDatetime != nil {
		endTs = endDatetime.Unix()
	}
	var costStr any
	if totalCost != nil {
		costStr = totalCost.Serialize()
	}
	result, err := as.DB.Exec(
		`update job set status = ?, end_datetime = ?, total_cost = ? where id = ?`,
		string(*status), endTs, costStr, id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateJob",
			Values:  []string{"status", "end_datetime", "total_cost"},
			Action:  "update",
			Table:   "job",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) DeleteJob(id *uuid.UUID) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "DeleteJob",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "job",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(`delete from job where id = ?`, id.String())
	if err != nil {
		return 0, &appError.DBError{
			Context: "DeleteJob",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "job",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) NewJobSession(jobID *uuid.UUID, startTime time.Time, notes *string) (*uuid.UUID, error) {
	if jobID == nil {
		return nil, &appError.DBError{
			Context: "NewJobSession",
			Values:  []string{"jobID"},
			Action:  "insert",
			Table:   "job_session",
			Err:     appError.DBErrBadArgument,
		}
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into job_session (id, job_id, session_date, start_time, notes) values (?, ?, ?, ?, ?)`,
		idString,
		jobID.String(),
		startTime.Unix(),
		startTime.Unix(),
		notes,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewJobSession",
			Values:  []string{"id", "job_id", "session_date", "start_time", "notes"},
			Action:  "insert",
			Table:   "job_session",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetJobSession(id *uuid.UUID) (*model.JobSession, error) {
	if id == nil {
		return nil, &appError.DBError{
			Context: "GetJobSession",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "job_session",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(
		`select id, job_id, session_date, start_time, end_time, submitted_at, submitted_by, notes
		 from job_session where id = ?`,
		id.String(),
	)
	var s model.JobSession
	var idStr, jobIDStr string
	var endTime, submittedAt sql.NullFloat64
	var sessionDate, startTime float64
	var submittedBy, notes sql.NullString
	err := row.Scan(
		&idStr, &jobIDStr, &sessionDate, &startTime,
		&endTime, &submittedAt, &submittedBy, &notes,
	)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetJobSession",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "job_session",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetJobSession",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "job_session",
			Err:     appError.DBErrQuery,
		}
	}
	s.ID, _ = uuid.Parse(idStr)
	s.JobID, _ = uuid.Parse(jobIDStr)
	s.SessionDate = time.Unix(int64(sessionDate), 0)
	s.StartTime = time.Unix(int64(startTime), 0)
	if endTime.Valid {
		t := time.Unix(int64(endTime.Float64), 0)
		s.EndTime = &t
	}
	if submittedAt.Valid {
		t := time.Unix(int64(submittedAt.Float64), 0)
		s.SubmittedAt = &t
	}
	if submittedBy.Valid {
		u, _ := uuid.Parse(submittedBy.String)
		s.SubmittedBy = &u
	}
	if notes.Valid {
		s.Notes = &notes.String
	}
	return &s, nil
}

func (as *AppStorage) GetJobSessions(jobID *uuid.UUID) (*[]model.JobSession, error) {
	if jobID == nil {
		return nil, &appError.DBError{
			Context: "GetJobSessions",
			Values:  []string{"jobID"},
			Action:  "query",
			Table:   "job_session",
			Err:     appError.DBErrBadArgument,
		}
	}
	rows, err := as.DB.Query(
		`select s.id, s.job_id, s.session_date, s.start_time, s.end_time, s.submitted_at, s.submitted_by, s.notes, o.name
           from job_session s left join operative o on o.id = s.submitted_by where s.job_id = ?;`,
		jobID.String(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetJobSessions",
			Values:  []string{"jobID"},
			Action:  "query",
			Table:   "job_session",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var sessions []model.JobSession
	for rows.Next() {
		var s model.JobSession
		var idStr, jobIDStr string
		var endTime, submittedAt sql.NullFloat64
		var sessionDate, startTime float64
		var submittedBy, notes, submittedByName sql.NullString
		if err := rows.Scan(
			&idStr, &jobIDStr, &sessionDate, &startTime,
			&endTime, &submittedAt, &submittedBy, &notes, &submittedByName,
		); err != nil {
			return nil, &appError.DBError{
				Context: "GetJobSessions",
				Values:  []string{},
				Action:  "scan",
				Table:   "job_session",
				Err:     appError.DBErrQuery,
			}
		}
		s.ID, _ = uuid.Parse(idStr)
		s.JobID, _ = uuid.Parse(jobIDStr)
		s.SessionDate = time.Unix(int64(sessionDate), 0)
		s.StartTime = time.Unix(int64(startTime), 0)
		if endTime.Valid {
			t := time.Unix(int64(endTime.Float64), 0)
			s.EndTime = &t
		}
		if submittedAt.Valid {
			t := time.Unix(int64(submittedAt.Float64), 0)
			s.SubmittedAt = &t
		}
		if submittedBy.Valid {
			u, _ := uuid.Parse(submittedBy.String)
			s.SubmittedBy = &u
		}
		if submittedByName.Valid {
			s.SubmittedByName = &submittedByName.String
		}
		if notes.Valid {
			s.Notes = &notes.String
		}
		sessions = append(sessions, s)
	}
	return &sessions, nil
}

func (as *AppStorage) UpdateJobSession(id *uuid.UUID, endTime *time.Time, submittedAt *time.Time, submittedBy *uuid.UUID) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateJobSession",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "job_session",
			Err:     appError.DBErrBadArgument,
		}
	}
	var endTs, subAt, subBy any
	if endTime != nil {
		endTs = endTime.Unix()
	}
	if submittedAt != nil {
		subAt = submittedAt.Unix()
	}
	if submittedBy != nil {
		subBy = submittedBy.String()
	}
	result, err := as.DB.Exec(
		`update job_session set end_time = ?, submitted_at = ?, submitted_by = ? where id = ?`,
		endTs, subAt, subBy, id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateJobSession",
			Values:  []string{"end_time", "submitted_at", "submitted_by"},
			Action:  "update",
			Table:   "job_session",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) NewJobOperative(jobID, sessionID, operativeID *uuid.UUID, arrivalTime time.Time, rateSnapshot model.Money) (*uuid.UUID, error) {
	if jobID == nil || sessionID == nil || operativeID == nil {
		return nil, &appError.DBError{
			Context: "NewJobOperative",
			Values:  []string{"jobID", "sessionID", "operativeID"},
			Action:  "insert",
			Table:   "job_operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into job_operative (id, job_id, session_id, operative_id, arrival_time, rate_snapshot)
		 values (?, ?, ?, ?, ?, ?)`,
		idString,
		jobID.String(),
		sessionID.String(),
		operativeID.String(),
		arrivalTime.Unix(),
		rateSnapshot.Serialize(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewJobOperative",
			Values:  []string{"id", "job_id", "session_id", "operative_id", "arrival_time", "rate_snapshot"},
			Action:  "insert",
			Table:   "job_operative",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetJobOperatives(sessionID *uuid.UUID) (*[]model.JobOperative, error) {
	if sessionID == nil {
		return nil, &appError.DBError{
			Context: "GetJobOperatives",
			Values:  []string{"sessionID"},
			Action:  "query",
			Table:   "job_operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	rows, err := as.DB.Query(
		`select j.id, j.job_id, j.session_id, j.operative_id, j.arrival_time, j.departure_time, j.rate_snapshot, j.calculated_cost, o.name
		 from job_operative j inner join operative o on o.id = j.operative_id where session_id = ?`,
		sessionID.String(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetJobOperatives",
			Values:  []string{"sessionID"},
			Action:  "query",
			Table:   "job_operative",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var operatives []model.JobOperative
	for rows.Next() {
		var jo model.JobOperative
		var idStr, jobIDStr, sessionIDStr, operativeIDStr, rateStr string
		var arrivalTime float64
		var departureTime sql.NullFloat64
		var calculatedCost, operativeName sql.NullString
		if err := rows.Scan(
			&idStr, &jobIDStr, &sessionIDStr, &operativeIDStr,
			&arrivalTime, &departureTime, &rateStr, &calculatedCost, &operativeName,
		); err != nil {
			return nil, &appError.DBError{
				Context: "GetJobOperatives",
				Values:  []string{},
				Action:  "scan",
				Table:   "job_operative",
				Err:     appError.DBErrQuery,
			}
		}
		jo.ID, _ = uuid.Parse(idStr)
		jo.JobID, _ = uuid.Parse(jobIDStr)
		jo.SessionID, _ = uuid.Parse(sessionIDStr)
		jo.OperativeID, _ = uuid.Parse(operativeIDStr)
		jo.ArrivalTime = time.Unix(int64(arrivalTime), 0)
		m, err := model.MoneyFromString(rateStr)
		if err != nil {
			return nil, &appError.DBError{
				Context: "GetJobOperatives",
				Values:  []string{"rate_snapshot"},
				Action:  "deserialize",
				Table:   "job_operative",
				Err:     appError.DBErrSerialization,
			}
		}
		jo.RateSnapshot = *m
		if departureTime.Valid {
			t := time.Unix(int64(departureTime.Float64), 0)
			jo.DepartureTime = &t
		}
		if calculatedCost.Valid {
			m, err := model.MoneyFromString(calculatedCost.String)
			if err != nil {
				return nil, &appError.DBError{
					Context: "GetJobOperatives",
					Values:  []string{"calculated_cost"},
					Action:  "deserialize",
					Table:   "job_operative",
					Err:     appError.DBErrSerialization,
				}
			}
			jo.CalculatedCost = m
		}
		if operativeName.Valid {
			jo.OperativeName = &operativeName.String
		}
		operatives = append(operatives, jo)
	}
	return &operatives, nil
}

func (as *AppStorage) UpdateJobOperative(id *uuid.UUID, departureTime *time.Time, calculatedCost *model.Money) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateJobOperative",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "job_operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	var depTs, costStr any
	if departureTime != nil {
		depTs = departureTime.Unix()
	}
	if calculatedCost != nil {
		costStr = calculatedCost.Serialize()
	}
	result, err := as.DB.Exec(
		`update job_operative set departure_time = ?, calculated_cost = ? where id = ?`,
		depTs, costStr, id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateJobOperative",
			Values:  []string{"departure_time", "calculated_cost"},
			Action:  "update",
			Table:   "job_operative",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) NewJobResource(jobID, sessionID, resourceID *uuid.UUID, quantity *float64, durationHours *float64, arrivalTime *time.Time, rateSnapshot model.Money, calculatedCost *model.Money) (*uuid.UUID, error) {
	if jobID == nil || sessionID == nil || resourceID == nil {
		return nil, &appError.DBError{
			Context: "NewJobResource",
			Values:  []string{"jobID", "sessionID", "resourceID"},
			Action:  "insert",
			Table:   "job_resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	var arrTs any
	if arrivalTime != nil {
		arrTs = arrivalTime.Unix()
	}
	var costStr any
	if calculatedCost != nil {
		costStr = calculatedCost.Serialize()
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into job_resource (id, job_id, session_id, resource_id, quantity, duration_hours, arrival_time, rate_snapshot, calculated_cost)
		 values (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		idString,
		jobID.String(),
		sessionID.String(),
		resourceID.String(),
		quantity,
		durationHours,
		arrTs,
		rateSnapshot.Serialize(),
		costStr,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewJobResource",
			Values:  []string{"id", "job_id", "session_id", "resource_id", "quantity", "duration_hours", "arrival_time", "rate_snapshot", "calculated_cost"},
			Action:  "insert",
			Table:   "job_resource",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetJobResources(sessionID *uuid.UUID) (*[]model.JobResource, error) {
	if sessionID == nil {
		return nil, &appError.DBError{
			Context: "GetJobResources",
			Values:  []string{"sessionID"},
			Action:  "query",
			Table:   "job_resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	rows, err := as.DB.Query(
		`select j.id, j.job_id, j.session_id, j.resource_id, j.quantity, j.duration_hours, j.arrival_time, j.rate_snapshot, j.calculated_cost, r.name, r.resource_type
		 from job_resource j inner join resource r on r.id = j.resource_id
		 where j.session_id = ?`,
		sessionID.String(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetJobResources",
			Values:  []string{"sessionID"},
			Action:  "query",
			Table:   "job_resource",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var resources []model.JobResource
	for rows.Next() {
		var jr model.JobResource
		var idStr, jobIDStr, sessionIDStr, resourceIDStr, rateStr string
		var quantity, durationHours sql.NullFloat64
		var arrivalTime sql.NullFloat64
		var calculatedCost sql.NullString
		if err := rows.Scan(
			&idStr, &jobIDStr, &sessionIDStr, &resourceIDStr,
			&quantity, &durationHours, &arrivalTime, &rateStr, &calculatedCost,
			&jr.ResourceName, &jr.ResourceType,
		); err != nil {
			return nil, &appError.DBError{
				Context: "GetJobResources",
				Values:  []string{},
				Action:  "scan",
				Table:   "job_resource",
				Err:     appError.DBErrQuery,
			}
		}
		jr.ID, _ = uuid.Parse(idStr)
		jr.JobID, _ = uuid.Parse(jobIDStr)
		jr.SessionID, _ = uuid.Parse(sessionIDStr)
		jr.ResourceID, _ = uuid.Parse(resourceIDStr)
		m, err := model.MoneyFromString(rateStr)
		if err != nil {
			return nil, &appError.DBError{
				Context: "GetJobResources",
				Values:  []string{"rate_snapshot"},
				Action:  "deserialize",
				Table:   "job_resource",
				Err:     appError.DBErrSerialization,
			}
		}
		jr.RateSnapshot = *m
		if quantity.Valid {
			jr.Quantity = &quantity.Float64
		}
		if durationHours.Valid {
			jr.DurationHours = &durationHours.Float64
		}
		if arrivalTime.Valid {
			t := time.Unix(int64(arrivalTime.Float64), 0)
			jr.ArrivalTime = &t
		}
		if calculatedCost.Valid {
			m, err := model.MoneyFromString(calculatedCost.String)
			if err != nil {
				return nil, &appError.DBError{
					Context: "GetJobResources",
					Values:  []string{"calculated_cost"},
					Action:  "deserialize",
					Table:   "job_resource",
					Err:     appError.DBErrSerialization,
				}
			}
			jr.CalculatedCost = m
		}
		resources = append(resources, jr)
	}
	return &resources, nil
}

func (as *AppStorage) UpdateJobResource(id *uuid.UUID, calculatedCost *model.Money) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateJobResource",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "job_resource",
			Err:     appError.DBErrBadArgument,
		}
	}
	var costStr any
	if calculatedCost != nil {
		costStr = calculatedCost.Serialize()
	}
	result, err := as.DB.Exec(
		`update job_resource set calculated_cost = ? where id = ?`,
		costStr, id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateJobResource",
			Values:  []string{"calculated_cost"},
			Action:  "update",
			Table:   "job_resource",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) GetOpenSessionByOperative(operativeID string) (*model.JobSession, error) {
	row := as.DB.QueryRow(
		`select id, job_id, session_date, start_time, end_time, submitted_at, submitted_by, notes
		 from job_session
		 where submitted_by = ? and submitted_at is null
		 limit 1`,
		operativeID,
	)
	var s model.JobSession
	var idStr, jobIDStr string
	var sessionDate, startTime float64
	var endTime sql.NullFloat64
	var submittedBy, notes sql.NullString
	var submittedAt sql.NullFloat64
	err := row.Scan(
		&idStr, &jobIDStr, &sessionDate, &startTime,
		&endTime, &submittedAt, &submittedBy, &notes,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOpenSessionByOperative",
			Values:  []string{"operativeID"},
			Action:  "query",
			Table:   "job_session",
			Err:     appError.DBErrQuery,
		}
	}
	s.ID, _ = uuid.Parse(idStr)
	s.JobID, _ = uuid.Parse(jobIDStr)
	s.SessionDate = time.Unix(int64(sessionDate), 0)
	s.StartTime = time.Unix(int64(startTime), 0)
	if endTime.Valid {
		t := time.Unix(int64(endTime.Float64), 0)
		s.EndTime = &t
	}
	if submittedAt.Valid {
		t := time.Unix(int64(submittedAt.Float64), 0)
		s.SubmittedAt = &t
	}
	if submittedBy.Valid {
		u, _ := uuid.Parse(submittedBy.String)
		s.SubmittedBy = &u
	}
	if notes.Valid {
		s.Notes = &notes.String
	}
	return &s, nil
}
