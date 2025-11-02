package api

import (
	"cloud_store/api/file"
	"cloud_store/api/user"
)

type ApiHandler struct {
	UserApi user.UserApi
	FileApi file.FileApi
}

var Handler ApiHandler
