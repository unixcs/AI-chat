package store

import "golang.org/x/crypto/bcrypt"

// bcrypt is wire-compatible with the Node backend's bcryptjs ($2a$/$2b$ hashes).

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), 10)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
