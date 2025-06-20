package telnavi

import (
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	providerdataprocessing "PhoneNumberCheck/providerDataProcessing"
	"PhoneNumberCheck/providers"
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
	driver           *webscraping.WebDriverWrapper
	vitalInfoChannel chan providers.VitalInfo
	currentVitalInfo *providers.VitalInfo
}

func (t *TelnaviSource) VitalInfoChannel() <-chan providers.VitalInfo {
	return t.vitalInfoChannel
}

func (t *TelnaviSource) CloseVitalInfoChannel() {
	close(t.vitalInfoChannel)
}

func Initialize() (*TelnaviSource, error) {
	driver, err := webscraping.InitializeDriver()
	if err != nil {
		fmt.Println("here")
		return &TelnaviSource{}, err
	}
	return &TelnaviSource{
		driver:           driver,
		vitalInfoChannel: make(chan providers.VitalInfo),
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
		panic(fmt.Errorf("business name has multiple parts?: %v", input))
	}
	return ""
}

func (t *TelnaviSource) getGraphData(graphData *[]providers.GraphData) error {
	script := `
var callback = arguments[arguments.length - 1];
var interval = setInterval(() => {
  if (window.pageview_stat) {
    clearInterval(interval);
    callback(JSON.stringify(window.pageview_stat)); // return JSON string of the object
  }
}, 100);`

	rawData, err := t.driver.ExecuteScriptAsync(script)
	if err != nil {
		return err
	}

	rawDataString, ok := rawData.(string)
	if !ok {
		return fmt.Errorf("unexpected graph data %v of type %T", rawData, rawData)
	}

	if err := utils.ParseGraphData(rawDataString, graphData); err != nil {
		return err
	}
	return nil
}

/*
TODO:

- CurrentVitalInfo for each provider.
- Is an address which I set to the address of the instantiated data's vitalInfo for each provider


*/

func (t *TelnaviSource) getPhoneNumberInfo(data *providers.NumberDetails, tableEntries []providers.TableEntry) error {
	for _, entry := range tableEntries {
		key := entry.Key
		val := entry.Value
		switch key {
		case "事業者名":
			if t.currentVitalInfo.Name == "" {
				cleanName := extractBusinessName(val)
				cleanName, suffixes := utils.CleanCompanyName(cleanName)
				data.BusinessDetails.NameSuffixes = suffixes
				t.currentVitalInfo.Name = cleanName
				t.vitalInfoChannel <- *t.currentVitalInfo
			}
		case "住所":
			if data.BusinessDetails.LocationDetails == (providers.LocationDetails{}) {
				if err := japaneseinfo.GetAddressInfo(val, &data.BusinessDetails.LocationDetails); err != nil {
					return err
				}
			}
		case "回線種別":

			lineType, err := utils.GetLineType(val)
			if err != nil {
				fmt.Println("Error, failed to get line type: ", err)
			}
			t.currentVitalInfo.LineType = lineType
			t.vitalInfoChannel <- *t.currentVitalInfo
		case "業種タグ":
			t.currentVitalInfo.Industry = val
			t.vitalInfoChannel <- *t.currentVitalInfo
		case "ユーザー評価":
			rating, err := getCleanRating(val)
			if err != nil {
				fmt.Println("Error, failed to get clean rating: ", err)
			}
			data.SiteInfo.UserRating = rating
		case "アクセス数":
			val = strings.TrimSpace(val)
			if val == "10回未満" {

			}
			re := regexp.MustCompile(`[^0-9]`)
			cleanedAccessCount := re.ReplaceAllString(val, "")
			accessCount, err := strconv.Atoi(cleanedAccessCount)
			if err != nil {
				fmt.Printf("CleanedAccessCount: %s\naccessedCount: %d", cleanedAccessCount, accessCount)
				panic(fmt.Errorf("failed to parse access count: %v", err))
			}
			data.SiteInfo.AccessCount = accessCount
		case "迷惑電話度":
			//TODO: Channel
			fraudScore, err := t.calculateFraudScore(entry.Element)
			if err != nil {
				if strings.Contains(err.Error(), "no such element") {
					t.currentVitalInfo.FraudulentDetails.FraudScore = 0
					t.vitalInfoChannel <- *t.currentVitalInfo
				} else {
					return err
				}
			} else {
				t.currentVitalInfo.FraudulentDetails.FraudScore = fraudScore
				t.vitalInfoChannel <- *t.currentVitalInfo
			}
		default:
			continue

		}
	}
	return nil
}

func (t *TelnaviSource) getBusinessInfo(data *providers.NumberDetails, businessTableEntries []providers.TableEntry) error {
	//TODO: Check if doesn't exist
	for _, entry := range businessTableEntries {
		key := entry.Key
		val := entry.Value

		switch key {
		case "事業者名":
			cleanName := extractBusinessName(val)
			cleanName, suffixes := utils.CleanCompanyName(cleanName)
			data.BusinessDetails.NameSuffixes = suffixes
			t.currentVitalInfo.Name = cleanName
			t.vitalInfoChannel <- *t.currentVitalInfo
			//TODO: Channel
		case "住所":
			if err := japaneseinfo.GetAddressInfo(val, &data.BusinessDetails.LocationDetails); err != nil {
				return err
			}
		}
	}

	return nil
}

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

func (t *TelnaviSource) getUserCommentsContainer() (selenium.WebElement, error) {

	userCommentsContainer, err := t.driver.FindElement("div.kuchikomi_thread_content")
	if err != nil {
		return userCommentsContainer, err
	}
	return userCommentsContainer, nil
}

func (t *TelnaviSource) GetData(phoneNumber string) (providers.NumberDetails, error) {
	var data providers.NumberDetails
	t.currentVitalInfo = &data.VitalInfo
	data.Number = phoneNumber
	phoneNumberInfoPageUrl := fmt.Sprintf("%s/%s", baseUrl, phoneNumber)
	t.driver.GotoUrl(phoneNumberInfoPageUrl)

	businessTableContainer, err := t.driver.FindElement("div.info_table:nth-child(1) > table > tbody:nth-child(1)")
	if err != nil {
		if strings.Contains(err.Error(), "no such element") {
		} else {
			return providers.NumberDetails{}, err
		}
	} else {
		businessTableEntries, err := utils.GetTableInformation(t.driver, businessTableContainer, "th", "td")
		if err != nil {
			return providers.NumberDetails{}, err
		}
		if err := t.getBusinessInfo(&data, businessTableEntries); err != nil {
			return providers.NumberDetails{}, err
		}
	}
	phoneNumberTableContainer, err := t.driver.FindElement("div.info_table:nth-child(2) > table > tbody")
	if err != nil {
		return providers.NumberDetails{}, err
	}

	phoneNumberTableEntries, err := utils.GetTableInformation(t.driver, phoneNumberTableContainer, "th", "td")
	if err != nil {
		return providers.NumberDetails{}, err
	}
	if err := t.getPhoneNumberInfo(&data, phoneNumberTableEntries); err != nil {
		return providers.NumberDetails{}, err
	}

	// businessInfoContainer, err = businessInfoContainer.FindElement(selenium.ByCSSSelector, "table:nth-child(1) > tbody:nth-child(1)")
	// if err != nil {
	// 	return err
	// }

	// rawUserRating, err := t.driver.GetInnerText(phoneNumberInfoContainer, "tr:nth-child(13) > td")
	// if err != nil {
	// 	return providers.NumberDetails{}, err
	// }
	userCommentsContainer, err := t.getUserCommentsContainer()
	if err != nil {
		return providers.NumberDetails{}, err
	}

	paginationControlElement, err := userCommentsContainer.FindElement(selenium.ByCSSSelector, "div.paginationControl")
	if err != nil {
		panic(fmt.Errorf("Error comments have no pagination element: %v", err))
	}
	spans, err := paginationControlElement.FindElements(selenium.ByTagName, "span")
	if err != nil {
		panic(fmt.Errorf("Error, pagination has no span elements"))
	}
	pages := []int{1}
	if len(spans) < 2 {
		links, err := paginationControlElement.FindElements(selenium.ByTagName, "a")
		if err != nil {
			panic(fmt.Errorf("can't find any link elements: %v", err))
		}
		for _, elem := range links {
			rawPageNumber, err := elem.Text()
			if err != nil {
				panic(fmt.Errorf("Couldn't get page number text: %v", err))
			}
			parsedPageNumber, err := strconv.Atoi(rawPageNumber)
			if err != nil {
				fmt.Printf("Couldn't parse %s into int: ", rawPageNumber)
				continue
			}
			pages = append(pages, parsedPageNumber)
		}
	}
	reviewCount := 0

	for _, pageNumber := range pages {
		if pageNumber != 1 {
			t.driver.GotoUrl(fmt.Sprintf("%s?page=%d", phoneNumberInfoPageUrl, pageNumber))
		}
		//TODO: Make this into a function. Pretty much make everything comment wise into separated functions. And maybe later for re-usability on other providers
		userCommentsContainer, err = t.getUserCommentsContainer()
		if err != nil {
			return providers.NumberDetails{}, err
		}
		commentsElements, err := userCommentsContainer.FindElements(selenium.ByCSSSelector, "#thread")
		if err != nil {
			panic(fmt.Errorf("no comments?: %v", err))
		}

		reviewCount += len(commentsElements)
		for _, elem := range commentsElements {
			var comment providers.Comment

			tableBody, err := elem.FindElement(selenium.ByCSSSelector, "tbody")
			if err != nil {
				panic(fmt.Errorf("Couldn't get comment table body? %v", err))
			}
			dateElement, err := tableBody.FindElement(selenium.ByCSSSelector, "tr:nth-child(1) > td:nth-child(1) > time:nth-child(1)")
			if err != nil {
				panic(fmt.Errorf("Failed to get date element: %v", err))
			}
			dateString, err := dateElement.Text()
			if err != nil {
				panic(fmt.Errorf("Failed to get content attr. from date elem: %v", err))
			}
			formattedDate, err := utils.ParseDate("2006年1月2日 15時4分", dateString)
			if err != nil {
				panic(fmt.Errorf("Failed to parse date: %v", err))
			}
			comment.PostDate = formattedDate

			commentText, err := t.driver.GetInnerText(tableBody, "tr:nth-child(2) > td > div")
			if err != nil {
				panic(fmt.Errorf("Failed to get comment text: %v", err))
			}
			comment.Text = commentText
			data.SiteInfo.Comments = append(data.SiteInfo.Comments, comment)
		}
	}
	data.SiteInfo.ReviewCount = reviewCount

	var graphData []providers.GraphData
	if err := t.getGraphData(&graphData); err != nil {
		return data, err
	}
	numberRiskInput := providerdataprocessing.NumberRiskInput{
		SourceName:  "telnavi",
		GraphData:   graphData,
		RecentAbuse: &data.VitalInfo.FraudulentDetails.RecentAbuse,
		FraudScore:  &data.VitalInfo.FraudulentDetails.FraudScore,
		Comments:    data.SiteInfo.Comments,
	}
	overallFraudScore := providerdataprocessing.EvaluateSource(numberRiskInput)
	t.currentVitalInfo.OverallFraudScore = overallFraudScore
	t.vitalInfoChannel <- *t.currentVitalInfo
	return data, nil
}
