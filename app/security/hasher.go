package security

type PasswordEncryptor interface {
	GenerateHash(rawPassword string) (string, error)
	CompareHasAndPassword(rawPassword string, encryptedPassword string) bool
}
