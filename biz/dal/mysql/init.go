package mysql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

func Init(dbUser string, dbPassword string, dbHost string, dbPort string, dbName string, zone string, gormLogger logger.Interface, slaveHost string, slavePort string, slaveUser string, slavePassword string) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, zone)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(err)
	}

	if slaveHost != "" {
		slaveDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=%s",
			slaveUser, slavePassword, slaveHost, slavePort, dbName, zone)
		err = DB.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{mysql.Open(slaveDsn)},
		}))
		if err != nil {
			panic(err)
		}
	}

	return DB
}
