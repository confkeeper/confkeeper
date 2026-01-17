package dal

import (
	"confkeeper/biz/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func CreateConfigInfo(configInfo []*model.ConfigInfo) error {
	for _, info := range configInfo {
		exists, err := IsConfigInfoExistsWithTenant(info.DataID, info.GroupID, info.TenantID, info.Version)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("配置已存在: data_id=%s, group_id=%s", info.DataID, info.GroupID)
		}
	}
	return DB.Create(&configInfo).Error
}

func GetConfigInfoByID(ConfigInfoID string) (*model.ConfigInfo, error) {
	var ConfigInfo model.ConfigInfo
	if err := DB.Where("id = ?", ConfigInfoID).First(&ConfigInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在时返回 nil
		}
		return nil, err // 其他错误
	}
	return &ConfigInfo, nil
}

func IsConfigInfoExists(dataId string, groupId string, tenantId string) (bool, error) {
	var count int64
	err := DB.Model(&model.ConfigInfo{}).Where("data_id = ?", dataId).Where("group_id = ?", groupId).Where("tenant_id = ?", tenantId).Count(&count).Error
	return count > 0, err
}

func IsConfigInfoExistsWithTenant(dataId string, groupId string, tenantId string, version int) (bool, error) {
	var count int64
	err := DB.Model(&model.ConfigInfo{}).Where("data_id = ?", dataId).Where("group_id = ?", groupId).Where("tenant_id = ?", tenantId).Where("version = ?", version).Count(&count).Error
	return count > 0, err
}

// GetMaxVersionByDataIdGroupAndTenant 获取指定data_id、group_id、tenant_id的最大版本号
func GetMaxVersionByDataIdGroupAndTenant(dataId string, groupId string, tenantId string) (int, error) {
	var maxVersion int
	err := DB.Model(&model.ConfigInfo{}).
		Where("data_id = ? AND group_id = ? AND tenant_id = ?", dataId, groupId, tenantId).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	return maxVersion, err
}

// GetConfigInfoByDataIdAndGroupWithMaxVersion 获取指定data_id和group_id的最大版本配置
func GetConfigInfoByDataIdAndGroupWithMaxVersion(dataId string, groupId string, tenantId string) (*model.ConfigInfo, error) {
	var configInfo model.ConfigInfo
	subQuery := DB.Model(&model.ConfigInfo{}).
		Where("data_id = ? AND group_id = ? AND tenant_id = ?", dataId, groupId, tenantId).
		Select("MAX(version)")

	err := DB.Where("data_id = ? AND group_id = ? AND tenant_id = ? AND version = (?)", dataId, groupId, tenantId, subQuery).
		First(&configInfo).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 配置不存在时返回 nil
		}
		return nil, err // 其他错误
	}
	return &configInfo, nil
}

// DeleteConfigInfoByDataIdAndGroup 根据data_id和group_id删除所有版本配置
func DeleteConfigInfoByDataIdAndGroup(dataId string, groupId string) error {
	return DB.Where("data_id = ? AND group_id = ?", dataId, groupId).
		Delete(&model.ConfigInfo{}).Error
}

// GetConfigInfoListWithMaxVersion 获取配置列表，只返回每个data_id和group_id组合的最大版本
func GetConfigInfoListWithMaxVersion(pageSize, offset int, dataId, groupId, Type, tenantId string) ([]*model.ConfigInfo, int64, error) {
	var configInfos []*model.ConfigInfo

	// where 条件组装
	conditions := "tenant_id = ?"
	args := []interface{}{tenantId}

	if dataId != "" {
		conditions += " AND data_id LIKE ?"
		args = append(args, "%"+dataId+"%")
	}
	if groupId != "" {
		conditions += " AND group_id LIKE ?"
		args = append(args, "%"+groupId+"%")
	}
	if Type != "" {
		conditions += " AND type = ?"
		args = append(args, Type)
	}

	// 子查询1：排序用（version=1 的 id）
	sorterSubQuery := DB.
		Table("config_info").
		Select("data_id, group_id, MIN(id) AS base_id").
		Where("version = 1 AND "+conditions, args...).
		Group("data_id, group_id")

	// 子查询2：取最大 version
	latestVersionSubQuery := DB.
		Table("config_info").
		Select("data_id, group_id, MAX(version) AS max_version").
		Where(conditions, args...).
		Group("data_id, group_id")

	// 主查询
	mainQuery := DB.
		Table("config_info AS ci").
		Select("ci.*").
		Joins("JOIN (?) AS lv ON lv.data_id = ci.data_id AND lv.group_id = ci.group_id AND lv.max_version = ci.version", latestVersionSubQuery).
		Joins("JOIN (?) AS sorter ON sorter.data_id = ci.data_id AND sorter.group_id = ci.group_id", sorterSubQuery).
		Where("ci.tenant_id = ?", tenantId).
		Order("sorter.base_id ASC").
		Limit(pageSize).
		Offset(offset)

	// 计算 total
	var total int64
	countQuery := DB.
		Table("(?) AS sub", latestVersionSubQuery).
		Count(&total)
	if countQuery.Error != nil {
		return nil, 0, countQuery.Error
	}

	// 查询结果
	if err := mainQuery.Find(&configInfos).Error; err != nil {
		return nil, 0, err
	}

	// 👇 保证返回的条数和 total 对齐
	if int64(len(configInfos)) < total {
		// 补齐，直接重新从 latestVersionSubQuery 拉数据
		var allConfigs []*model.ConfigInfo
		err := DB.
			Table("config_info AS ci").
			Select("ci.*").
			Joins("JOIN (?) AS lv ON lv.data_id = ci.data_id AND lv.group_id = ci.group_id AND lv.max_version = ci.version", latestVersionSubQuery).
			Where("ci.tenant_id = ?", tenantId).
			Order("ci.id ASC").
			Limit(pageSize).
			Offset(offset).
			Find(&allConfigs).Error
		if err != nil {
			return nil, 0, err
		}
		return allConfigs, total, nil
	}

	return configInfos, total, nil
}

// GetAllVersionsByDataIdAndGroup 根据data_id和group_id查询所有版本，按版本倒序返回
func GetAllVersionsByDataIdAndGroup(dataId string, groupId string) ([]*model.ConfigInfo, error) {
	var configInfos []*model.ConfigInfo
	err := DB.Model(&model.ConfigInfo{}).
		Where("data_id = ? AND group_id = ?", dataId, groupId).
		Order("version DESC").
		Find(&configInfos).Error
	return configInfos, err
}

// GetConfigInfoByDataIdGroupAndTenant 根据data_id、group_id和tenant_id查询配置
func GetConfigInfoByDataIdGroupAndTenant(dataId string, groupId string, tenantId string) (*model.ConfigInfo, error) {
	var configInfo model.ConfigInfo
	if err := DB.Where("data_id = ? AND group_id = ? AND tenant_id = ?", dataId, groupId, tenantId).
		First(&configInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 配置不存在时返回 nil
		}
		return nil, err // 其他错误
	}
	return &configInfo, nil
}

// IsConfigInfoExistsByTenantId 检查租户下是否还有配置记录
func IsConfigInfoExistsByTenantId(tenantId string) (bool, error) {
	var count int64
	err := DB.Model(&model.ConfigInfo{}).Where("tenant_id = ?", tenantId).Count(&count).Error
	return count > 0, err
}

// ClearOldConfigVersions 清理除最大版本外的旧版本配置
// 如果最大版本为1，则不删除任何记录
func ClearOldConfigVersions() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 直接用子查询找出最大 version，然后删除小于它的
		err := tx.Exec(`
			DELETE FROM config_info ci
			USING (
				SELECT data_id, group_id, tenant_id, MAX(version) AS max_ver
				FROM config_info
				GROUP BY data_id, group_id, tenant_id
				HAVING MAX(version) > 1
			) mv
			WHERE ci.data_id = mv.data_id
			  AND ci.group_id = mv.group_id
			  AND ci.tenant_id = mv.tenant_id
			  AND ci.version < mv.max_ver
		`).Error
		if err != nil {
			return err
		}
		return nil
	})
}
