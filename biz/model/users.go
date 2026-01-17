package model

type User struct {
	ID       uint   `gorm:"primarykey;comment:主键ID" json:"id"`
	Username string `gorm:"unique;column:username;type:varchar(255);comment:用户名" json:"username"`
	Password string `gorm:"column:password;type:varchar(255);comment:密码" json:"password"`
	Enable   bool   `gorm:"column:enabled;type:boolean;comment:是否启用" json:"enable"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) TableComment() string {
	return "用户表"
}
