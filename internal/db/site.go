package db

import (
	"database/sql"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
	"github.com/google/uuid"
)

func (as *AppStorage) NewSite(name, address string) (*uuid.UUID, error) {
	if name == "" {
		return nil, &appError.DBError{
			Context: "NewSite",
			Values:  []string{"name"},
			Action:  "insert",
			Table:   "site",
			Err:     appError.DBErrBadArgument,
		}
	}
	if address == "" {
		return nil, &appError.DBError{
			Context: "NewSite",
			Values:  []string{"address"},
			Action:  "insert",
			Table:   "site",
			Err:     appError.DBErrBadArgument,
		}
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into site (id, name, address) values (?, ?, ?)`,
		idString, name, address,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewSite",
			Values:  []string{"id", "name", "address"},
			Action:  "insert",
			Table:   "site",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetSite(id *uuid.UUID) (*model.Site, error) {
	if id == nil {
		return nil, &appError.DBError{
			Context: "GetSite",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "site",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(
		`select id, name, address from site where id = ?`,
		id.String(),
	)
	var s model.Site
	var idStr string
	err := row.Scan(&idStr, &s.Name, &s.Address)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetSite",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "site",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetSite",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "site",
			Err:     appError.DBErrQuery,
		}
	}
	s.ID, _ = uuid.Parse(idStr)
	return &s, nil
}

func (as *AppStorage) GetSites() (*[]model.Site, error) {
	rows, err := as.DB.Query(`select id, name, address from site`)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetSites",
			Values:  []string{},
			Action:  "query",
			Table:   "site",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var sites []model.Site
	for rows.Next() {
		var s model.Site
		var idStr string
		if err := rows.Scan(&idStr, &s.Name, &s.Address); err != nil {
			return nil, &appError.DBError{
				Context: "GetSites",
				Values:  []string{},
				Action:  "scan",
				Table:   "site",
				Err:     appError.DBErrQuery,
			}
		}
		s.ID, _ = uuid.Parse(idStr)
		sites = append(sites, s)
	}
	return &sites, nil
}
