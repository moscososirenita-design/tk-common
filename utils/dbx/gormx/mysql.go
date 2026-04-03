package gormx

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	sqlmysql "github.com/go-sql-driver/mysql"
	"github.com/moscososirenita-design/tk-common/utils/ctxx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DBConfig 定义数据库配置。
type DBConfig struct {
	DSN             string          // 连接字符串
	MaxIdleConns    int             // 最大空闲连接数
	MaxOpenConns    int             // 最大打开连接数
	ConnMaxLifetime time.Duration   // 连接最大生命周期
	LogLevel        logger.LogLevel // GORM 日志等级
}

// Config 与历史命名保持兼容。
type Config = DBConfig

// DefaultConfig 返回默认数据库配置。
func DefaultConfig() DBConfig {
	// 返回当前处理结果。
	return DBConfig{
		// 默认空闲连接池大小。
		MaxIdleConns: 10,
		// 默认最大连接数。
		MaxOpenConns: 100,
		// 默认连接生命周期。
		ConnMaxLifetime: time.Hour,
		// 默认日志等级。
		LogLevel: logger.Warn,
	}
}

// DefaultDBConfig 与历史命名保持兼容。
func DefaultDBConfig() DBConfig {
	// 返回当前处理结果。
	return DefaultConfig()
}

// NewMySQL 创建 MySQL 数据库连接。
func NewMySQL(cfg DBConfig) (*gorm.DB, error) {
	// 判断条件并进入对应分支逻辑。
	if cfg.DSN == "" {
		// 返回当前处理结果。
		return nil, fmt.Errorf("database DSN is empty")
	}

	// 固定数据库驱动解析/写入时区为东八区，避免宿主机处于东京、北美等时区时，
	// “每天固定开奖时刻”这类 DATETIME 被错误地偏移后再写入数据库。
	normalizedDSN, err := normalizeMySQLDSNToEast8(cfg.DSN)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, fmt.Errorf("failed to normalize database dsn: %w", err)
	}

	// 定义并初始化 GORM 配置。
	gormConfig := &gorm.Config{
		// 使用单数表名，兼容现有表结构。
		NamingStrategy: schema.NamingStrategy{
			// 设置单数表名策略。
			SingularTable: true,
		},
		// 设置 GORM 日志等级。
		Logger: logger.Default.LogMode(cfg.LogLevel),
		// 统一使用 UTC 时间写库。
		NowFunc: func() time.Time {
			// 返回当前处理结果。
			return time.Now().UTC()
		},
	}

	// 打开数据库连接。
	db, err := gorm.Open(mysql.Open(normalizedDSN), gormConfig)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 获取底层 sql.DB。
	sqlDB, err := db.DB()
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 配置连接池参数。
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	// 配置连接池最大打开连接数。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	// 配置连接最大生命周期。
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 主动探活，避免启动后首次请求才暴露连接问题。
	if err := sqlDB.Ping(); err != nil {
		// 返回当前处理结果。
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 会话时区也统一固定为东八区，保证 NOW()/默认值/数据库函数行为与开奖业务一致。
	if err := db.Exec("SET time_zone = '+08:00'").Error; err != nil {
		return nil, fmt.Errorf("failed to set mysql session time_zone: %w", err)
	}

	// 输出连接成功日志。
	log.Println("MySQL database connected successfully")
	// 返回当前处理结果。
	return db, nil
}

// NewMySQLDB 与历史命名保持兼容。
func NewMySQLDB(cfg DBConfig) (*gorm.DB, error) {
	// 返回当前处理结果。
	return NewMySQL(cfg)
}

// ContextWithDB 向上下文写入数据库连接。
func ContextWithDB(ctx context.Context, db *gorm.DB) context.Context {
	// 返回当前处理结果。
	return ctxx.With(ctx, ctxx.DBKey, db)
}

// DBFromContext 从上下文读取数据库连接。
func DBFromContext(ctx context.Context) *gorm.DB {
	// 判断条件并进入对应分支逻辑。
	if db, ok := ctxx.Get[*gorm.DB](ctx, ctxx.DBKey); ok {
		// 返回当前处理结果。
		return db
	}
	// 兼容历史字符串键读取，降低迁移风险。
	if ctx != nil {
		// 判断条件并进入对应分支逻辑。
		if db, ok := ctx.Value("db").(*gorm.DB); ok {
			// 返回当前处理结果。
			return db
		}
	}
	// 返回当前处理结果。
	return nil
}

// LogLevelFromString 将字符串转换为 GORM 日志级别。
func LogLevelFromString(level string) logger.LogLevel {
	// 根据表达式进入多分支处理。
	switch level {
	case "silent":
		// 返回当前处理结果。
		return logger.Silent
	case "error":
		// 返回当前处理结果。
		return logger.Error
	case "warn":
		// 返回当前处理结果。
		return logger.Warn
	case "info":
		// 返回当前处理结果。
		return logger.Info
	default:
		// 返回当前处理结果。
		return logger.Warn
	}
}

// GormLogLevelFromString 与历史命名保持兼容。
func GormLogLevelFromString(level string) logger.LogLevel {
	// 返回当前处理结果。
	return LogLevelFromString(level)
}

// normalizeMySQLDSNToEast8 统一修正 MySQL DSN 的时间参数。
// 目标：
// 1) parseTime 必须开启，避免 DATETIME 以字符串形式流转；
// 2) loc 固定为 Asia/Shanghai，避免 loc=Local 随宿主机时区漂移；
// 3) session time_zone 通过 SQL 单独设置，这里只处理驱动解析/格式化行为。
func normalizeMySQLDSNToEast8(raw string) (string, error) {
	cfg, err := sqlmysql.ParseDSN(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return "", err
	}

	cfg.ParseTime = true
	cfg.Loc = loc

	// 保留 DSN 里其他 query 参数，仅补齐/覆盖驱动时间相关配置。
	params := url.Values{}
	for key, value := range cfg.Params {
		params.Set(key, value)
	}
	cfg.Params = map[string]string{}
	for key, values := range params {
		if len(values) == 0 {
			continue
		}
		cfg.Params[key] = values[0]
	}

	return cfg.FormatDSN(), nil
}
