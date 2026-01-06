/*
 * @Author              : 寂情啊
 * @Date                : 2025-11-18 11:11:08
 * @LastEditors         : 寂情啊
 * @LastEditTime        : 2025-11-18 11:11:31
 * @FilePath            : frp-web-testbackendinternalutiljwt_test.go
 * @Description         : 说明
 * 倾尽绿蚁花尽开，问潭底剑仙安在哉
 */
package util

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	userID := uint(1)
	username := "testuser"
	expireHours := 24

	token, err := GenerateToken(userID, username, secret, expireHours)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, userID)
	}

	if claims.Username != username {
		t.Errorf("Username mismatch: got %s, want %s", claims.Username, username)
	}
}

func TestParseExpiredToken(t *testing.T) {
	secret := "test-secret"
	token, _ := GenerateToken(1, "test", secret, -1)

	time.Sleep(2 * time.Second)

	_, err := ParseToken(token, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestParseInvalidToken(t *testing.T) {
	secret := "test-secret"
	invalidToken := "invalid.token.here"

	_, err := ParseToken(invalidToken, secret)
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}
