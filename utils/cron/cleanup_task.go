package cron

import (
	"confkeeper/biz/dal"
	"confkeeper/bootstrao"
	"confkeeper/utils/config"
	"strings"
	"time"

	"github.com/gookit/slog"
	"github.com/robfig/cron/v3"

	"gorm.io/gorm"
)

// CleanupTask 数据库清理任务
func CleanupTask() {
	c := cron.New(cron.WithSeconds())
	_, err := c.AddFunc(config.Cfg.Server.DeleteDataCron, func() {
		performCleanup()
		if err := bootstrao.Migrate(dal.DB); err != nil {
			slog.Errorf("初始化数据失败: %v", err)
		}
	})
	if err != nil {
		slog.Errorf("添加定时任务失败: %v", err)
		return
	}

	c.Start()
	slog.Info("CleanupTask 定时任务已启动")

	// 阻塞主线程，防止退出
	select {}
}

// performCleanup 执行数据库清理
func performCleanup() {
	start := time.Now()

	dbType := config.Cfg.Db.Type

	err := dal.DB.Transaction(func(tx *gorm.DB) error {
		// 获取所有表
		tables, err := tx.Migrator().GetTables()
		if err != nil {
			return err
		}

		var filteredTables []string

		switch dbType {
		case "sqlite3":
			for _, t := range tables {
				if t != "sqlite_sequence" { // 过滤掉系统表
					filteredTables = append(filteredTables, t)
				}
			}

		case "postgres":
			for _, t := range tables {
				// Postgres 系统表都在 pg_catalog 和 information_schema，gorm.Migrator() 默认不会返回这些
				filteredTables = append(filteredTables, t)
			}

		case "mysql":
			for _, t := range tables {
				// 确保不是系统库的表
				if !(strings.HasPrefix(t, "mysql.") ||
					strings.HasPrefix(t, "sys.") ||
					strings.HasPrefix(t, "performance_schema.") ||
					strings.HasPrefix(t, "information_schema.")) {
					filteredTables = append(filteredTables, t)
				}
			}
		}

		// 删除过滤后的表
		for _, table := range filteredTables {
			if err := tx.Migrator().DropTable(table); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		slog.Errorf("数据库清理失败: %v", err)
	} else {
		elapsed := time.Since(start)
		slog.Infof("数据库清理完成，耗时: %v", elapsed)
	}
}
