package model

type TenantInfo struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	TenantID   string `gorm:"type:varchar(128);default:'';comment:命名空间ID" json:"tenant_id"`
	TenantName string `gorm:"type:varchar(128);default:'';comment:命名空间名称" json:"tenant_name"`
	TenantDesc string `gorm:"type:varchar(256);comment:命名空间描述" json:"tenant_desc"`
}

func (tenant *TenantInfo) TableName() string {
	return "tenant_info"
}

func (tenant *TenantInfo) TableComment() string {
	return "命名空间表"
}

//`id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
//`tenant_id` varchar(128) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NULL DEFAULT '' COMMENT '命名空间ID',
//`tenant_name` varchar(128) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NULL DEFAULT '' COMMENT '命名空间名称',
//`tenant_desc` varchar(256) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin NULL DEFAULT NULL COMMENT '命名空间描述',
