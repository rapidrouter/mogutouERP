package token

import (
	"encoding/json"
	"log"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// CustomClaims 自定义的 metadata在加密后作为 JWT 的第二部分返回给客户端
type CustomClaims struct {
	UserName string   `json:"user_name"`
	UserID   string   `json:"user_id"`
	PerAddr  string   `json:"per_addr"`
	Roles    []string `json:"roles"`

	jwt.RegisteredClaims
}

// Token jwt服务
var privateKey []byte

// InitConfig 初始化
func InitConfig(filePath string, path ...string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Fatal(err)
	}

	var current json.RawMessage = data
	if len(path) > 0 {
		current = nil
		for _, key := range path {
			v, exists := raw[key]
			if !exists {
				log.Fatalf("key %s not found in config", key)
			}
			var nested map[string]json.RawMessage
			if err := json.Unmarshal(v, &nested); err == nil {
				raw = nested
				current = v
			} else {
				current = v
			}
		}
	}

	var key string
	if err := json.Unmarshal(current, &key); err == nil {
		privateKey = []byte(key)
	} else {
		privateKey = current
	}
}

// Decode 解码
func Decode(tokenStr string) (*CustomClaims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return privateKey, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := t.Claims.(*CustomClaims); ok && t.Valid {
		return claims, nil
	}

	return nil, err
}

// Encode 将 User 用户信息加密为 JWT 字符串
func Encode(userName, userID, perAddr string, roles []string, issuer string, expireTime int64) (string, error) {
	claims := CustomClaims{
		UserName: userName,
		UserID:   userID,
		PerAddr:  perAddr,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Unix(expireTime, 0)),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(privateKey)
}