package utils

import "github.com/alexedwards/argon2id"

func HashPassword(password string) (string, error) {

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordAndHash(password, hash string) error {
	_, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return err
	}

	return nil
}
