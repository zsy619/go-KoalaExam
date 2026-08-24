package encrypt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptPassword bcrypt 加密密码
func BcryptPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// BcryptCheck 校验密码
func BcryptCheck(hashed, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pwd)) == nil
}

// SHA256Hex SHA-256 摘要（用于成绩防篡改签名）
func SHA256Hex(data string, salt string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s|%s", salt, data)))
	return hex.EncodeToString(h.Sum(nil))
}
