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
	if err := DB.Model(&model.Roles{}).Where("role IN ?", uniqueRoles).Order("role").Find(&roles).Error; err != nil {
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
		newRoles := make([]*model.Roles, 0, len(usernames))
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

// FindMissingUsernames 检查一个用户名列表，并返回那些在数据库中不存在的用户名。
func FindMissingUsernames(usernames []string) ([]string, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	var existingUsernames []string
	// 找出给定的用户名中有哪些已经存在。
	err := DB.Model(&model.User{}).Where("username IN ?", usernames).Pluck("username", &existingUsernames).Error
	if err != nil {
		return nil, err
	}

	// 使用 map 来高效查找已存在的用户名。
	existingSet := make(map[string]struct{}, len(existingUsernames))
	for _, u := range existingUsernames {
		existingSet[u] = struct{}{}
	}

	// 找出哪些用户名是缺失的。
	var missingUsernames []string
	for _, u := range usernames {
		if _, ok := existingSet[u]; !ok {
			missingUsernames = append(missingUsernames, u)
		}
	}
	return missingUsernames, nil
}
