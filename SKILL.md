# confkeeper 项目开发规范

## 项目概述

confkeeper 是一个基于 Go 和 Gin 框架的配置中心，支持 SQLite、MySQL、PostgreSQL 数据库。

**技术栈**: Go + Gin + GORM

**目录结构**:
```
biz/
├── handler/     # API 处理器（按模块组织）
├── model/       # 数据模型
├── dal/         # 数据访问层
├── router/      # 路由定义
└── mw/          # 中间件
```

---

## 一、Handler（处理器）

### 文件位置
`biz/handler/<模块>/<操作>.go`

### 代码模板

```go
package <模块>

import (
	"confkeeper/biz/dal"
	"confkeeper/biz/handler"
	"confkeeper/biz/model"
	"net/http"
	"github.com/gin-gonic/gin"
)

type <操作>Req struct {
	字段 string `json:"字段" binding:"required,min=1,max=255"`
}

type <操作>Data struct {
	// 响应数据
}

type <操作>Resp struct {
	handler.CommonResp
	Data *<操作>Data `json:"data"`
}

// <处理器名> 处理器描述
//
//	@Tags			<模块>
//	@Summary		简短描述
//	@Description	详细描述
//	@Accept			application/json
//	@Produce		application/json
//	@Param			req	body		<操作>Req	true	"请求参数"
//	@Success		200	{object}	<操作>Resp
//	@router			/api/<模块>/<操作> [POST]
func <处理器名>(c *gin.Context) {
	req := new(<操作>Req)
	if err := c.ShouldBind(req); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	
	resp := new(<操作>Resp)
	
	// 业务逻辑：调用 DAL 函数
	
	resp.CommonResp = handler.CommonResp{
		Code: handler.Code_Success,
		Msg:  "操作成功",
	}
	
	c.JSON(http.StatusOK, resp)
}
```

### 响应码

| 响应码 | 值 | 说明 |
|--------|-----|------|
| Code_Success | 200 | 成功 |
| Code_ParamErr | 400 | 参数错误 |
| Code_Unauthorized | 401 | 未授权 |
| Code_Err | 500 | 通用错误 |
| Code_DBErr | 501 | 数据库错误 |
| Code_PasswordErr | 502 | 密码错误 |
| Code_AlreadyExists | 503 | 资源已存在 |
| Code_CaptchaErr | 504 | 验证码错误 |
| Code_UserErr | 505 | 用户错误 |

### 规范要点
- 使用 Gin binding 标签验证输入
- 中文错误消息
- 添加 Swagger 注释
- 处理器专注 HTTP 层，业务逻辑在 DAL 层
- 使用 `c.ShouldBind()` 进行 JSON 绑定

---

## 二、Model（数据模型）

### 文件位置
`biz/model/<实体名>.go`

### 代码模板

```go
package model

type <实体名> struct {
	ID     uint   `gorm:"primarykey;comment:主键ID" json:"id"`
	字段1  string `gorm:"column:字段1;type:varchar(255);comment:字段说明" json:"字段1"`
	字段2  int    `gorm:"column:字段2;type:int;comment:字段说明" json:"字段2"`
	字段3  bool   `gorm:"column:字段3;type:boolean;comment:字段说明" json:"字段3"`
}

func (e *<实体名>) TableName() string {
	return "<表名>"
}

func (e *<实体名>) TableComment() string {
	return "<表注释>"
}
```

### 常用 GORM 标签

| 标签 | 说明 | 示例 |
|------|------|------|
| primarykey | 主键 | `gorm:"primarykey"` |
| column | 列名 | `gorm:"column:user_name"` |
| type | 数据类型 | `gorm:"type:varchar(255)"` |
| comment | 注释 | `gorm:"comment:用户名"` |
| unique | 唯一约束 | `gorm:"unique"` |
| not null | 非空约束 | `gorm:"not null"` |
| default | 默认值 | `gorm:"default:0"` |
| index | 创建索引 | `gorm:"index"` |

### 规范要点
- 中文注释描述字段
- 始终包含 `ID` 主键
- 数据库列名：snake_case
- Go 字段名：PascalCase
- 为常用查询字段添加约束和索引

### 数据库迁移

在 `bootstrap/db.go` 中添加：
```go
err = DB.AutoMigrate(
	&model.User{},
	&model.<新模型>{},
)
```

---

## 三、Router（路由）

### 文件位置
`biz/router/<模块>.go`

### 代码模板

```go
package router

import (
	h<模块> "confkeeper/biz/handler/<模块>"
	"confkeeper/biz/mw"
	"confkeeper/utils"
	"github.com/gin-gonic/gin"
)

func <模块>Routes(apiGroup *gin.RouterGroup) {
	<模块>Group := apiGroup.Group("/<模块>")
	{
		// 公开路由
		<模块>Group.GET("/public", h<模块>.PublicHandler)
		
		// 认证路由
		<模块>Group.GET("/list", mw.JWTAuthMiddleware(), h<模块>.List)
		
		// 管理员路由
		<模块>Group.DELETE("/delete/:id", 
			mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true}), 
			h<模块>.Delete)
	}
}
```

### 路由注册

在 `biz/router/register_routes.go` 中：
```go
func RegisterRoutes(r *gin.Engine) {
	apiGroup := r.Group("/api")
	{
		userRoutes(apiGroup)
		<模块>Routes(apiGroup)  // 添加新路由
	}
}
```

### 中间件配置

| 中间件 | 用途 |
|--------|------|
| `mw.JWTAuthMiddleware()` | 基础认证 |
| `mw.JWTAuthMiddleware(utils.AuthOptions{CheckAdmin: true})` | 仅管理员 |
| `mw.JWTAuthMiddleware(utils.AuthOptions{IsShortTerm: true})` | 短期令牌 |
| `mw.PermissionCheck("permission")` | 权限检查 |

### 路由模式

| 操作 | 路由 | 方法 |
|------|------|------|
| 列表 | `/<模块>/list` | GET |
| 创建 | `/<模块>/add` | PUT |
| 更新 | `/<模块>/update/:id` | POST |
| 删除 | `/<模块>/delete/:id` | DELETE |
| 详情 | `/<模块>/info/:id` | GET |
| 搜索 | `/<模块>/search` | GET |

### 规范要点
- 相关路由组织在同一路径前缀
- 一致的命名（list、create、update、delete）
- 受保护路由应用认证中间件
- 路径参数使用小写和下划线（`:user_id`）

---

## 四、Middleware（中间件）

### 文件位置
`biz/mw/<名称>.go`

### 代码模板

```go
package mw

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// <中间件名> 中间件描述
func <中间件名>() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 前置处理
		
		if someCondition {
			c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code": http.StatusUnauthorized,
				"msg":  "错误信息",
			})
			c.Abort()
			return
		}
		
		c.Set("key", "value")
		c.Next()
		
		// 后置处理（可选）
	}
}
```

### 上下文值

| 键 | 类型 | 说明 |
|------|------|------|
| `userid` | int | 用户 ID |
| `username` | string | 用户名 |
| `token` | string | JWT 令牌 |

### 使用方式

```go
// 单个路由
userGroup.GET("/profile", mw.JWTAuthMiddleware(), mw.UserStatusCheck(), hUser.GetProfile)

// 路由组
auditGroup := apiGroup.Group("/audit")
auditGroup.Use(mw.JWTAuthMiddleware(), mw.AuditLog())
```

### 规范要点
- 单一职责
- 中文错误消息
- 除非中止，否则调用 `c.Next()`
- 错误时使用 `c.Abort()` 终止

---

## 五、DAL（数据访问层）

### 文件位置
`biz/dal/<实体名>.go`

### 代码模板

```go
package dal

import (
	"confkeeper/biz/model"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// Create<Entity> 创建实体
func Create<Entity>(entities []*model.<Entity>) error {
	return DB.Create(entities).Error
}

// Get<Entity>ByID 根据 ID 获取
func Get<Entity>ByID(id int) (*model.<Entity>, error) {
	var entity model.<Entity>
	if err := DB.First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Get<Entity>List 获取列表（分页）
func Get<Entity>List(pageSize, offset int, filter string) ([]*model.<Entity>, int64, error) {
	var entities []*model.<Entity>
	
	query := DB.Model(&model.<Entity>{})
	if filter != "" {
		query = query.Where("field LIKE ?", "%"+filter+"%")
	}
	
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	if err := query.Order("id").Offset(offset).Limit(pageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	
	return entities, total, nil
}

// Update<Entity> 更新实体
func Update<Entity>(entity *model.<Entity>) error {
	return DB.Model(entity).Updates(map[string]interface{}{
		"field1": entity.Field1,
		"field2": entity.Field2,
	}).Error
}

// Delete<Entity> 删除实体
func Delete<Entity>(id int) error {
	var entity model.<Entity>
	if err := DB.First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("记录不存在或已被删除")
		}
		return err
	}
	return DB.Delete(&entity).Error
}
```

### 常用模式

#### 存在性检查
```go
func Is<Entity>Exists(field string) (bool, error) {
	var count int64
	err := DB.Model(&model.<Entity>{}).Where("field = ?", field).Count(&count).Error
	return count > 0, err
}
```

#### 事务操作
```go
func Create<Entity>WithTransaction(entity *model.<Entity>, related *model.Related) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entity).Error; err != nil {
			return err
		}
		related.EntityID = entity.ID
		return tx.Create(related).Error
	})
}
```

### 规范要点
- 有意义的函数名（Create、Get、Update、Delete）
- 单条记录：`*model.Entity`，列表：`[]*model.Entity`
- 错误作为最后一个返回值
- 中文错误消息
- 明确处理 `gorm.ErrRecordNotFound`
- 列表查询使用分页
- 多表操作使用事务

---

## 六、开发流程

### 添加新功能

1. **创建 Model** (`biz/model/<实体>.go`)
   - 定义数据结构
   - 添加 GORM 标签
   - 在 `bootstrap/db.go` 注册迁移

2. **创建 DAL** (`biz/dal/<实体>.go`)
   - 实现 CRUD 操作
   - 处理错误

3. **创建 Handler** (`biz/handler/<模块>/<操作>.go`)
   - 定义请求/响应结构
   - 实现业务逻辑
   - 添加 Swagger 注释

4. **创建 Router** (`biz/router/<模块>.go`)
   - 定义路由
   - 应用中间件
   - 在 `register_routes.go` 注册

5. **测试**
   - 启动应用：`go run . -c=config/config.yaml`
   - 测试 API

---

## 七、通用规范

### 命名规范
- 文件名：snake_case（`user_login.go`）
- 函数名：PascalCase（`UserLogin`）
- 数据库列名：snake_case（`user_name`）
- JSON 字段：camelCase（`userName`）

### 注释规范
- 中文注释
- 公共函数添加文档注释
- 复杂逻辑添加说明

### 错误处理
- 使用中文错误消息
- 明确处理 `gorm.ErrRecordNotFound`
- 返回适当的响应码

### 数据库兼容性
- 考虑 SQLite、MySQL、PostgreSQL
- 使用 GORM 抽象层
- 避免数据库特定语法

---

## 八、常用命令

```bash
# 启动开发服务器
go run . -c=config/config.yaml

# 编译
go build -o confkeeper

# 运行测试
go test ./...

# 生成 Swagger 文档
swag init
```

---

## 九、项目约定

1. **语言**: 所有错误消息、注释使用中文
2. **架构**: 分层架构（Handler → DAL → Model）
3. **认证**: JWT + 中间件
4. **数据库**: GORM，支持多数据库
5. **API 文档**: Swagger
6. **日志**: gookit/slog

---

**遵循以上规范可确保代码风格一致、可维护性高、易于协作。**
