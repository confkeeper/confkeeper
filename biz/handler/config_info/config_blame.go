package config_info

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"confkeeper/biz/mw"
	"confkeeper/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// blameMaxLines 超过该行数则跳过 blame 计算，避免大文件的 O(n*m) 性能问题
const blameMaxLines = 5000

// maxVersions 超过该版本则跳过 blame 计算，避免大文件的 O(n*m) 性能问题
const maxVersions = 100

type BlameUriReq struct {
	ConfigId string `uri:"config_id" binding:"required"`
}

type BlameRun struct {
	StartLine  int    `json:"start_line"` // 1-indexed
	EndLine    int    `json:"end_line"`   // 1-indexed
	Author     string `json:"author"`
	CreateTime string `json:"create_time"`
	Version    int    `json:"version"`
}

type BlameData struct {
	TotalLines int         `json:"total_lines"`
	Runs       []*BlameRun `json:"runs"`
}

type BlameResp struct {
	handler.CommonResp
	Data *BlameData `json:"data"`
}

// diffEntry 表示一行差异
type diffEntry struct {
	typ     string // "equal" | "added" | "removed"
	oldLine int    // 0-indexed
	newLine int    // 0-indexed
}

// splitLines 将内容按行切分，统一换行符为 \n
func splitLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if normalized == "" {
		return []string{}
	}
	return strings.Split(normalized, "\n")
}

// diffLines 基于 LCS 的逐行差异算法，与前端 editorDiffDecorations 保持一致
func diffLines(oldLines, newLines []string) []diffEntry {
	n := len(oldLines)
	m := len(newLines)
	width := m + 1
	dp := make([]int, (n+1)*width)

	for i := n - 1; i >= 0; i-- {
		rowOffset := i * width
		nextRowOffset := (i + 1) * width
		oldLine := oldLines[i]
		for j := m - 1; j >= 0; j-- {
			if oldLine == newLines[j] {
				dp[rowOffset+j] = dp[nextRowOffset+j+1] + 1
			} else {
				v1 := dp[nextRowOffset+j]
				v2 := dp[rowOffset+j+1]
				if v1 > v2 {
					dp[rowOffset+j] = v1
				} else {
					dp[rowOffset+j] = v2
				}
			}
		}
	}

	result := make([]diffEntry, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		if oldLines[i] == newLines[j] {
			result = append(result, diffEntry{typ: "equal", oldLine: i, newLine: j})
			i++
			j++
		} else if dp[(i+1)*width+j] >= dp[i*width+j+1] {
			result = append(result, diffEntry{typ: "removed", oldLine: i})
			i++
		} else {
			result = append(result, diffEntry{typ: "added", newLine: j})
			j++
		}
	}
	for i < n {
		result = append(result, diffEntry{typ: "removed", oldLine: i})
		i++
	}
	for j < m {
		result = append(result, diffEntry{typ: "added", newLine: j})
		j++
	}
	return result
}

// computeBlame 计算最新版本每一行最后被哪个版本修改。
// versions 必须按版本升序排列。返回按行号分组的 blame 区间。
func computeBlame(versions []*model.ConfigInfo) ([]*BlameRun, int) {
	if len(versions) == 0 {
		return nil, 0
	}

	if len(versions) > maxVersions {
		versions = versions[len(versions)-maxVersions:]
	}

	latestLines := splitLines(versions[len(versions)-1].Content)
	totalLines := len(latestLines)
	if totalLines == 0 || totalLines > blameMaxLines {
		return nil, totalLines
	}

	for _, v := range versions {
		if strings.Count(v.Content, "\n") > blameMaxLines {
			return nil, totalLines
		}
	}

	// blame[i] = line i (0-indexed) 对应的 versions 数组下标
	// 初始：所有行归属最旧版本(下标 0)
	prevLines := splitLines(versions[0].Content)
	prevBlame := make([]int, len(prevLines))
	for i := range prevBlame {
		prevBlame[i] = 0
	}

	// 逐版本推进：与上一版本做 diff，未变更的行继承上一版本的 blame，变更/新增的行归属当前版本
	for k := 1; k < len(versions); k++ {
		curLines := splitLines(versions[k].Content)
		curBlame := make([]int, len(curLines))
		for i := range curBlame {
			curBlame[i] = k
		}
		for _, d := range diffLines(prevLines, curLines) {
			if d.typ == "equal" {
				curBlame[d.newLine] = prevBlame[d.oldLine]
			}
		}
		prevLines = curLines
		prevBlame = curBlame
	}

	// 此时 prevBlame 即为最新版本每行的归属
	blame := prevBlame

	// 将连续归属同一版本的行合并为一个区间
	runs := make([]*BlameRun, 0)
	start := 0
	for i := 1; i <= totalLines; i++ {
		if i == totalLines || blame[i] != blame[start] {
			v := versions[blame[start]]
			runs = append(runs, &BlameRun{
				StartLine:  start + 1,
				EndLine:    i,
				Author:     v.Author,
				CreateTime: v.CreateTime.Format("2006-01-02 15:04:05"),
				Version:    v.Version,
			})
			start = i
		}
	}
	return runs, totalLines
}

// ConfigBlame 获取配置每行最后修改人/时间（类似 GitLens blame）
//
//	@Tags			配置
//	@Summary		获取配置行修改记录
//	@Description	通过config_id查询所有版本，计算每行最后修改人和时间，按行号区间返回
//	@Accept			application/json
//	@Produce		application/json
//	@Param			config_id	path		string	true	"配置ID"
//	@Success		200			{object}	BlameResp
//	@Failure		400			{object}	BlameResp	"参数错误"
//	@Failure		404			{object}	BlameResp	"配置不存在"
//	@Failure		500			{object}	BlameResp	"服务器错误"
//	@router			/api/config/blame/{config_id} [GET]
func ConfigBlame(c *gin.Context) {
	req := new(BlameUriReq)
	if err := c.ShouldBindUri(req); err != nil {
		c.JSON(http.StatusBadRequest, &BlameResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_ParamErr,
				Msg:  "参数错误: " + err.Error(),
			},
		})
		return
	}
	resp := new(BlameResp)

	// 获取配置信息以检查权限
	configInfoData, err := dal.GetConfigInfoByID(req.ConfigId)
	if err != nil {
		c.JSON(http.StatusOK, &BlameResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "数据库查询错误: " + err.Error(),
			},
		})
		return
	}
	if configInfoData == nil {
		c.JSON(http.StatusNotFound, &BlameResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_Err,
				Msg:  "配置不存在",
			},
		})
		return
	}

	// 权限检查：管理员或有命名空间r/rw权限的用户
	if err := utils.IsAdmin(c); err != nil {
		hasPermission, err := mw.CheckNamespaceReadOrWritePermissionHTTP(c, configInfoData.TenantID)
		if err != nil || !hasPermission {
			c.JSON(http.StatusOK, &BlameResp{
				CommonResp: handler.CommonResp{
					Code: handler.Code_Unauthorized,
					Msg:  "没有查看配置的权限",
				},
			})
			return
		}
	}

	// 获取所有版本（按版本升序，过滤 tenant）
	versions, err := dal.GetAllVersionsByDataIdGroupAndTenantAsc(configInfoData.DataID, configInfoData.GroupID, configInfoData.TenantID)
	if err != nil {
		c.JSON(http.StatusOK, &BlameResp{
			CommonResp: handler.CommonResp{
				Code: handler.Code_DBErr,
				Msg:  "数据库查询错误: " + err.Error(),
			},
		})
		return
	}

	runs, totalLines := computeBlame(versions)

	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "获取配置行修改信息成功",
	}
	resp.Data = &BlameData{
		TotalLines: totalLines,
		Runs:       runs,
	}

	c.JSON(http.StatusOK, resp)
}
