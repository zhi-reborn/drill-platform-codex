package repository

import (
	"errors"
	"fmt"
	"time"

	"drill-platform/internal/domain/entity"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var ErrNotFound = errors.New("record not found")

// Config 数据库配置
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	MaxIdleConns int
	MaxOpenConns int
}

// InitDB 初始化数据库连接
func InitDB(cfg *Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	fmt.Print(cfg.Password)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// AutoMigrate 自动建表
func AutoMigrate() error {
	return DB.AutoMigrate(
		&entity.User{},
		&entity.DrillTemplate{},
		&entity.StepTemplate{},
		&entity.DrillInstance{},
		&entity.StepInstance{},
		&entity.StepInstanceLog{},
		&entity.DrillAssignee{},
	)
}
