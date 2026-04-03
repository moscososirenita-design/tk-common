// Package utils 提供跨服务通用工具方法。
//
// 约定：
// 1) 仅放置与业务无关、可复用的纯工具函数；
// 2) 禁止依赖具体服务层（api/business/user/admin）的内部包；
// 3) 任何服务可直接通过 tk-common/utils 引用。
package utils
