package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/mw"
	"confkeeper/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchReq struct {
	Keyword  string `form:"keyword" binding:"required"`
	TenantId string `form:"tenant_id" binding:"omitempty"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type SearchMatch struct {
	LineNo  int    `json:"line_no"`
	Content string `json:"content"`
}

type SearchResultData struct {
	ConfigId uint          `json:"config_id"`
	DataId   string        `json:"data_id"`
	GroupId  string        `json:"group_id"`
	TenantId string        `json:"tenant_id"`
	Matches  []SearchMatch `json:"matches"`
}

type SearchResp struct {
	Code  handler.Code        `json:"code"`
	Msg   string              `json:"msg"`
	Total int64               `json:"total"`
	Data  []*SearchResultData `json:"data"`
}

// SearchConfig 搜索配置
//
//	@Tags			配置
//	@Summary		搜索配置内容
//	@Description	搜索配置内容，返回匹配行及上下文
//	@Accept			application/json
//	@Produce		application/json
//	@Param			keyword		query		string	true	"关键字"
//	@Param			tenant_id	query		string	false	"命名空间id"
//	@Param			page		query		int		false	"页码"	default(1)
//	@Param			page_size	query		int		false	"每页数量"	default(10)
//	@Success		200			{object}	SearchResp
//	@Security		ApiKeyAuth
//	@router			/api/config/search [GET]
func SearchConfig(c *gin.Context) {
	req := new(SearchReq)
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(http.StatusOK, &SearchResp{
			Code: handler.Code_ParamErr,
			Msg:  "参数错误: " + err.Error(),
		})
		return
	}

	// 权限检查
	if err := utils.IsAdmin(c); err != nil {
		// 如果指定了tenant_id，检查是否有该租户权限
		if req.TenantId != "" {
			hasPermission, err := mw.CheckNamespaceReadOrWritePermissionHTTP(c, req.TenantId)
			if err != nil || !hasPermission {
				c.JSON(http.StatusOK, &SearchResp{
					Code: handler.Code_Unauthorized,
					Msg:  "没有查看该命名空间配置的权限",
				})
				return
			}
		}
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	configs, total, err := dal.SearchConfigContent(req.Keyword, req.TenantId, offset, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, &SearchResp{
			Code: handler.Code_DBErr,
			Msg:  "搜索失败: " + err.Error(),
		})
		return
	}

	var results []*SearchResultData

	for _, config := range configs {
		// 再次权限检查（针对没指定tenant_id的情况）
		if req.TenantId == "" {
			if err := utils.IsAdmin(c); err != nil {
				hasPermission, _ := mw.CheckNamespaceReadOrWritePermissionHTTP(c, config.TenantID)
				if !hasPermission {
					continue
				}
			}
		}

		lines := strings.Split(config.Content, "\n")

		// 优化的上下文提取逻辑
		matchedLines := make(map[int]bool)
		for i, line := range lines {
			if strings.Contains(line, req.Keyword) {
				// 标记前后2行
				for j := i - 2; j <= i+2; j++ {
					if j >= 0 && j < len(lines) {
						matchedLines[j] = true
					}
				}
			}
		}

		// 正确的有序提取
		var finalMatches []SearchMatch
		for i := 0; i < len(lines); i++ {
			if matchedLines[i] {
				finalMatches = append(finalMatches, SearchMatch{
					LineNo:  i + 1, // 1-based index
					Content: lines[i],
				})
			}
		}

		if len(finalMatches) > 0 {
			results = append(results, &SearchResultData{
				ConfigId: config.ID,
				DataId:   config.DataID,
				GroupId:  config.GroupID,
				TenantId: config.TenantID,
				Matches:  finalMatches,
			})
		}
	}

	c.JSON(http.StatusOK, &SearchResp{
		Code:  handler.Code_Success,
		Msg:   "搜索成功",
		Total: total,
		Data:  results,
	})
}
