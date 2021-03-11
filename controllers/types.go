package controllers

import (
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type DatabaseBackupsController struct {
	databaseIDs []string
	scwClient   *scw.Client
}
