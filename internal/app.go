package internal

import (
	"log"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/db"
)

const (
	SerConstAccessTokenExpiry  = 60
	SerConstRefreshTokenExpiry = 168
)

type AppConfig struct {
	JWTSecret    string
	SecureCookie bool
}

type JMcCannBackPackApp struct {
	ErrorLog *log.Logger
	InfoLog  *log.Logger
	DB       *db.AppStorage
	Config   AppConfig
	// TemplateCache *ui.TemplateCache
}
