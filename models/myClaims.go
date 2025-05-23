package models

import (
	"github.com/golang-jwt/jwt/v5"
)

// MyClaims はJWTクレームの構造体定義です。
type MyClaims struct {
	UserID             uint   `json:"userid"`
	SubscriptionStatus string `json:"subscriptionStatus"`
	jwt.RegisteredClaims
}
