package handlers

import (
	"time"

	"github.com/google/uuid"
	"ytsruh.com/envoy/server/utils"
)

func CreateTestUser() *utils.JWTClaims {
	return &utils.JWTClaims{
		UserID: uuid.New().String(),
		Email:  "test@example.com",
		Iat:    time.Now().Unix(),
		Exp:    time.Now().Add(time.Hour).Unix(),
	}
}
