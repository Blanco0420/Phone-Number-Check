package jpnumber

import (
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	"PhoneNumberCheck/source"
	"PhoneNumberCheck/types"
	"PhoneNumberCheck/utils"
	webscraping "PhoneNumberCheck/webScraping"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/tebeka/selenium"
	"github.com/ttacon/libphonenumber"
)

const (
	baseUrl                        = "https://www.jpnumber.com"
	initialPhoneNumberInfoSelector = ".frame-728-orange-l > table:nth-child(2) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1)"
	lineTypeSelector               = ".frame-728-orange-l > table:nth-child(2) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1) > dt:nth-child(3)"
	phoneNumberInfoSelector        = "div.frame-728-green-l:nth-child(2) > div:nth-child(2) > table:nth-child(1) > tbody:nth-child(1)"
	searchSelector                 = "#number"
	commentDateSelector            = ".title-background-pink table tbody tr td:nth-child(2) table  tbody tr td:nth-child(1)"
	commentTextSelector            = "div:nth-child(2) > dt:nth-child(1)"
)

//TODO: Search for the city in the businessDetails Address (for each (rune?token?) check if exists in japaneseInfo)

type JpNumberSource struct {
	driver *webscraping.WebDriverWrapper
}

type rawGraphData struct {
	accessPoints [][]interface{}
	searchPoints [][]interface{}
}

func Initialize() (*JpNumberSource, error) {
	driver, err := webscraping.InitializeDriver()
	if err != nil {
		return &JpNumberSource{}, err
	}
	return &JpNumberSource{
		driver: driver,
	}, nil
}

func (s *JpNumberSource) Close() {
	s.driver.Close()
}

func (s *JpNumberSource) getGraphData(*types.GraphData) error {
	script := `
	var callback = arguments[arguments.length - 1];
var interval = setInterval(() => {
  var script = document.querySelector('script.code');
  if (script) {
    clearInterval(interval);
    callback(script.innerHTML); // this returns to Go
  }
}, 100);
	`
	rawScriptText, err := s.driver.ExecuteScriptAsync(script)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(`(?s)var accessPoints=\[(.*?)\];`)
	scriptText := rawScriptText.(string)
	match := re.FindStringSubmatch(scriptText)

	if len(match) > 1 {
		graph := match[1]
		fmt.Println("AccessPoints data:")
		fmt.Println("[" + graph + "]")
	} else {
		fmt.Println("No match found.")
	}

	scriptText, ok := rawScriptText.(string)
	if !ok || scriptText == "" {
		return fmt.Errorf("expected string from script execution, got nil or non-string value")
	}
	accessPointsRegex := regexp.MustCompile(`var\s+accessPoints\s*=\s*(\[[\s\S]*?\]);`)
	accessMatch := accessPointsRegex.FindStringSubmatch(scriptText)

	if accessMatch == nil {
		return fmt.Errorf("Could not find accessMatch....")
	}

	accessPointsStr := accessMatch[1]
	accessPointsJson := strings.ReplaceAll(accessPointsStr, `'`, `"`)

	var rawGraphData rawGraphData
	if err := json.Unmarshal([]byte(accessPointsJson), &rawGraphData); err != nil {
		return err
	}

	fmt.Println(accessPointsJson)

	graphData := []types.GraphData{}
	for _, row := range rawGraphData.accessPoints {
		fmt.Printf("ROW: %v\n", row)
		if len(row) == 2 {
			date, _ := row[0].(string)
			rawAccesses, _ := row[1].(string)
			parsedDate, err := utils.ParseDate("2006-01-02", date)
			if err != nil {
				return err
			}
			accesses, err := strconv.Atoi(rawAccesses)
			if err != nil {
				return err
			}
			graphData = append(graphData, types.GraphData{Date: parsedDate, Accesses: accesses})
		}
	}

	return nil
}

func (s *JpNumberSource) getNumberParts(number string) ([]string, error) {
	num, err := libphonenumber.Parse(number, "JP")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(libphonenumber.Format(num, libphonenumber.NATIONAL), "-"), nil
}

func (s *JpNumberSource) getDetailsPageURL(lineType source.LineType, number string) (string, error) {
	// https://www.jpnumber.com/ipphone/numberinfo_050_5482_2807.html
	var url = baseUrl
	switch lineType {
	case source.LineTypeMobile:
		url = url + "/mobile"
	case source.LineTypeTollFree:
		url = url + "/freedial"
	case source.LineTypeVOIP:
		url = url + "/ipphone"
	}
	parts, err := s.getNumberParts(number)
	if err != nil {
		return "", err
	}
	return url + fmt.Sprintf("/numberinfo_%s_%s_%s.html", parts[0], parts[1], parts[2]), nil
}

func (s *JpNumberSource) getComments() ([]source.Comment, error) {
	var comments []source.Comment
	commentsContainer, err := s.driver.FindElement("#result-main-right > span:nth-child(6)")
	if err != nil {
		return []source.Comment{}, err
	}
	commentElements, err := commentsContainer.FindElements(selenium.ByCSSSelector, "div.frame-728-gray-l")
	commentElements = commentElements[:len(commentElements)-1]
	if err != nil {
		return []source.Comment{}, err
	}

	for i, elem := range commentElements {
		dateElement, err := elem.FindElement(selenium.ByCSSSelector, commentDateSelector)

		if err != nil {
			fmt.Printf("COMMENT INDEX: %d", i)
			elem, err := json.MarshalIndent(dateElement, "", "  ")
			fmt.Println(string(elem))
			return []source.Comment{}, fmt.Errorf("Error getting date Element: %v", err)
		}
		commentText, err := s.driver.GetInnerText(elem, "div:nth-child(2) > dt:nth-child(1)")
		if err != nil {
			return []source.Comment{}, fmt.Errorf("Comment text error!\n%v\n\n", err)
		}
		dateText, err := s.driver.GetInnerText(elem, ".title-background-pink table tbody tr td:nth-child(2) table  tbody tr td:nth-child(1)")
		if err != nil {
			return []source.Comment{}, fmt.Errorf("Comment date error!\n%v\n\n", err)
		}

		parsedDate, err := utils.ParseDate("2006/01/02 15:04:05", dateText)
		if err != nil {
			return []source.Comment{}, fmt.Errorf("Parsing date error:\n%v\n\n", err)
		}

		comments = append(comments, source.Comment{Comment: commentText, PostDate: parsedDate})

	}
	return comments, nil
}

// func getTextFromTd(row selenium.WebElement) (label string, value string, err error) {
// 	cols, err := row.FindElements(selenium.ByTagName, "td")
// 	if err != nil || len(cols) < 2 {
// 		return "", "", fmt.Errorf("Invalid table row format")
// 	}
// 	label, err = cols[0].Text()
// 	if err != nil {
// 		return "", "", err
// 	}
// 	value, err = cols[1].Text()
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return label, value, nil
// }

func (s *JpNumberSource) getBusinessInfo(businessInfoTableContainer selenium.WebElement, businessDetails *source.BusinessDetails) error {

	_, err := businessInfoTableContainer.FindElement(selenium.ByCSSSelector, "#result-0")
	if err == nil {
		return fmt.Errorf("no business details available")
	}

	businessInfoTableElement, err := businessInfoTableContainer.FindElement(selenium.ByCSSSelector, "div.content > table > tbody")
	if err != nil {
		return err
	}

	tableEntries, err := utils.GetTableInformation(s.driver, businessInfoTableElement, "td", "td")
	if err != nil {
		return err
	}

	// rows, err := businessInfoElementContainer.FindElements(selenium.ByTagName, "tr")
	// if err != nil {
	// 	return err
	// }

	for _, entry := range tableEntries {

		key := entry.Key
		value := entry.Value

		switch key {
		case "Name", "事業者名称":
			businessDetails.Name = value
		case "Industry", "業種":
			businessDetails.Industry = value
		case "Address", "住所":
			japaneseinfo.GetAddressInfo(value, &businessDetails.LocationDetails)
		case "Official website", "公式サイト":
			businessDetails.Website = value
		case "Business", "事業紹介":
			businessDetails.CompanyOverview = value
		}
	}
	return nil
}

func (s *JpNumberSource) GetData(number string) (source.NumberDetails, error) {
	numberQuery := fmt.Sprintf("%s/searchnumber.do?number=%s", baseUrl, number)
	var data source.NumberDetails

	data.Number = number
	var siteInfo source.SiteInfo
	s.driver.GotoUrl(numberQuery)

	//TODO: use the utils getTableInfo function eventually (jpnumber is difficult and doesn't split their table by tr > th,td . Instead, tr >td,td,td,td for like 3 different key:val pairs)

	// Get line type
	initialPhoneNumberInfoContainer, err := s.driver.FindElement(initialPhoneNumberInfoSelector)
	text, err := s.driver.GetInnerText(initialPhoneNumberInfoContainer, "dt:nth-child(3)")
	if err != nil {
		return source.NumberDetails{}, err
	}
	rawLineType := strings.ReplaceAll(strings.Split(text, ">")[0], " ", "")
	lineType, err := utils.GetLineType(rawLineType)
	if err != nil {
		return source.NumberDetails{}, err
	}
	data.LineType = lineType

	// goto detailed page
	detailesPagesUrl, err := s.getDetailsPageURL(lineType, number)
	if err != nil {
		return source.NumberDetails{}, err
	}
	s.driver.GotoUrl(detailesPagesUrl)

	// businessName, err := s.driver.GetInnerText(businessSelector)
	businessInfoTableContainer, err := s.driver.FindElement("div.frame-728-green-l:nth-child(4)")
	if err != nil {
		return source.NumberDetails{}, err
	}
	if err := s.getBusinessInfo(businessInfoTableContainer, &data.BusinessDetails); err != nil {
		if strings.Contains(err.Error(), "no business details available") {
		} else {
			return source.NumberDetails{}, err
		}
	}

	if err := s.getGraphData(&data.GraphData); err != nil {
		return source.NumberDetails{}, err
	}

	//TODO: Move all of this to another function (getNumberMainInfo)
	phoneNumberInfoContainer, err := s.driver.FindElement(phoneNumberInfoSelector)
	if err != nil {
		fmt.Print(err)
		panic("phone number info container doesn't exist??")
	}
	prefecture, _ := s.driver.GetInnerText(phoneNumberInfoContainer, "tr:nth-child(4)>td:nth-child(2)")
	data.BusinessDetails.LocationDetails.Prefecture = prefecture

	carrier, _ := s.driver.GetInnerText(phoneNumberInfoContainer, "tr:nth-child(3)>td:nth-child(4)")
	data.Carrier = carrier

	reviewCount, err := s.driver.GetInnerText(phoneNumberInfoContainer, "span.red")
	if err != nil {
		return source.NumberDetails{}, err
	}
	if reviewCount != "" {
		i, err := strconv.Atoi(reviewCount)
		if err != nil {
			return source.NumberDetails{}, err
		}
		siteInfo.ReviewCount = i
	}

	if siteInfo.ReviewCount >= 1 {
		comments, err := s.getComments()
		if err != nil {
			fmt.Printf("\n\n\n\nCOMMENTS ERROR: %s\n\n\n\n", number)
			return source.NumberDetails{}, err
		}
		siteInfo.Comments = comments
	}

	data.SiteInfo = siteInfo

	return data, nil
}
