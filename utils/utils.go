package utils

import (
	"PhoneNumberCheck/source"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

/*
01 - month
02 - day
03 - hour (12h)
04 - minute
05 - second
06 - year
07 - time zone offset
*/
func ParseDate(layout, date string) (time.Time, error) {
	parsedDate, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}

func CleanText(text string) string {
	re := regexp.MustCompile(`[^\p{L}\p{N}一]`)
	return re.ReplaceAllString(text, "")
}

func GetLineType(rawLineType string) (source.LineType, error) {
	rawLineType = strings.TrimSpace(strings.ToLower(rawLineType))
	rawLineType = strings.Trim(rawLineType, "-_")

	switch rawLineType {
	case "landline", "固定電話":
		return source.LineTypeLandline, nil
	case "mobile", "携帯電話":
		return source.LineTypeMobile, nil
	case "tollfree", "freedial", "フリーダイヤル":
		return source.LineTypeTollFree, nil
	case "voip", "IP電話":
		return source.LineTypeVOIP, nil
	case "premiumrate", "paging", "satellite":
		return source.LineTypeOther, fmt.Errorf("Line type is of type other. Actual text: %s", rawLineType)
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
