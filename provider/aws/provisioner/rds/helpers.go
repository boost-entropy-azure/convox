package rds

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/convox/convox/pkg/options"
)

func convertToStringArray(s string) []string {
	arr := strings.Split(s, ",")
	for i := range arr {
		arr[i] = strings.TrimSpace(arr[i])
	}
	return arr
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "not found")
}

func targetExistsInStringArray(ss []string, target string) bool {
	if target == "" && len(ss) == 0 {
		return true
	}
	for i := range ss {
		if ss[i] == target {
			return true
		}
	}
	return false
}

func convertToStringPtr(v interface{}) (*string, error) {
	if v == nil {
		return nil, nil
	}

	switch v.(type) {
	case int32:
		vInt := v.(int32)
		return options.String(strconv.FormatInt(int64(vInt), 10)), nil
	case string:
		vStr := v.(string)
		return &vStr, nil
	case int64:
		vInt := v.(int64)
		return options.String(strconv.FormatInt(int64(vInt), 10)), nil
	case bool:
		vBool := v.(bool)
		return options.String(strconv.FormatBool(vBool)), nil
	case *int32:
		vInt, _ := v.(*int32)
		if vInt == nil {
			return nil, nil
		}
		return options.String(strconv.FormatInt(int64(*vInt), 10)), nil
	case *string:
		vStr, _ := v.(*string)
		if vStr == nil {
			return nil, nil
		}
		return vStr, nil
	case *int64:
		vInt, _ := v.(*int64)
		if vInt == nil {
			return nil, nil
		}
		return options.String(strconv.FormatInt(int64(*vInt), 10)), nil
	case *bool:
		vBool, _ := v.(*bool)
		if vBool == nil {
			return nil, nil
		}
		return options.String(strconv.FormatBool(*vBool)), nil
	case []string:
		vArr := v.([]string)
		return options.String(strings.Join(vArr, ",")), nil
	default:
		return nil, fmt.Errorf("not supported")
	}
}

func GenerateSecurePassword(length int) (string, error) {
	const (
		letters         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits          = "0123456789"
		specialChars    = "$#?"
		allChars        = letters + digits + specialChars
		minLetters      = 4
		minDigits       = 1
		minSpecialChars = 1
	)

	if length < minLetters+minDigits+minSpecialChars {
		return "", fmt.Errorf("length must be at least %d characters", minLetters+minDigits+minSpecialChars)
	}

	generateChar := func(charSet string) (byte, error) {
		max := big.NewInt(int64(len(charSet)))
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return 0, err
		}
		return charSet[n.Int64()], nil
	}

	password := make([]byte, length)
	var err error

	for i := 0; i < minLetters; i++ {
		password[i], err = generateChar(letters)
		if err != nil {
			return "", err
		}
	}
	for i := minLetters; i < minLetters+minDigits; i++ {
		password[i], err = generateChar(digits)
		if err != nil {
			return "", err
		}
	}
	for i := minLetters + minDigits; i < minLetters+minDigits+minSpecialChars; i++ {
		password[i], err = generateChar(specialChars)
		if err != nil {
			return "", err
		}
	}

	for i := minLetters + minDigits + minSpecialChars; i < length; i++ {
		password[i], err = generateChar(allChars)
		if err != nil {
			return "", err
		}
	}

	// Shuffle
	for i := range password {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(len(password))))
		if err != nil {
			return "", err
		}
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
}

func GetValueFromStringPtr(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}
