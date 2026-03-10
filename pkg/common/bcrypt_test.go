package common_test

import (
	"fmt"
	"go-fiber-stater-kit/pkg/common"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, _ := common.HashPassword("atv.admin")
	fmt.Println(hash)
}

func TestComparePasswords(t *testing.T) {
	hash := "$2a$14$bC1NJxyMLt5HAaSWKYyrAu37.np8WwW9DmaXsZqQXQhEaRuQIy9aG"
	result := common.ComparePasswords(hash, "atv.admin")
	if !result {
		t.Error("Password not match")
	}
}
