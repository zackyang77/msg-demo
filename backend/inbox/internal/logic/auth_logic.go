package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/pineapple/msg-demo/backend/inbox/internal/svc"
	"github.com/pineapple/msg-demo/backend/inbox/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (*types.AuthResponse, error) {
	if err := validateCredentials(req.Username, req.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("生成密码哈希失败: %w", err)
	}

	result, err := l.svcCtx.DB.ExecContext(
		l.ctx,
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		req.Username,
		string(hash),
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, fmt.Errorf("用户名已被占用")
		}
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取用户ID失败: %w", err)
	}

	token, err := l.generateToken(userID)
	if err != nil {
		return nil, err
	}

	return &types.AuthResponse{
		Token: token,
		User: types.User{
			Id:       userID,
			Username: req.Username,
		},
	}, nil
}

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (*types.AuthResponse, error) {
	if err := validateCredentials(req.Username, req.Password); err != nil {
		return nil, err
	}

	var (
		id           int64
		passwordHash string
	)

	err := l.svcCtx.DB.QueryRowContext(
		l.ctx,
		`SELECT id, password_hash FROM users WHERE username = ?`,
		req.Username,
	).Scan(&id, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("用户名或密码错误")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}

	token, err := l.generateToken(id)
	if err != nil {
		return nil, err
	}

	return &types.AuthResponse{
		Token: token,
		User: types.User{
			Id:       id,
			Username: req.Username,
		},
	}, nil
}

func (l *RegisterLogic) generateToken(userID int64) (string, error) {
	return generateToken(l.svcCtx, userID)
}

func (l *LoginLogic) generateToken(userID int64) (string, error) {
	return generateToken(l.svcCtx, userID)
}

func generateToken(ctx *svc.ServiceContext, userID int64) (string, error) {
	expireAt := time.Now().Add(ctx.AccessExpire).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"exp":    expireAt,
	})

	signed, err := token.SignedString(ctx.AccessSecret)
	if err != nil {
		return "", fmt.Errorf("生成 token 失败: %w", err)
	}

	return signed, nil
}

func validateCredentials(username, password string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 {
		return fmt.Errorf("用户名至少 3 个字符")
	}
	if len(password) < 6 {
		return fmt.Errorf("密码至少 6 个字符")
	}
	return nil
}
