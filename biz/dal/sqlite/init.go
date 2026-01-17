package sqlite

import (
	"fmt"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/glebarez/sqlite" // 这是基于modernc.org/sqlite的纯Go GORM驱动
)

var DB *gorm.DB

func Init(Database string, gormLogger logger.Interface) *gorm.DB {
	// 定义数据库文件的路径
	directory := "data/db"
	dbFile := fmt.Sprintf("%s/%s.db", directory, Database)

	// 检查文件夹是否存在，如果不存在则创建
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// 创建文件夹
		if err := os.MkdirAll(directory, os.ModePerm); err != nil {
			panic(fmt.Sprintf("无法创建文件夹: %v", err))
		}
	}

	// 打开 SQLite 数据库
	var err error
	DB, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
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

	return DB
}
