package mw

import (
	"confkeeper/utils"
	"errors"

	"confkeeper/biz/dal"

	"github.com/gin-gonic/gin"
)

var ErrUnauthorized = errors.New("unauthorized")

// CheckNamespaceReadOrWritePermissionHTTP 检查用户是否有命名空间的读取或读写权限
func CheckNamespaceReadOrWritePermissionHTTP(c *gin.Context, namespace string) (bool, error) {
	username, err := utils.GetUsernameFromContext(c)
	if err != nil || username == "" {
		return false, ErrUnauthorized
	}

	// 先检查是否有读权限
	readPermission, err := dal.PermissionService.CheckNamespaceReadPermission(username, namespace)
	if err != nil {
		return false, err
	}
	if readPermission {
		return true, nil
	}

	// 再检查是否有读写权限（只检查rw，不检查w权限）
	rwPermission, err := dal.PermissionService.CheckNamespacePermission(username, namespace, "rw")
	if err != nil {
		return false, err
	}
	return rwPermission, nil
}

// CheckNamespaceWritePermissionHTTP 检查用户是否有命名空间的读写权限
func CheckNamespaceWritePermissionHTTP(c *gin.Context, namespace string) (bool, error) {
	username, err := utils.GetUsernameFromContext(c)
	if err != nil || username == "" {
		return false, ErrUnauthorized
	}

	return dal.PermissionService.CheckNamespaceWritePermission(username, namespace)
}
