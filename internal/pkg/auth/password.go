package auth

import "golang.org/x/crypto/bcrypt"

const (
	// DefaultCost 默认 bcrypt cost
	DefaultCost = 12
)

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string, cost ...int) (string, error) {
	c := DefaultCost
	if len(cost) > 0 && cost[0] > 0 {
		c = cost[0]
	}
	
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), c)
	if err != nil {
		return "", err
	}
	
	return string(hashedBytes), nil
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
