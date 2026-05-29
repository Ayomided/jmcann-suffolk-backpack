package db

import (
	"database/sql"
	"time"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (as *AppStorage) NewUser(name, email, password string, role model.UserRole) (*uuid.UUID, error) {
	if name == "" {
		return nil, &appError.DBError{
			Context: "NewUser",
			Values:  []string{"name"},
			Action:  "insert",
			Table:   "user",
			Err:     appError.DBErrBadArgument,
		}
	}
	if email == "" {
		return nil, &appError.DBError{
			Context: "NewUser",
			Values:  []string{"email"},
			Action:  "insert",
			Table:   "user",
			Err:     appError.DBErrBadArgument,
		}
	}
	if password == "" {
		return nil, &appError.DBError{
			Context: "NewUser",
			Values:  []string{"password"},
			Action:  "insert",
			Table:   "user",
			Err:     appError.DBErrBadArgument,
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewUser",
			Values:  []string{"password"},
			Action:  "hash",
			Table:   "user",
			Err:     appError.DBErrFatal,
		}
	}
	id, idString := UUIDWithString()
	_, err = as.DB.Exec(
		`insert into user (id, name, email, role, password_hash) values (?, ?, ?, ?, ?)`,
		idString, name, email, string(role), string(hash),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewUser",
			Values:  []string{"id", "name", "email", "role", "password_hash"},
			Action:  "insert",
			Table:   "user",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetUserByEmail(email string) (*model.User, error) {
	row := as.DB.QueryRow(
		`select id, name, email, role, password_hash from user where email = ?`,
		email,
	)
	var u model.User
	var idStr, roleStr string
	err := row.Scan(&idStr, &u.Name, &u.Email, &roleStr, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetUserByEmail",
			Values:  []string{"email"},
			Action:  "query",
			Table:   "user",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetUserByEmail",
			Values:  []string{"email"},
			Action:  "query",
			Table:   "user",
			Err:     appError.DBErrQuery,
		}
	}
	u.ID, _ = uuid.Parse(idStr)
	u.Role = model.UserRole(roleStr)
	return &u, nil
}

func (as *AppStorage) GetUserById(id string) (*model.User, error) {
	row := as.DB.QueryRow(
		`select id, name, email, role, password_hash from user where id = ?`,
		id,
	)
	var u model.User
	var idStr, roleStr string
	err := row.Scan(&idStr, &u.Name, &u.Email, &roleStr, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetUserById",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "user",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetUserById",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "user",
			Err:     appError.DBErrQuery,
		}
	}
	u.ID, _ = uuid.Parse(idStr)
	u.Role = model.UserRole(roleStr)
	return &u, nil
}

func (as *AppStorage) NewOperative(userID *uuid.UUID, name, email string, phone, trade *string) (*uuid.UUID, error) {
	if userID == nil {
		return nil, &appError.DBError{
			Context: "NewOperative",
			Values:  []string{"userID"},
			Action:  "insert",
			Table:   "operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	if name == "" || email == "" {
		return nil, &appError.DBError{
			Context: "NewOperative",
			Values:  []string{"name", "email"},
			Action:  "insert",
			Table:   "operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into operative (id, user_id, name, email, phone, trade) values (?, ?, ?, ?, ?, ?)`,
		idString, userID.String(), name, email, phone, trade,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewOperative",
			Values:  []string{"id", "user_id", "name", "email", "phone", "trade"},
			Action:  "insert",
			Table:   "operative",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetOperative(id *uuid.UUID) (*model.Operative, error) {
	if id == nil {
		return nil, &appError.DBError{
			Context: "GetOperative",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(
		`select id, user_id, name, email, phone, trade from operative where id = ?`,
		id.String(),
	)
	var o model.Operative
	var idStr, userIDStr string
	var phone, trade sql.NullString
	err := row.Scan(&idStr, &userIDStr, &o.Name, &o.Email, &phone, &trade)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetOperative",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperative",
			Values:  []string{"id"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrQuery,
		}
	}
	o.ID, _ = uuid.Parse(idStr)
	o.UserID, _ = uuid.Parse(userIDStr)
	if phone.Valid {
		o.Phone = phone.String
	}
	if trade.Valid {
		o.Trade = &trade.String
	}
	return &o, nil
}

func (as *AppStorage) GetOperativeByUserID(userID string) (*model.Operative, error) {
	row := as.DB.QueryRow(
		`select id, user_id, name, email, phone, trade from operative where user_id = ?`,
		userID,
	)
	var o model.Operative
	var idStr, userIDStr string
	var phone, trade sql.NullString
	err := row.Scan(&idStr, &userIDStr, &o.Name, &o.Email, &phone, &trade)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetOperativeByUserID",
			Values:  []string{"userID"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperativeByUserID",
			Values:  []string{"userID"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrQuery,
		}
	}
	o.ID, _ = uuid.Parse(idStr)
	o.UserID, _ = uuid.Parse(userIDStr)
	if phone.Valid {
		o.Phone = phone.String
	}
	if trade.Valid {
		o.Trade = &trade.String
	}
	return &o, nil
}

func (as *AppStorage) GetOperativeByUserIDWithRate(userID string) (*model.Operative, error) {
	row := as.DB.QueryRow(
		`select o.id, o.user_id, o.name, o.email, o.phone, o.trade, r.rate_per_hour from operative o inner join operative_rate r on r.operative_id = o.id where o.user_id = ?`,
		userID,
	)
	var o model.Operative
	var idStr, userIDStr string
	var phone, trade, rate sql.NullString
	err := row.Scan(&idStr, &userIDStr, &o.Name, &o.Email, &phone, &trade, &rate)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetOperativeByUserIDWithRate",
			Values:  []string{"userID"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperativeByUserIDWithRate",
			Values:  []string{"userID"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrQuery,
		}
	}
	o.ID, _ = uuid.Parse(idStr)
	o.UserID, _ = uuid.Parse(userIDStr)
	if phone.Valid {
		o.Phone = phone.String
	}
	if trade.Valid {
		o.Trade = &trade.String
	}
	if rate.Valid {
		rateAmount, err := model.MoneyFromString(rate.String)
		if err != nil {
			return nil, &appError.DBError{
				Context: "GetOperativeByUserIDWithRate",
				Values:  []string{"rate"},
				Action:  "convert rate from string to money",
				Table:   "operative",
				Err:     appError.DBErrQuery,
			}
		}
		o.Rate = rateAmount
	}
	return &o, nil
}

func (as *AppStorage) GetOperativeTradeByUserID(userID string) (*string, error) {
	row := as.DB.QueryRow(
		`select trade from operative where user_id = ?`,
		userID,
	)
	var trade sql.NullString
	err := row.Scan(&trade)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetOperativeTradeByUserID",
			Values:  []string{"userID"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperativeTradeByUserID",
			Values:  []string{"userID"},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrQuery,
		}
	}
	if trade.Valid {
		return &trade.String, nil
	}
	return nil, &appError.DBError{
		Context: "GetOperativeTradeByUserID",
		Values:  []string{"trade"},
		Action:  "query",
		Table:   "operative",
		Err:     appError.DBErrNotFound,
	}
}

func (as *AppStorage) GetOperatives() (*[]model.Operative, error) {
	rows, err := as.DB.Query(
		`select id, user_id, name, email, phone, trade from operative`,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperatives",
			Values:  []string{},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var operatives []model.Operative
	for rows.Next() {
		var o model.Operative
		var idStr, userIDStr string
		var phone, trade sql.NullString
		if err := rows.Scan(&idStr, &userIDStr, &o.Name, &o.Email, &phone, &trade); err != nil {
			return nil, &appError.DBError{
				Context: "GetOperatives",
				Values:  []string{},
				Action:  "scan",
				Table:   "operative",
				Err:     appError.DBErrQuery,
			}
		}
		o.ID, _ = uuid.Parse(idStr)
		o.UserID, _ = uuid.Parse(userIDStr)
		if phone.Valid {
			o.Phone = phone.String
		}
		if trade.Valid {
			o.Trade = &trade.String
		}
		operatives = append(operatives, o)
	}
	return &operatives, nil
}

func (as *AppStorage) GetOperativesWithRates() (*[]model.Operative, error) {
	rows, err := as.DB.Query(
		`select o.id, o.user_id, o.name, o.email, o.phone, o.trade, r.rate_per_hour from operative o inner join operative_rate r on r.operative_id = o.id`,
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperativesWithRates",
			Values:  []string{},
			Action:  "query",
			Table:   "operative",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var operatives []model.Operative
	for rows.Next() {
		var o model.Operative
		var idStr, userIDStr string
		var phone, trade, rate sql.NullString
		if err := rows.Scan(&idStr, &userIDStr, &o.Name, &o.Email, &phone, &trade, &rate); err != nil {
			return nil, &appError.DBError{
				Context: "GetOperativesWithRates",
				Values:  []string{"idStr", "userIDStr", "o.Name", "o.Email", "phone", "trade", "rate"},
				Action:  "scan",
				Table:   "operative",
				Err:     appError.DBErrQuery,
			}
		}
		o.ID, _ = uuid.Parse(idStr)
		o.UserID, _ = uuid.Parse(userIDStr)
		if phone.Valid {
			o.Phone = phone.String
		}
		if trade.Valid {
			o.Trade = &trade.String
		}
		if rate.Valid {
			rateAmount, err := model.MoneyFromString(rate.String)
			if err != nil {
				return nil, &appError.DBError{
					Context: "GetOperativeByUserID",
					Values:  []string{"rate"},
					Action:  "convert rate from string to money",
					Table:   "operative",
					Err:     appError.DBErrQuery,
				}
			}
			o.Rate = rateAmount
		}
		operatives = append(operatives, o)
	}
	return &operatives, nil
}

func (as *AppStorage) UpdateOperative(id *uuid.UUID, name, email string, phone, trade *string) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateOperative",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(
		`update operative set name = ?, email = ?, phone = ?, trade = ? where id = ?`,
		name, email, phone, trade, id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateOperative",
			Values:  []string{"name", "email", "phone", "trade"},
			Action:  "update",
			Table:   "operative",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) DeleteOperative(id *uuid.UUID) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "DeleteOperative",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "operative",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(`delete from operative where id = ?`, id.String())
	if err != nil {
		return 0, &appError.DBError{
			Context: "DeleteOperative",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "operative",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) NewOperativeRate(operativeID *uuid.UUID, ratePerHour model.Money, effectiveFrom time.Time) (*uuid.UUID, error) {
	if operativeID == nil {
		return nil, &appError.DBError{
			Context: "NewOperativeRate",
			Values:  []string{"operativeID"},
			Action:  "insert",
			Table:   "operative_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	id, idString := UUIDWithString()
	_, err := as.DB.Exec(
		`insert into operative_rate (id, operative_id, rate_per_hour, effective_from) values (?, ?, ?, ?)`,
		idString, operativeID.String(), ratePerHour.Serialize(), effectiveFrom.Unix(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "NewOperativeRate",
			Values:  []string{"id", "operative_id", "rate_per_hour", "effective_from"},
			Action:  "insert",
			Table:   "operative_rate",
			Err:     appError.DBErrInsert,
		}
	}
	return &id, nil
}

func (as *AppStorage) GetOperativeRate(operativeID *uuid.UUID) (*model.OperativeRate, error) {
	if operativeID == nil {
		return nil, &appError.DBError{
			Context: "GetOperativeRate",
			Values:  []string{"operativeID"},
			Action:  "query",
			Table:   "operative_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	row := as.DB.QueryRow(
		`select id, operative_id, rate_per_hour, effective_from, effective_to
		 from operative_rate
		 where operative_id = ? and (effective_to is null or effective_to > ?)
		 order by effective_from desc limit 1`,
		operativeID.String(), time.Now().Unix(),
	)
	var r model.OperativeRate
	var idStr, operativeIDStr, rateStr string
	var effectiveFrom float64
	var effectiveTo sql.NullFloat64
	err := row.Scan(&idStr, &operativeIDStr, &rateStr, &effectiveFrom, &effectiveTo)
	if err == sql.ErrNoRows {
		return nil, &appError.DBError{
			Context: "GetOperativeRate",
			Values:  []string{"operativeID"},
			Action:  "query",
			Table:   "operative_rate",
			Err:     appError.DBErrNotFound,
		}
	}
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperativeRate",
			Values:  []string{"operativeID"},
			Action:  "query",
			Table:   "operative_rate",
			Err:     appError.DBErrQuery,
		}
	}
	r.ID, _ = uuid.Parse(idStr)
	r.OperativeID, _ = uuid.Parse(operativeIDStr)
	r.EffectiveFrom = time.Unix(int64(effectiveFrom), 0)
	m, err := model.MoneyFromString(rateStr)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperativeRate",
			Values:  []string{"rate_per_hour"},
			Action:  "deserialize",
			Table:   "operative_rate",
			Err:     appError.DBErrSerialization,
		}
	}
	r.RatePerHour = *m
	if effectiveTo.Valid {
		t := time.Unix(int64(effectiveTo.Float64), 0)
		r.EffectiveTo = &t
	}
	return &r, nil
}

func (as *AppStorage) GetOperativesRates(operativeID *uuid.UUID) (*[]model.OperativeRate, error) {
	if operativeID == nil {
		return nil, &appError.DBError{
			Context: "GetOperativesRates",
			Values:  []string{"operativeID"},
			Action:  "query",
			Table:   "operative_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	rows, err := as.DB.Query(
		`select id, operative_id, rate_per_hour, effective_from, effective_to
		 from operative_rate where operative_id = ? order by effective_from desc`,
		operativeID.String(),
	)
	if err != nil {
		return nil, &appError.DBError{
			Context: "GetOperativesRates",
			Values:  []string{"operativeID"},
			Action:  "query",
			Table:   "operative_rate",
			Err:     appError.DBErrQuery,
		}
	}
	defer rows.Close()
	var rates []model.OperativeRate
	for rows.Next() {
		var r model.OperativeRate
		var idStr, operativeIDStr, rateStr string
		var effectiveFrom float64
		var effectiveTo sql.NullFloat64
		if err := rows.Scan(&idStr, &operativeIDStr, &rateStr, &effectiveFrom, &effectiveTo); err != nil {
			return nil, &appError.DBError{
				Context: "GetOperativesRates",
				Values:  []string{},
				Action:  "scan",
				Table:   "operative_rate",
				Err:     appError.DBErrQuery,
			}
		}
		r.ID, _ = uuid.Parse(idStr)
		r.OperativeID, _ = uuid.Parse(operativeIDStr)
		r.EffectiveFrom = time.Unix(int64(effectiveFrom), 0)
		m, err := model.MoneyFromString(rateStr)
		if err != nil {
			return nil, &appError.DBError{
				Context: "GetOperativesRates",
				Values:  []string{"rate_per_hour"},
				Action:  "deserialize",
				Table:   "operative_rate",
				Err:     appError.DBErrSerialization,
			}
		}
		r.RatePerHour = *m
		if effectiveTo.Valid {
			t := time.Unix(int64(effectiveTo.Float64), 0)
			r.EffectiveTo = &t
		}
		rates = append(rates, r)
	}
	return &rates, nil
}

func (as *AppStorage) UpdateOperativeRate(id *uuid.UUID, effectiveTo time.Time) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "UpdateOperativeRate",
			Values:  []string{"id"},
			Action:  "update",
			Table:   "operative_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(
		`update operative_rate set effective_to = ? where id = ?`,
		effectiveTo.Unix(), id.String(),
	)
	if err != nil {
		return 0, &appError.DBError{
			Context: "UpdateOperativeRate",
			Values:  []string{"effective_to"},
			Action:  "update",
			Table:   "operative_rate",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}

func (as *AppStorage) DeleteOperativeRate(id *uuid.UUID) (int64, error) {
	if id == nil {
		return 0, &appError.DBError{
			Context: "DeleteOperativeRate",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "operative_rate",
			Err:     appError.DBErrBadArgument,
		}
	}
	result, err := as.DB.Exec(`delete from operative_rate where id = ?`, id.String())
	if err != nil {
		return 0, &appError.DBError{
			Context: "DeleteOperativeRate",
			Values:  []string{"id"},
			Action:  "delete",
			Table:   "operative_rate",
			Err:     appError.DBErrInsert,
		}
	}
	return result.RowsAffected()
}
