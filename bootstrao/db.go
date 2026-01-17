package bootstrao

import (
	"confkeeper/biz/model"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// 自动迁移表结构
	if err := db.AutoMigrate(
		&model.User{},
		&model.ConfigInfo{},
		&model.TenantInfo{},
		&model.Roles{},
		&model.Permissions{},
	); err != nil {
		return err
	}

	err := InitData(db)
	if err != nil {
		return err
	}

	return nil
}
