package util

import (
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var (
	// 手机号（中国大陆）
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

	// 登录ID：字母开头，允许字母数字下划线，4-20位
	loginIDRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{3,19}$`)

	// 用户名（中文/英文/数字/下划线，2-20位）
	usernameRegex = regexp.MustCompile(`^[\p{Han}a-zA-Z0-9_]{2,20}$`)

	// 身份证号（18位，最后一位可为X）
	idCardRegex = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)

	// 邮编（中国大陆6位）
	postcodeRegex = regexp.MustCompile(`^[1-9]\d{5}$`)

	// 数字
	numberRegex = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

	// 整数
	intRegex = regexp.MustCompile(`^-?\d+$`)

	// 正整数
	positiveIntRegex = regexp.MustCompile(`^\d+$`)

	// 字母
	alphaRegex = regexp.MustCompile(`^[a-zA-Z]+$`)

	// 字母+数字
	alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	// 十六进制颜色
	hexColorRegex = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

	// IP v4
	ipv4Regex = regexp.MustCompile(`^((25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}(25[0-5]|2[0-4]\d|[01]?\d?\d)$`)

	// 银行卡号（16-19位）
	bankCardRegex = regexp.MustCompile(`^\d{16,19}$`)

	// 统一社会信用代码（18位）
	usciRegex = regexp.MustCompile(`^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$`)
)

// ========== 邮箱 ==========
func IsEmail(s string) bool {
	if s == "" || len(s) > 254 {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

// ========== 手机号 ==========
func IsPhone(s string) bool {
	return phoneRegex.MatchString(s)
}

// ========== LoginID ==========
func IsLoginID(s string) bool {
	return loginIDRegex.MatchString(s)
}

// ========== 用户名 ==========
func IsUsername(s string) bool {
	return usernameRegex.MatchString(s)
}

// ========== URL ==========
func IsURL(s string) bool {
	if s == "" {
		return false
	}
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

// ========== 身份证 ==========
func IsIDCard(s string) bool {
	if !idCardRegex.MatchString(s) {
		return false
	}
	return checkIDCardChecksum(s)
}

func checkIDCardChecksum(s string) bool {
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
	sum := 0
	for i := 0; i < 17; i++ {
		sum += int(s[i]-'0') * weights[i]
	}
	return checkCodes[sum%11] == byte(unicode.ToUpper(rune(s[17])))
}

// ========== 邮编 ==========
func IsPostcode(s string) bool {
	return postcodeRegex.MatchString(s)
}

// ========== 数字 ==========
func IsNumber(s string) bool {
	return numberRegex.MatchString(s)
}

func IsInt(s string) bool {
	return intRegex.MatchString(s)
}

func IsPositiveInt(s string) bool {
	return positiveIntRegex.MatchString(s)
}

// ========== 字母/数字 ==========
func IsAlpha(s string) bool {
	return alphaRegex.MatchString(s)
}

func IsAlphanumeric(s string) bool {
	return alphanumericRegex.MatchString(s)
}

// ========== 颜色 ==========
func IsHexColor(s string) bool {
	return hexColorRegex.MatchString(s)
}

// ========== IP ==========
func IsIPv4(s string) bool {
	return ipv4Regex.MatchString(s)
}

// ========== 银行卡 ==========
func IsBankCard(s string) bool {
	if !bankCardRegex.MatchString(s) {
		return false
	}
	return luhnCheck(s)
}

// Luhn 算法校验
func luhnCheck(s string) bool {
	sum := 0
	alternate := false
	for i := len(s) - 1; i >= 0; i-- {
		d := int(s[i] - '0')
		if alternate {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alternate = !alternate
	}
	return sum%10 == 0
}

// ========== 统一社会信用代码 ==========
func IsUSCI(s string) bool {
	return usciRegex.MatchString(s)
}

// ========== 密码强度 ==========
// level: 1-弱 2-中 3-强
func PasswordStrength(s string) int {
	if len(s) < 8 {
		return 0
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range s {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	count := 0
	for _, ok := range []bool{hasUpper, hasLower, hasDigit, hasSpecial} {
		if ok {
			count++
		}
	}
	if len(s) >= 12 && count >= 3 {
		return 3
	}
	if count >= 2 {
		return 2
	}
	return 1
}

// ========== 中文 ==========
func IsChinese(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}

// ========== 常用综合 ==========

// IsEmpty 空字符串
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// InRange 长度范围
func InRange(s string, min, max int) bool {
	l := len([]rune(s))
	return l >= min && l <= max
}
