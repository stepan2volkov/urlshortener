package base58

import (
	"fmt"
)

var alphabet = []rune("123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ")
var alphabetIntValues map[rune]int
var base int = len(alphabet)

func init() {
	alphabetIntValues = make(map[rune]int)
	for i, r := range alphabet {
		alphabetIntValues[r] = i
	}
}

// Decode number into base58 string. The number should be positive or zero.
func Decode(num int) (string, error) {
	if num < 0 {
		return "", fmt.Errorf("num shouldn't be negative, got %d", num)
	}

	result := make([]rune, 0)
	for {
		result = append(result, alphabet[num%base])

		if num < base {
			break
		}
		num = num / base
	}

	reverseRuneStr(result)

	return string(result), nil
}

// Encode base58 string to number.
func Encode(str string) (int, error) {
	result := 0
	for _, r := range str {
		value, exists := alphabetIntValues[r]
		if !exists {
			return 0, fmt.Errorf("unexpected character in str: %v", r)
		}
		result = result*base + value
	}
	return result, nil
}

func reverseRuneStr(runeStr []rune) {
	for i, j := 0, len(runeStr)-1; i < len(runeStr)/2; i, j = i+1, j-1 {
		runeStr[i], runeStr[j] = runeStr[j], runeStr[i]
	}
}
