package api

import "cloud_store/api/user"

type ApiHandler struct {
	UserApi user.UserApi
}

var Handler ApiHandler
