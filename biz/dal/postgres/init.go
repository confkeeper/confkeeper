package postgres

import (
	"fmt"

	"confkeeper/utils/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

func Init(cfg *config.DbConfig, zone string, gormLogger logger.Interface) *gorm.DB {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, zone)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
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
		slaveUser := cfg.SlaveUser
		if slaveUser == "" {
			slaveUser = cfg.User
		}
		slavePassword := cfg.SlavePassword
		if slavePassword == "" {
			slavePassword = cfg.Password
		}
		slavePort := cfg.SlavePort
		if slavePort == "" {
			slavePort = cfg.Port
		}
		slaveDsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=%s",
			slaveUser, slavePassword, cfg.SlaveHost, slavePort, cfg.Database, zone)
		err = DB.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(slaveDsn)},
		}))
		if err != nil {
			panic(err)
		}
	}

	return DB
}
