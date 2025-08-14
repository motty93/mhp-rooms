package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateRoomCode ランダムな部屋コードを生成する（6文字の英数字）
func GenerateRoomCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6
	
	code := make([]byte, codeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("ランダムコード生成に失敗しました: %w", err)
		}
		code[i] = charset[num.Int64()]
	}
	
	return string(code), nil
}

// GenerateUniqueRoomCode 一意性チェック付きで部屋コードを生成する
// checkExists は指定されたコードが既に存在するかをチェックする関数
func GenerateUniqueRoomCode(checkExists func(string) (bool, error)) (string, error) {
	const maxAttempts = 10
	
	for i := 0; i < maxAttempts; i++ {
		code, err := GenerateRoomCode()
		if err != nil {
			return "", err
		}
		
		exists, err := checkExists(code)
		if err != nil {
			return "", fmt.Errorf("コード重複チェックに失敗しました: %w", err)
		}
		
		if !exists {
			return code, nil
		}
	}
	
	return "", fmt.Errorf("一意な部屋コードの生成に失敗しました（%d回試行）", maxAttempts)
}