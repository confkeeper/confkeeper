package mysql

import (
	"fmt"

	"confkeeper/utils/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

func Init(cfg *config.DbConfig, zone string, gormLogger logger.Interface) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, zone)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		panic(err)
	}

	if cfg.SlaveHost != "" {
		slaveDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=%s",
			cfg.SlaveUser, cfg.SlavePassword, cfg.SlaveHost, cfg.SlavePort, cfg.Database, zone)
		err = DB.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{mysql.Open(slaveDsn)},
		}))
		if err != nil {
			panic(err)
		}
	}

	return DB
}
