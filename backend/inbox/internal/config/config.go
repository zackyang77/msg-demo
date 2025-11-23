// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Mysql struct {
		DataSource string
	} `json:"Mysql" yaml:"Mysql"`
	Auth struct {
		AccessSecret string `json:"AccessSecret" yaml:"AccessSecret"`
		AccessExpire int64  `json:"AccessExpire" yaml:"AccessExpire"`
	} `json:"Auth" yaml:"Auth"`
}

func (m *Config) NewMysqlConn() sqlx.SqlConn {
	return sqlx.NewSqlConn("mysql", m.Mysql.DataSource)
}
