# tk-common - 公共工具包

## 项目概述

`tk-common` 是整个微服务体系的基础工具库，提供跨服务共享的模型、工具函数和中间件实现。

**项目类型**: Go 库（Library）  
**编译目标**: 无可执行文件，仅作为依赖包被其他项目导入

## 项目结构

```
tk-common/
├── models/           # 数据模型和数据库 ORM 模型
│   ├── 用户相关模型
│   ├── 论坛/帖子相关模型  
│   └── 彩票相关模型
├── utils/            # 工具函数库
│   ├── httpresp/      - HTTP 响应工具（状态码、错误处理）
│   ├── codes/         - 业务错误码定义
│   ├── redisx/        - Redis 客户端扩展
│   ├── cmdx/          - Redis 命令扩展
│   └── ...其他工具
├── go.mod            # 模块定义
├── go.sum            # 依赖版本锁定
└── README.md         # 本文件
```

## 关键组件

### 1. Models（models/）
定义数据模型，供 ORM（GORM）映射数据库表：
- 用户模型、论坛帖子模型、评论模型等
- 与数据库表结构一一对应
- 被 tk-user、tk-business 等业务服务导入使用

### 2. HTTP 响应工具（utils/httpresp/）
标准化 HTTP 响应格式：
```go
// 成功响应
httpresp.Ok(w, data)

// 错误响应
httpresp.Fail(w, http.StatusBadRequest, codes.InvalidRequest, "invalid input")
```

### 3. Redis 扩展（utils/redisx/）
提供 Redis 的便利函数：
- `GetJSON(ctx, client, key, out)` - JSON 读取
- `SetJSON(ctx, client, key, data, ttl)` - JSON 写入
- 支持自动序列化/反序列化

### 4. 错误码定义（utils/codes/）
统一的业务错误码和 HTTP 状态码映射

## 使用方式

### 在其他服务中导入

```go
import (
    "github.com/moscososirenita-design/tk-common/models"
    "github.com/moscososirenita-design/tk-common/utils/httpresp"
    redisx "github.com/moscososirenita-design/tk-common/utils/redisx/v9"
)
```

### 关键依赖

- **GORM** - ORM 框架，用于数据库操作
- **Redis v9** - 缓存客户端
- **其他**: protobuf、grpc 等

## 维护建议

1. **不在此库中加入业务逻辑** - 仅放置通用工具和模型
2. **版本管理** - 修改模型定义时，更新所有依赖项目的 go.mod
3. **文档完善** - 为新增工具函数添加 godoc 注释
4. **单元测试** - 工具类函数应有对应的测试用例

## 相关项目

- **tk-proto**: Proto 定义和 gRPC 代码生成
- **tk-api**: HTTP API 网关
- **tk-business**: 业务服务（gRPC）
- **tk-user**: 用户服务（gRPC）
- **tk-admin**: 后台管理系统

## 开发指南

### 添加新的工具函数

1. 在 `utils/` 下创建或编辑相应的文件
2. 添加 godoc 注释说明函数用途
3. 编写单元测试（`*_test.go`）
4. 更新本 README

### 修改数据模型

1. 编辑 `models/` 目录下的模型文件
2. 对应修改数据库迁移脚本
3. 在 tk-user、tk-business 等依赖项目中执行 `go get -u github.com/moscososirenita-design/tk-common`
4. 测试各项目的编译和运行

---

**最后更新**: 2026-04-02
