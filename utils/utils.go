package utils

import (
	"PhoneNumberCheck/source"
	"PhoneNumberCheck/types"
	webscraping "PhoneNumberCheck/webScraping"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/tebeka/selenium"
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

var lineTypeMap = map[string]source.LineType{
	"landline": source.LineTypeLandline,
	"固定電話":     source.LineTypeLandline,

	"mobile": source.LineTypeMobile,
	"携帯電話":   source.LineTypeMobile,

	"tollfree": source.LineTypeTollFree,
	"freedial": source.LineTypeTollFree,
	"フリーダイヤル":  source.LineTypeTollFree,

	"voip": source.LineTypeVOIP,
	"ip電話": source.LineTypeVOIP,
}

var lineTypeOtherMap = map[string]struct{}{
	"premiumrate": {},
	"paging":      {},
	"satellite":   {},
}

func GetLineType(rawLineType string) (source.LineType, error) {

	cleaned := strings.TrimSpace(strings.ToLower(rawLineType))
	cleaned = strings.Trim(rawLineType, "-_")
	if val, ok := lineTypeMap[cleaned]; ok {
		return val, nil
	}

	if _, ok := lineTypeOtherMap[cleaned]; ok {
		return source.LineTypeOther, fmt.Errorf("Line type is of type other. Actual text: %s", rawLineType)
	}
	return source.LineTypeOther, nil
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

func GetTableInformation(d *webscraping.WebDriverWrapper, tableBodyElement selenium.WebElement, tableKeyElementTagName string, tableValueElementTagName string) ([]types.TableEntry, error) {
	var tableEntries []types.TableEntry
	ignoredTableKeys := []string{"初回クチコミユーザー", "FAX番号", "市外局番", "市内局番", "加入者番号", "電話番号", "推定発信地域"}
	phoneNumberTableContainerRowElements, err := tableBodyElement.FindElements(selenium.ByCSSSelector, "tr")
	if err != nil {
		panic(fmt.Errorf("Could not get phone number info table rows: %v", err))
	}

	if tableKeyElementTagName == tableValueElementTagName {
		tableKeyElementTagName = tableKeyElementTagName + "nth-child(1)"
		tableValueElementTagName = tableValueElementTagName + "nth-child(2)"
	}

	for _, element := range phoneNumberTableContainerRowElements {
		key, err := d.GetInnerText(element, tableKeyElementTagName)
		if err != nil {
			continue
			//TODO: Fix this?
		}
		value, err := d.GetInnerText(element, tableValueElementTagName)
		if err != nil {
			return tableEntries, err
		}
		for _, v := range ignoredTableKeys {
			if key == v {
				continue
			}
		}
		tableEntries = append(tableEntries, types.TableEntry{Key: key, Value: value, Element: element})
	}
	return tableEntries, nil
}
