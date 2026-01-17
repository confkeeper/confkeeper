package dal

import (
	"confkeeper/biz/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func CreateTenant(Tenants []*model.TenantInfo) error {
	return DB.Create(Tenants).Error
}

func IsTenantIdExists(tenantId string) (bool, error) {
	var count int64
	err := DB.Model(&model.TenantInfo{}).Where("tenant_id = ?", tenantId).Count(&count).Error
	return count > 0, err
}

func DeleteTenant(TenantID uint) error {
	var Tenant model.TenantInfo
	if err := DB.First(&Tenant, "id = ?", TenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("命名空间不存在或已被删除")
		}
		return err
	}
	return DB.Delete(&Tenant).Error
}

func GetTenantList(pageSize, offset int) ([]*model.TenantInfo, int64, error) {
	var Tenants []*model.TenantInfo
	query := DB.Model(&model.TenantInfo{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id").Limit(pageSize).Offset(offset).Find(&Tenants).Error
	return Tenants, total, err
}

func GetTenantById(id uint) (*model.TenantInfo, error) {
	var tenant model.TenantInfo
	if err := DB.First(&tenant, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 租户不存在时返回 nil
		}
		return nil, err // 其他错误
	}
	return &tenant, nil
}

// GetTenantListByIDs 根据命名空间ID列表获取租户列表
func GetTenantListByIDs(tenantIDs []string, pageSize, offset int) ([]*model.TenantInfo, int64, error) {
	var tenants []*model.TenantInfo
	query := DB.Model(&model.TenantInfo{}).Where("tenant_id IN (?)", tenantIDs)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("id").Limit(pageSize).Offset(offset).Find(&tenants).Error
	return tenants, total, err
}
