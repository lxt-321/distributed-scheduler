package store

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"dscheduler/config"
)

// DB 全局 MySQL 连接
var DB *sql.DB

// InitMySQL 初始化 MySQL 连接池
func InitMySQL() error {
	c := config.Global.MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DB)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("MySQL 连接失败(请检查密码/端口): %w", err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	DB = db
	return nil
}
