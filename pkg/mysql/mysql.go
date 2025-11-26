package mysql

import "gorm.io/gorm"

// mysql 数据库客户端

type Client struct {
	DB *gorm.DB
}