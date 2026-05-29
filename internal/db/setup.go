package db

import (
	"database/sql"
	"fmt"

	appError "adediiji.uk/jmcann-suffolk-backpack-task/internal/error"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const initStatment string = `
create table if not exists user (
  id text not null,
  name text not null,
  email text not null,
  role text not null,
  password_hash text not null,
  primary key(id)
);
create table if not exists site (
  id text not null,
  name text not null,
  address text not null,
  primary key(id)
);
create table if not exists job (
  id text not null,
  reference text not null,
  name text not null,
  site_id text not null,
  created_by text not null,
  status text not null,
  start_datetime real not null,
  end_datetime real,
  total_cost text,
  primary key(id),
  foreign key (created_by) references user(id)
  on update cascade on delete restrict,
  foreign key (site_id) references site(id)
  on update cascade on delete restrict
);
create table if not exists resource (
  id text not null,
  name text not null,
  resource_type text not null,
  unit_of_measure text,
  primary key(id)
);
create table if not exists resource_rate (
  id text not null,
  resource_id text not null,
  rate text not null,
  cost_unit text not null,
  effective_from real not null,
  effective_to real,
  primary key(id),
  foreign key (resource_id) references resource(id)
  on update cascade on delete restrict
);
create table if not exists operative (
  id text not null,
  user_id text not null,
  name text not null,
  email text not null,
  phone text,
  trade text,
  primary key(id),
  foreign key (user_id) references user(id)
  on update cascade on delete restrict
);
create table if not exists operative_rate (
  id text not null,
  operative_id text not null,
  rate_per_hour text not null,
  effective_from real not null,
  effective_to real,
  primary key(id),
  foreign key (operative_id) references operative(id)
  on update cascade on delete restrict
);
create table if not exists job_session (
  id text not null,
  job_id text not null,
  session_date real not null default (unixepoch('now')),
  start_time real not null default (unixepoch('now')),
  end_time real,
  submitted_at real,
  submitted_by text,
  notes text,
  primary key(id),
  foreign key (job_id) references job(id)
  on update cascade on delete restrict,
  foreign key (submitted_by) references operative(id)
  on update cascade on delete restrict
);
create table if not exists job_operative (
  id text not null,
  job_id text not null,
  session_id text not null,
  operative_id text not null,
  arrival_time real not null,
  departure_time real,
  rate_snapshot text not null,
  calculated_cost text,
  primary key(id),
  foreign key (job_id) references job(id)
  on update cascade on delete restrict,
  foreign key (session_id) references job_session(id)
  on update cascade on delete restrict,
  foreign key (operative_id) references operative(id)
  on update cascade on delete restrict
);
create table if not exists job_resource (
    id text not null,
    job_id text not null,
    session_id text not null,
    resource_id text not null,
    quantity real,
    duration_hours real,
    arrival_time real,
    rate_snapshot text not null,
    calculated_cost text,
    primary key(id),
    foreign key (job_id) references job(id)
    on update cascade on delete restrict,
    foreign key (session_id) references job_session(id)
    on update cascade on delete restrict,
    foreign key (resource_id) references resource(id)
    on update cascade on delete restrict
);
create table if not exists job_resource_requirement (
  id text not null,
  job_id text not null,
  resource_id text not null,
  expected_quantity integer default 1,
  expected_duration_hours real,
  notes text,
  primary key(id),
  foreign key (job_id) references job(id)
  on update cascade on delete cascade,
  foreign key (resource_id) references resource(id)
  on update cascade on delete restrict
);
create table if not exists job_operative_requirement (
  id text not null,
  job_id text not null,
  expected_headcount integer default 1,
  notes text,
  primary key(id),
  foreign key (job_id) references job(id)
  on update cascade on delete cascade
);
`

type AppStorage struct {
	DB *sql.DB
}

func OpenDBConnection(db_path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", db_path))
	if err != nil {
		return nil, &appError.DBError{
			Context: "OpenDBConnection",
			Values:  []string{},
			Action:  "open connection",
			Table:   "",
			Err:     err,
		}
	}

	return db, nil
}

func SetupDB(db *sql.DB) error {
	_, err := db.Exec(initStatment)
	if err != nil {
		return &appError.DBError{
			Context: "SetupDB",
			Values:  []string{},
			Action:  "run create tables migration",
			Table:   "",
			Err:     err,
		}
	}
	return nil
}

func UUIDWithString() (uuid.UUID, string) {
	uuidd := uuid.New()

	return uuidd, uuidd.String()
}
