package postgre

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	config "yunyez/internal/common/config"

	logger "yunyez/pkg/logger"
	gorm_logger "gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// postgresql

type Client struct {
	DB *gorm.DB
}

var (
	defaultTimeout = time.Duration(5) * time.Second // 默认连接5秒超时
	once sync.Once
	mutex sync.RWMutex
	clientInstance *Client
)

func NewClient() error {
	var err error
	once.Do(func(){
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()
		// 读取配置
		host := config.GetString("database.postgres.host")
		port := config.GetInt("database.postgres.port")
		user := config.GetString("database.postgres.user")
		password := config.GetString("database.postgres.password")
		dbname := config.GetString("database.postgres.dbname")
		sslmode := config.GetString("database.postgres.sslmode")
		schemas := config.GetList("database.postgres.schemas")
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
		// 连接数据库
		clientInstance = &Client{}
		sqlLogger := logger.NewSQLLogger(logger.DefaultLogger).(*logger.SQLLogger)
		clientInstance.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			// 配置SQL日志记录器
			Logger: sqlLogger.LogMode(gorm_logger.Info),
		})
		if err != nil {
			logger.Error(ctx, "Failed to connect to PostgreSQL: %v", map[string]any{
				"Error": err,
			})
			return
		}
		sqlDB, err := clientInstance.DB.DB()
		if err != nil {
			logger.Error(ctx, "Failed to get PostgreSQL DB: %v", map[string]any{
				"Error": err,
			})
			return
		}

		// 设置连接池参数
		sqlDB.SetMaxIdleConns(config.GetInt("database.postgres.max_idle_conns"))
		sqlDB.SetMaxOpenConns(config.GetInt("database.postgres.max_open_conns"))
		sqlDB.SetConnMaxLifetime(time.Duration(config.GetInt("database.postgres.conn_max_lifetime")) * time.Second)

		// 测试数据库连接
		if err := sqlDB.PingContext(ctx); err != nil {
			logger.Error(ctx, "Failed to ping PostgreSQL: %v", map[string]any{
				"Error": err,
			})
			return
		}

		// 设置默认 schema
		if len(schemas) > 0 {
			searchPath := strings.Join(schemas, ",")
			err = clientInstance.DB.Exec(fmt.Sprintf("SET search_path TO %s", searchPath)).Error
			if err != nil {
				logger.Error(ctx, "Failed to set search_path: %v", map[string]any{
					"Error": err,
					"SearchPath": searchPath,
				})
			}
		}

		logger.Info(ctx, "PostgreSQL connected successfully", map[string]any{
			"Host": host,
			"Port": port,
			"User": user,
			// "Password": password,
			"DBName": dbname,
			"Schemas": schemas,
		})
	})

	return  err
}

// GetClient 获取 PostgreSQL 客户端实例
// 该函数返回全局的 PostgreSQL 客户端实例。
// 如果实例不存在，则会尝试初始化一个新的客户端。
// 如果初始化失败，会记录错误日志并返回 nil。
// 控制流程：
// 1. 先尝试读取锁（RLock），如果实例不存在，则释放锁并继续。
// 2. 如果实例不存在，尝试获取写锁（Lock）。
// 3. 再次检查实例是否存在，不存在则初始化。
// 4. 初始化完成后，释放写锁。
// 5. 最后返回实例（如果存在）。
func GetClient() *Client {
	ctx := context.Background()
	mutex.RLock()
	if clientInstance == nil {
		// 重新初始化
		mutex.RUnlock()
		mutex.Lock()
		// 双重检查
		if clientInstance == nil {
			err := NewClient()
			if err != nil {
				logger.Error(ctx, "Failed to initialize PostgreSQL client: %v", map[string]any{
					"Error": err,
				})
				return nil
			}
		}
		mutex.Unlock()
		mutex.RLock()
	}
	mutex.RUnlock()
	return clientInstance
}

// WithSchema 设置连接模式
// 该方法用于设置 PostgreSQL 客户端的连接模式。默认是 "public"
// 特殊情况下，需要切换 schema 时使用，一般情况下不建议使用
func (c *Client) WithSchema(schema string) (*gorm.DB, error) {
    if c == nil || c.DB == nil {
        return nil, errors.New("client not initialized")
    }

    // 获取底层 sql.DB
    sqlDB, err := c.DB.DB()
    if err != nil {
        return nil, err
    }

    // 创建一个新连接（不从池中取，避免影响其他请求）
    newConn, err := sqlDB.Conn(context.Background())
    if err != nil {
        return nil, err
    }
	defer newConn.Close()

    // 在新连接上设置 search_path
    _, err = newConn.ExecContext(context.Background(), fmt.Sprintf("SET search_path TO %s, public", schema))
    if err != nil {
        newConn.Close()
        return nil, err
    }

    // 用这个连接创建新的 gorm.DB
    newGormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: newConn}), &gorm.Config{})
    if err != nil {
        newConn.Close()
        return nil, err
    }

    return newGormDB, nil
}

// Close 关闭 PostgreSQL 客户端连接
// 该方法用于关闭 PostgreSQL 客户端的数据库连接。
// 它会尝试获取底层的 sql.DB 实例，并调用其 Close 方法来关闭连接。
// 如果关闭过程中发生错误，会返回该错误；否则返回 nil。
func (c *Client) Close() error {
    if c == nil || c.DB == nil {
        return errors.New("client not initialized")
    }

	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	if err != nil {
		return err
	}

	mutex.Lock()
	clientInstance = nil
	mutex.Unlock()

	return nil
}