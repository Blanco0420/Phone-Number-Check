package utils

import (
	"PhoneNumberCheck/source"
	"os"
	"strings"
	"time"
)

func ParseDate(layout, date string) (time.Time, error) {
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}
func GetLineType(rawLineType string) (source.LineType, error) {
	rawLineType = strings.TrimSpace(strings.ToLower(rawLineType))
	switch rawLineType {
	case "landline", "固定電話":
		return source.LineTypeLandline, nil
	case "mobile", "携帯電話":
		return source.LineTypeMobile, nil
	case "toll-free", "tollfree", "freedial", "フリーダイヤル":
		return source.LineTypeTollFree, nil
	case "voip", "IP電話":
		return source.LineTypeVOIP, nil
	default:
		return source.LineTypeUnknown, nil
	}
}

func CheckIfFileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}
