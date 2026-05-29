package internal

import (
	"log"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/db"
	"adediiji.uk/jmcann-suffolk-backpack-task/internal/ui"
)

const (
	SerConstAccessTokenExpiryMinutes = 60
	SerConstRefreshTokenExpiryHours  = 720
)

type AppConfig struct {
	JWTSecret    string
	SecureCookie bool
}

type JMcCannBackPackApp struct {
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	DB            *db.AppStorage
	Config        AppConfig
	TemplateCache *ui.TemplateCache
}
