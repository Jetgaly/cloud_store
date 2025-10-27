package core

import (
	"cloud_store/global"
	"cloud_store/utils"
)

func InitEmail(){
	global.EmailSender = &utils.EmailSender{}
	global.EmailSender.InitGomail(global.Config.Email.Host,global.Config.Email.Port,global.Config.Email.User,global.Config.Email.Password)
}