package dal

import (
	"confkeeper/biz/model"

	"gorm.io/gorm"
)

// CreateRole 为用户添加角色
func CreateRole(roles []*model.Roles) error {
	return DB.Create(roles).Error
}

// IsRoleExistsInRoles 检查角色是否存在（基于角色表）
func IsRoleExistsInRoles(role string) (bool, error) {
	var count int64
	err := DB.Model(&model.Roles{}).
		Where("role = ?", role).
		Count(&count).Error
	return count > 0, err
}

// DeleteRole 删除角色及其所有权限
func DeleteRole(role string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 删除角色的所有权限
		if err := tx.Where("role = ?", role).Delete(&model.Permissions{}).Error; err != nil {
			return err
		}

		// 删除用户的该角色
		if err := tx.Where("role = ?", role).Delete(&model.Roles{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// GetAllRolesWithPagination 分页获取所有角色列表（基于角色表）
func GetAllRolesWithPagination(pageSize int, offset int) ([]*model.Roles, int64, error) {
	var roles []*model.Roles

	query := DB.Model(&model.Roles{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	if err := query.Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// GetRoleListWithPagination 分页获取角色的聚合列表
func GetRoleListWithPagination(pageSize int, offset int) ([]*model.Roles, int64, error) {
	var total int64

	// 获取不重复的角色总数
	if err := DB.Model(&model.Roles{}).Distinct("role").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页获取不重复的角色名
	var uniqueRoles []string
if err := DB.Model(&model.Roles{}).Distinct("role").Order("role").Offset(offset).Limit(pageSize).Pluck("role", &uniqueRoles).Error; err != nil {
		return nil, 0, err
	}

	if len(uniqueRoles) == 0 {
		return nil, total, nil
	}

	// 获取这些角色下的所有记录
	var roles []*model.Roles
	if err := DB.Model(&model.Roles{}).Where("role IN ?", uniqueRoles).Find(&roles).Error; err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// UpdateRoleUsers 更新角色的用户列表（全量替换）
func UpdateRoleUsers(role string, usernames []string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 删除属于该角色的所有用户关系
		if err := tx.Where("role = ?", role).Delete(&model.Roles{}).Error; err != nil {
			return err
		}

		if len(usernames) == 0 {
			return nil
		}

		// 组装新的绑定关系
		var newRoles []*model.Roles
		for _, username := range usernames {
			newRoles = append(newRoles, &model.Roles{
				Role:     role,
				Username: username,
			})
		}

		// 批量插入
		if err := tx.Create(newRoles).Error; err != nil {
			return err
		}

		return nil
	})
}
