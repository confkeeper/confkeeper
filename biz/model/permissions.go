package model

type Permissions struct {
	Role     string `gorm:"type:varchar(50);not null;index;comment:角色名;uniqueIndex:idx_role_resource_action" json:"role"`
	Resource string `gorm:"type:varchar(255);not null;comment:资源路径;uniqueIndex:idx_role_resource_action" json:"resource"`
	Action   string `gorm:"type:varchar(8);not null;comment:操作权限;uniqueIndex:idx_role_resource_action" json:"action"`
}

func (permissions *Permissions) TableName() string {
	return "permissions"
}

func (permissions *Permissions) TableComment() string {
	return "命名空间对应用户组的权限"
}

//`role` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名',
//`resource` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资源路径',
//`action` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '操作权限',
