package core

func init() {
	InitConf()
	InitLogger()
	InitEmail()
	InitGorm()
	CreateTables()
	InitRedis()
}
