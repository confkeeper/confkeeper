package model

type Roles struct {
	Username string `gorm:"type:varchar(50);not null;comment:用户名;uniqueIndex:idx_user_role" json:"username"`
	Role     string `gorm:"type:varchar(50);not null;comment:角色名;uniqueIndex:idx_user_role" json:"role"`
}

func (roles *Roles) TableName() string {
	return "roles"
}

func (roles *Roles) TableComment() string {
	return "用户对应用户组"
}

//`username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户名',
//`role` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名',
