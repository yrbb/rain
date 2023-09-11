package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"hash/crc32"
	"os"
)

func MD5(str string) string {
	hash := md5.New()
	_, _ = hash.Write([]byte(str))

	return hex.EncodeToString(hash.Sum(nil))
}

func MD5File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := md5.New()
	_, _ = hash.Write(data)

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func Sha1(str string) string {
	hash := sha1.New()
	_, _ = hash.Write([]byte(str))

	return hex.EncodeToString(hash.Sum(nil))
}

func Sha1File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha1.New()
	_, _ = hash.Write(data)

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func Crc16(str string) int {
	var (
		crc = 0xFFFF
		l   = len(str)
	)

	for i := 0; i < l; i++ {
		crc ^= int(str[i])
		for j := 8; j != 0; j-- {
			if (crc & 0x0001) != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

func Crc32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Base64Decode(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
