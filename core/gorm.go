package core

import (
	"cloud_store/global"
	"cloud_store/model"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitGorm() {
	var mysqlLogger logger.Interface
	if global.Config.Mysql.LogLevel == "dev" {
		//dev模式显示所有sql
		mysqlLogger = logger.Default.LogMode(logger.Info)
	} else {
		//只打印错误sql
		mysqlLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(mysql.Open(global.Config.Mysql.Dsn()), &gorm.Config{
		Logger: mysqlLogger,
	})
	if err != nil {
		global.Logger.Fatal("gorm init err : " + err.Error())
	}
	global.Logger.Info("gorm init successfully")

	sqldb, _ := db.DB()
	sqldb.SetMaxIdleConns(10)  //最大空闲连接数
	sqldb.SetMaxOpenConns(100) //最多可容纳连接数
	global.DB = db
}

func CreateTables() {
	global.DB.SetupJoinTable(&model.User{}, "Files", &model.UserFile{})
	err := global.DB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		&model.User{},
		&model.File{},
		&model.UserFile{},
		&model.VolumeOpLog{},
	)
	if err != nil {
		global.Logger.Fatal(fmt.Sprintf("gorm create tables err: %s", err.Error()))
	}
}
