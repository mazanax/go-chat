package security

import "golang.org/x/crypto/bcrypt"

type BcryptEncryptor struct {
	cost int
}

func NewBcryptEncryptor(cost int) BcryptEncryptor {
	return BcryptEncryptor{
		cost: cost,
	}
}

func (encryptor *BcryptEncryptor) GenerateHash(rawPassword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), encryptor.cost)
	return string(bytes), err
}

func (encryptor *BcryptEncryptor) CompareHasAndPassword(rawPassword string, encryptedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(rawPassword))
	return err == nil
}
