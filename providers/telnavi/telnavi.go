package telnavi

import (
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	"PhoneNumberCheck/source"
	"PhoneNumberCheck/utils"
	webscraping "PhoneNumberCheck/webScraping"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tebeka/selenium"
)

const (
	baseUrl = "https://www.telnavi.jp/phone"
	//NOTE: 推定発信地域 estimated outgoing area (maybe use later)
)

type TelnaviSource struct {
	driver *webscraping.WebDriverWrapper
}

type tableEntry struct {
	key     string
	value   string
	element selenium.WebElement
}

func Initialize() (*TelnaviSource, error) {
	driver, err := webscraping.InitializeDriver()
	if err != nil {
		return &TelnaviSource{}, err
	}
	return &TelnaviSource{
		driver: driver,
	}, nil
}

func (t *TelnaviSource) Close() {
	t.driver.Close()
}

func (t *TelnaviSource) calculateFraudScore(ratingTableContainer selenium.WebElement) (int, error) {
	tableRows, err := ratingTableContainer.FindElement(selenium.ByCSSSelector, "td > table > tbody")
	if err != nil {
		panic(fmt.Errorf("Row exists but not the info?? : %v", err))
	}
	values, err := tableRows.FindElements(selenium.ByCSSSelector, "tr")
	if err != nil {
		panic("IDEK")
	}
	percentageRegex := regexp.MustCompile(`(\d+)%`)
	var ratings []int
	for _, val := range values {
		rawString, err := t.driver.GetInnerText(val, "td:nth-child(1)")
		if err != nil {
			panic(fmt.Errorf("Could not get the inner text of fraud score? : %v", err))
		}
		matches := percentageRegex.FindStringSubmatch(rawString)
		if len(matches) < 2 {
			panic(fmt.Errorf("matches len is < 0 : %v", matches))
		} else if len(matches) > 2 {
			panic(fmt.Errorf("matches length is more than 2???: %v", matches))
		}
		score, err := strconv.Atoi(matches[1])
		if err != nil {
			panic(fmt.Errorf("Errr formatting decimal: %v", err))
		}
		ratings = append(ratings, score)
	}
	fraudScore := ratings[2] + ratings[1]/2
	return fraudScore, nil
}

func extractBusinessName(input string) string {
	cleaned := strings.TrimSpace(input)

	if cleaned == "" {
		return ""
	}

	parts := strings.Split(cleaned, "\n")

	if len(parts) == 2 {
		return parts[0]
	} else if len(parts) == 1 {
		probablyNum := parts[0]
		if probablyNum[0] == '"' && probablyNum[len(probablyNum)-1] == '"' {
			return ""
		}
	} else {
		panic(fmt.Errorf("Business name has multiple parts?: %v", input))
	}
	return ""
}

func (t *TelnaviSource) getPhoneNumberInfo(data *source.NumberDetails, tableEntries []tableEntry) error {
	for _, entry := range tableEntries {
		key := entry.key
		val := entry.value
		switch key {
		case "事業者名":
			if data.BusinessDetails.Name == "" {
				data.BusinessDetails.Name = extractBusinessName(val)
			}
		case "住所":
			if data.BusinessDetails.LocationDetails.Address == "" {
				addressInfo, err := getAddressInfo(val)
				data.BusinessDetails.LocationDetails = addressInfo
			}
		case "回線種別":
			lineType, err := utils.GetLineType(val)
			if err != nil {
				fmt.Println("Error, failed to get line type: ", err)
			}
			data.LineType = lineType
		case "業種タグ":
			data.BusinessDetails.Industry = val
		case "ユーザー評価":
			rating, err := getCleanRating(val)
			if err != nil {
				fmt.Println("Error, failed to get clean rating: ", err)
			}
			data.SiteInfo.UserRating = rating
		case "アクセス数":
			re := regexp.MustCompile(`[^0-9]`)
			cleanedAccessCount := re.ReplaceAllString(val, "")
			accessCount, err := strconv.Atoi(cleanedAccessCount)
			if err != nil {
				fmt.Printf("CleanedAccessCount: %s\naccessedCount: %d", cleanedAccessCount, accessCount)
				panic(fmt.Errorf("failed to parse access count: %v", err))
			}
			data.SiteInfo.AccessCount = accessCount
		case "迷惑電話度":
			fraudScore, err := t.calculateFraudScore(entry.element)
			if err != nil {
				if strings.Contains(err.Error(), "no such element") {
					data.FraudScore = 0
				} else {
					return err
				}
			} else {
				data.FraudScore = fraudScore
			}
		default:
			continue

		}
	}
	return nil
}

func (t *TelnaviSource) getBusinessInfo(businessDetails *source.BusinessDetails, businessTableEntries []tableEntry) error {
	//TODO: Check if doesn't exist
	for _, entry := range businessTableEntries {
		key := entry.key
		val := entry.value

		switch key {
		case "事業者名":
			businessDetails.Name = val
		case "住所":
			if err := japaneseinfo.GetAddressInfo(val, &businessDetails.LocationDetails); err != nil {
				return err
			}
			// postcode, exists := utils.FindPostcodeInText(val)
			// postcode = "1138654"
			// exists = true
			// if exists {
			// 	address, err := japaneseinfo.GetJapaneseInfoFromPostcode(postcode)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	businessDetails.LocationDetails.City = address.City
			// 	businessDetails.LocationDetails.Prefecture = address.Prefecture
			// 	businessDetails.LocationDetails.PostCode = address.PostCode
			// 	val = strings.Replace(val, postcode, "", 1)
			// 	val = utils.CleanText(val)
			// }
			// businessDetails.LocationDetails.Address = val
		}
	}

	return nil
}

// func (t *TelnaviSource) getElementOnTableWhereKeyEquals(phoneNumberInfoContainerRowElements []selenium.WebElement, key string) (selenium.WebElement, error) {
//
// 	var returnElement selenium.WebElement
// 	for _, element := range phoneNumberInfoContainerRowElements {
// 		text, err := t.driver.GetInnerText(element, "th")
// 		if err != nil {
// 			continue
// 			//TODO: Fix this?
// 		}
// 		if text == key {
// 			returnElement = element
// 			break
// 		}
// 	}
// 	if returnElement == nil {
// 		return returnElement, fmt.Errorf("element not found")
// 	}
// 	return returnElement, nil
// }

func getCleanRating(rawUserRating string) (float32, error) {
	var rating float32
	re := regexp.MustCompile(`[^0-9.]`)
	cleaned := re.ReplaceAllString(rawUserRating, "")
	if cleaned == "" {
		rating = 0
	} else {
		f64UserRating, err := strconv.ParseFloat(cleaned, 32)
		if err != nil {
			return 0, err
		}
		userRating := float32(f64UserRating)
		rating = userRating

	}
	return rating, nil
}

func (t *TelnaviSource) getTableInformation(tableBodyElement selenium.WebElement) ([]tableEntry, error) {
	var tableEntries []tableEntry
	ignoredTableKeys := []string{"初回クチコミユーザー", "FAX番号", "市外局番", "市内局番", "加入者番号", "電話番号", "推定発信地域"}
	phoneNumberTableContainerRowElements, err := tableBodyElement.FindElements(selenium.ByCSSSelector, "tr")
	if err != nil {
		panic(fmt.Errorf("Could not get phone number info table rows: %v", err))
	}
	for _, element := range phoneNumberTableContainerRowElements {
		key, err := t.driver.GetInnerText(element, "th")
		if err != nil {
			continue
			//TODO: Fix this?
		}
		value, err := t.driver.GetInnerText(element, "td")
		if err != nil {
			return tableEntries, err
		}
		for _, v := range ignoredTableKeys {
			if key == v {
				continue
			}
		}
		tableEntries = append(tableEntries, tableEntry{key: key, value: value, element: element})
	}
	return tableEntries, nil
}

func (t *TelnaviSource) GetData(phoneNumber string) (source.NumberDetails, error) {
	var data source.NumberDetails
	data.Number = phoneNumber
	numberQuery := fmt.Sprintf("%s/%s", baseUrl, phoneNumber)
	t.driver.GotoUrl(numberQuery)

	businessTableContainer, err := t.driver.FindElement("div.info_table:nth-child(1) > table > tbody:nth-child(1)")
	if err != nil {
		if strings.Contains(err.Error(), "no such element") {
		} else {
			return source.NumberDetails{}, err
		}
	} else {
		businessTableEntries, err := t.getTableInformation(businessTableContainer)
		if err != nil {
			return source.NumberDetails{}, err
		}
		if err := t.getBusinessInfo(&data.BusinessDetails, businessTableEntries); err != nil {
			return source.NumberDetails{}, err
		}
	}
	phoneNumberTableContainer, err := t.driver.FindElement("div.info_table:nth-child(2) > table > tbody")
	if err != nil {
		return source.NumberDetails{}, err
	}

	phoneNumberTableEntries, err := t.getTableInformation(phoneNumberTableContainer)
	if err != nil {
		return source.NumberDetails{}, err
	}
	if err := t.getPhoneNumberInfo(&data, phoneNumberTableEntries); err != nil {
		return source.NumberDetails{}, err
	}

	// businessInfoContainer, err = businessInfoContainer.FindElement(selenium.ByCSSSelector, "table:nth-child(1) > tbody:nth-child(1)")
	// if err != nil {
	// 	return err
	// }

	// rawUserRating, err := t.driver.GetInnerText(phoneNumberInfoContainer, "tr:nth-child(13) > td")
	// if err != nil {
	// 	return source.NumberDetails{}, err
	// }

	userCommentsContainer, err := t.driver.FindElement("div.kuchikomi_thread_content")
	if err != nil {
		return source.NumberDetails{}, err
	}
	commentsElements, err := userCommentsContainer.FindElements(selenium.ByCSSSelector, "#thread")
	if err != nil {
		panic(fmt.Errorf("no comments?: %v", err))
	}
	data.SiteInfo.ReviewCount = len(commentsElements)

	for _, elem := range commentsElements {
		var comment source.Comment

		tableBody, err := elem.FindElement(selenium.ByCSSSelector, "tbody")
		if err != nil {
			panic(fmt.Errorf("Couldn't get comment table body? %v", err))
		}
		dateElement, err := tableBody.FindElement(selenium.ByCSSSelector, "tr:nth-child(1) > td:nth-child(1) > time:nth-child(1)")
		if err != nil {
			panic(fmt.Errorf("Failed to get date element: %v", err))
		}
		dateString, err := dateElement.GetAttribute("content")
		if err != nil {
			panic(fmt.Errorf("Failed to get content attr. from date elem: %v", err))
		}
		formattedDate, err := utils.ParseDate("2006-01-02", dateString)
		if err != nil {
			panic(fmt.Errorf("Failed to parse date: %v", err))
		}
		comment.PostDate = formattedDate

		commentText, err := t.driver.GetInnerText(tableBody, "tr:nth-child(2) > td > div")
		if err != nil {
			panic(fmt.Errorf("Failed to get comment text: %v", err))
		}
		comment.Comment = commentText
		data.SiteInfo.Comments = append(data.SiteInfo.Comments, comment)
	}

	// data.BusinessDetails.

	return data, nil
}
