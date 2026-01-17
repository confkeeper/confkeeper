package model

import "time"

type ConfigInfo struct {
	ID         uint      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	DataID     string    `gorm:"type:varchar(255);not null;comment:配置ID;uniqueIndex:idx_data_group_version" json:"data_id"`
	GroupID    string    `gorm:"type:varchar(255);comment:分组ID;uniqueIndex:idx_data_group_version" json:"group_id"`
	Content    string    `gorm:"type:text;not null;comment:配置内容" json:"content"`
	TenantID   string    `gorm:"type:varchar(128);default:'';comment:命名空间ID;uniqueIndex:idx_data_group_version" json:"tenant_id"`
	Type       string    `gorm:"type:varchar(64);comment:配置类型" json:"type"`
	Version    int       `gorm:"type:int;not null;default:1;comment:版本号;uniqueIndex:idx_data_group_version" json:"version"`
	Author     string    `gorm:"type:varchar(255);default:'';comment:修改人" json:"author"`
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP" json:"create_time"`
}

func (cfg *ConfigInfo) TableName() string {
	return "config_info"
}

func (cfg *ConfigInfo) TableComment() string {
	return "配置表"
}

//`id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
//`data_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL COMMENT '配置ID',
//`group_id` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NULL DEFAULT NULL COMMENT '分组ID',
//`content` longtext CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NOT NULL COMMENT '配置内容',
//`tenant_id` varchar(128) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NULL DEFAULT '' COMMENT '租户ID',
//`type` varchar(64) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NULL DEFAULT NULL COMMENT '配置类型',
//`version` int NOT NULL DEFAULT 1 COMMENT '版本号',
//`author` varchar(255) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NULL DEFAULT '' COMMENT '修改人',
// 联合唯一键: (data_id, group_id, version, tenant_id)
