package postgres

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var DB *gorm.DB

func Init(dbUser string, dbPassword string, dbHost string, dbPort string, dbName string, zone string, gormLogger logger.Interface, slaveHost string, slavePort string, slaveUser string, slavePassword string) *gorm.DB {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, zone)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
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
		slaveDsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable TimeZone=%s",
			slaveUser, slavePassword, slaveHost, slavePort, dbName, zone)
		err = DB.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(slaveDsn)},
		}))
		if err != nil {
			panic(err)
		}
	}

	return DB
}
