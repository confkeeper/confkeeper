package dal

import (
	"confkeeper/biz/model"
)

// AddRolePermission 为角色添加权限
func AddRolePermission(role, resource, action string) error {
	permission := &model.Permissions{
		Role:     role,
		Resource: resource,
		Action:   action,
	}
	return DB.Create(permission).Error
}

// RemoveRolePermission 移除角色的权限
func RemoveRolePermission(role, resource, action string) error {
	return DB.Where("role = ? AND resource = ? AND action = ?", role, resource, action).
		Delete(&model.Permissions{}).Error
}

// IsPermissionExists 检查权限是否存在
func IsPermissionExists(role, resource, action string) (bool, error) {
	var count int64
	err := DB.Model(&model.Permissions{}).
		Where("role = ? AND resource = ? AND action = ?", role, resource, action).
		Count(&count).Error
	return count > 0, err
}

// GetRolePermissionsList 分页获取角色的权限列表
func GetRolePermissionsList(role string, offset, pageSize int) ([]*model.Permissions, int64, error) {
	var permissions []*model.Permissions
	var total int64

	query := DB.Model(&model.Permissions{})
	if role != "" {
		query = query.Where("role LIKE ?", "%"+role+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(pageSize).Find(&permissions).Error
	return permissions, total, err
}

// GetUserRoles 获取用户的所有角色
func GetUserRoles(username string) ([]string, error) {
	var roles []model.Roles
	err := DB.Where("username = ?", username).Find(&roles).Error
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Role
	}

	return roleNames, nil
}

// GetUserNamespacePermissions 获取用户对指定命名空间的所有权限
func GetUserNamespacePermissions(username string, namespace string) ([]*model.Permissions, error) {
	roles, err := GetUserRoles(username)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return []*model.Permissions{}, nil
	}

	var permissions []*model.Permissions
	err = DB.Where("role IN (?) AND resource = ?", roles, namespace).Find(&permissions).Error
	return permissions, err
}

// HasNamespacePermission 检查用户是否有指定命名空间的权限
func HasNamespacePermission(username string, namespace string, action string) (bool, error) {
	roles, err := GetUserRoles(username)
	if err != nil {
		return false, err
	}

	if len(roles) == 0 {
		return false, nil
	}

	var count int64
	err = DB.Model(&model.Permissions{}).
		Where("role IN (?) AND resource = ? AND action = ?", roles, namespace, action).
		Count(&count).Error
	return count > 0, err
}
