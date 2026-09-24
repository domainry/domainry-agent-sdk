package extraction

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var decimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+|[0-9]{1,3}(?:,[0-9]{3})+)(?:\.[0-9]+)?$`)
var datePattern = regexp.MustCompile(`^([0-9]{4})(?:年|[-/])([0-9]{1,2})(?:月|[-/])([0-9]{1,2})日?$`)

func normalize(raw, kind string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > 1024 || !utf8.ValidString(value) {
		return "", false
	}
	switch kind {
	case "text":
		return value, true
	case "boolean":
		switch strings.ToLower(value) {
		case "true", "是":
			return "true", true
		case "false", "否":
			return "false", true
		}
		return "", false
	case "integer", "decimal":
		if len(value) > 160 || !decimalPattern.MatchString(value) {
			return "", false
		}
		value = strings.ReplaceAll(value, ",", "")
		negative := strings.HasPrefix(value, "-")
		value = strings.TrimPrefix(strings.TrimPrefix(value, "-"), "+")
		parts := strings.Split(value, ".")
		if kind == "integer" && len(parts) > 1 || len(parts[0]) > 128 || len(parts) > 1 && len(parts[1]) > 32 {
			return "", false
		}
		whole := strings.TrimLeft(parts[0], "0")
		if whole == "" {
			whole = "0"
		}
		fraction := ""
		if len(parts) > 1 {
			fraction = strings.TrimRight(parts[1], "0")
		}
		value = whole
		if fraction != "" {
			value += "." + fraction
		}
		if negative && value != "0" {
			value = "-" + value
		}
		return value, true
	case "date":
		parts := datePattern.FindStringSubmatch(value)
		if len(parts) != 4 {
			return "", false
		}
		y, _ := strconv.Atoi(parts[1])
		m, _ := strconv.Atoi(parts[2])
		d, _ := strconv.Atoi(parts[3])
		if y < 1 || m < 1 || m > 12 || d < 1 || d > 31 {
			return "", false
		}
		parsed := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
		if parsed.Year() != y || int(parsed.Month()) != m || parsed.Day() != d {
			return "", false
		}
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d), true
	}
	return "", false
}
