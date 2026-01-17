package dal

import (
	"fmt"
)

// PermissionService 权限服务
var PermissionService = &permissionService{}

type permissionService struct{}

// CheckNamespacePermission 检查用户对指定命名空间的权限
// username: 用户名
// namespace: 命名空间ID
// action: 操作类型，支持 "r" (读取) 或 "rw" (读写)
// 返回值: 是否有权限，错误信息
func (s *permissionService) CheckNamespacePermission(username string, namespace string, action string) (bool, error) {
	// 获取用户的所有角色
	roles, err := GetUserRoles(username)
	if err != nil {
		return false, fmt.Errorf("查询用户角色失败: %v", err)
	}

	if len(roles) == 0 {
		return false, nil // 用户没有任何角色
	}

	// 根据操作类型检查权限
	hasPermission, err := HasNamespacePermission(username, namespace, action)
	if err != nil {
		return false, fmt.Errorf("查询权限失败: %v", err)
	}

	return hasPermission, nil
}

// GetUserRoles 获取用户的所有角色
func (s *permissionService) GetUserRoles(username string) ([]string, error) {
	return GetUserRoles(username)
}

// GetUserNamespacePermissions 获取用户对命名空间的所有权限
func (s *permissionService) GetUserNamespacePermissions(username string, namespace string) ([]string, error) {
	roles, err := s.GetUserRoles(username)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return []string{}, nil
	}

	permissions, err := GetUserNamespacePermissions(username, namespace)
	if err != nil {
		return nil, fmt.Errorf("查询权限失败: %v", err)
	}

	actions := make([]string, len(permissions))
	for i, perm := range permissions {
		actions[i] = perm.Action
	}

	return actions, nil
}

// CheckNamespaceReadPermission 检查用户是否有命名空间的读取权限
func (s *permissionService) CheckNamespaceReadPermission(username string, namespace string) (bool, error) {
	return s.CheckNamespacePermission(username, namespace, "r")
}

// CheckNamespaceWritePermission 检查用户是否有命名空间的写权限（rw或w）
func (s *permissionService) CheckNamespaceWritePermission(username string, namespace string) (bool, error) {
	// 先检查rw权限
	rwPermission, err := s.CheckNamespacePermission(username, namespace, "rw")
	if err != nil {
		return false, err
	}
	if rwPermission {
		return true, nil
	}
	// 再检查w权限
	return s.CheckNamespacePermission(username, namespace, "w")
}
