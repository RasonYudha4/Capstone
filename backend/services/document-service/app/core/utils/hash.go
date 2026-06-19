package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
)

func GenerateHMAC(filebytes []byte) string {
	secret := []byte(os.Getenv("HMAC_SECRET"))

	mac := hmac.New(sha256.New, secret)
	mac.Write(filebytes)

	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyHMAC(filebytes []byte, storedHash string) bool {
	hash := GenerateHMAC(filebytes)

	return hmac.Equal(
		[]byte(hash),
		[]byte(storedHash),
	)
}