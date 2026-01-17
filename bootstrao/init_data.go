package bootstrao

import (
	"confkeeper/biz/model"
	"confkeeper/utils"
	"confkeeper/utils/config"

	"github.com/gookit/slog"
	"gorm.io/gorm"
)

func InitData(db *gorm.DB) error {
	// 创建管理员用户
	adminUser := &model.User{
		Username: config.Cfg.Admin.Username,
		Password: utils.MD5(config.Cfg.Admin.Password),
		Enable:   true,
	}

	userResult := db.Where(model.User{Username: config.Cfg.Admin.Username}).FirstOrCreate(adminUser)
	if userResult.Error != nil {
		return userResult.Error
	}

	if userResult.RowsAffected > 0 {
		slog.Infof("创建管理员用户成功，用户名: %s, 密码: %s", config.Cfg.Admin.Username, config.Cfg.Admin.Password)

		// 只有在创建了用户的情况下才创建默认命名空间
		defaultTenant := &model.TenantInfo{
			TenantID:   "default",
			TenantName: "default",
			TenantDesc: "default",
		}

		// 先检查是否已存在默认命名空间
		tenantResult := db.Where(model.TenantInfo{TenantID: "default"}).FirstOrCreate(defaultTenant)
		if tenantResult.Error != nil {
			return tenantResult.Error
		}

	} else {
		slog.Infof("管理员用户已存在: %s", config.Cfg.Admin.Username)
	}

	return nil
}
