// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"database/sql"
	"time"

	"github.com/pineapple/msg-demo/backend/inbox/internal/config"
	"github.com/pineapple/msg-demo/backend/inbox/internal/middleware"
)

type ServiceContext struct {
	Config         config.Config
	DB             *sql.DB
	AuthMiddleware *middleware.AuthMiddleware
	AccessSecret   []byte
	AccessExpire   time.Duration
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := c.NewMysqlConn()

	sqlDB, err := conn.RawDB()
	if err != nil {
		panic(err)
	}

	return &ServiceContext{
		Config:         c,
		DB:             sqlDB,
		AuthMiddleware: middleware.NewAuthMiddleware(c.Auth.AccessSecret),
		AccessSecret:   []byte(c.Auth.AccessSecret),
		AccessExpire:   time.Duration(c.Auth.AccessExpire) * time.Second,
	}
}
