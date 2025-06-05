package jpnumber

import (
	"PhoneNumberCheck/source"
	"PhoneNumberCheck/utils"
	webscraping "PhoneNumberCheck/webScraping"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tebeka/selenium"
	"github.com/ttacon/libphonenumber"
)

const (
	baseUrl             = "https://www.jpnumber.com"
	lineTypeSelector    = ".frame-728-orange-l > table:nth-child(2) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > div:nth-child(1) > dt:nth-child(3)"
	carrierSelector     = "div.frame-728-green-l:nth-child(2) > div:nth-child(2) > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(3) > td:nth-child(4)"
	businessSelector    = "div.frame-728-green-l:nth-child(2) > div:nth-child(2) > table:nth-child(1) > tbody:nth-child(1) > tr:nth-child(4) > td:nth-child(4)"
	searchSelector      = "#number"
	commentDateSelector = ".title-background-pink table tbody tr td:nth-child(2) table  tbody tr td:nth-child(1)"
	commentTextSelector = "div:nth-child(2) > dt:nth-child(1)"
)

type JpNumberSource struct {
	driver *webscraping.WebDriverWrapper
}

func Initialize(driver *webscraping.WebDriverWrapper) *JpNumberSource {
	return &JpNumberSource{
		driver: driver,
	}
}

func (s *JpNumberSource) getLineType() (source.LineType, error) {

	text, err := s.driver.GetInnerText(lineTypeSelector)
	if err != nil {
		return source.LineTypeUnknown, err
	}
	lineType := strings.ReplaceAll(strings.Split(text, ">")[0], " ", "")
	switch lineType {
	case "固定電話":
		return source.LineTypeLandline, nil
	case "携帯電話":
		return source.LineTypeMobile, nil
	case "フリーダイヤル":
		return source.LineTypeTollFree, nil
	case "IP電話":
		return source.LineTypeVOIP, nil
	default:
		return source.LineTypeUnknown, nil
	}
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
		/*
			TODO: This
							Fix: Use More Robust and Readable Selector

							Here’s a much simpler and resilient selector for the date inside the div.frame-728-gray-l block:

							commentDateSelector := "td > table td:first-child"

							If you want to scope it just a bit more strictly (but still reliably), do:

							commentDateSelector := ".title-background-pink td > table td:first-child"
		*/
		if i == 58 {
			fmt.Print("Here")
		}
		if err != nil {
			fmt.Printf("COMMENT INDEX: %d", i)
			elem, err := json.MarshalIndent(dateElement, "", "  ")
			fmt.Println(string(elem))
			return []source.Comment{}, fmt.Errorf("Error getting date Element: %v", err)
		}
		commentElement, err := elem.FindElement(selenium.ByCSSSelector, commentTextSelector)
		if err != nil {
			return []source.Comment{}, fmt.Errorf("Error getting comment text element: %v", err)
		}
		commentText, err := s.driver.GetInnerTextFromElement(commentElement)
		if err != nil {
			return []source.Comment{}, fmt.Errorf("Comment text error!\n%v\n\n", err)
		}
		dateText, err := s.driver.GetInnerTextFromElement(dateElement)
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

func getTextFromTd(row selenium.WebElement) (label string, value string, err error) {
	cols, err := row.FindElements(selenium.ByTagName, "td")
	if err != nil || len(cols) < 2 {
		return "", "", fmt.Errorf("Invalid table row format")
	}
	label, err = cols[0].Text()
	if err != nil {
		return "", "", err
	}
	value, err = cols[1].Text()
	if err != nil {
		return "", "", err
	}
	return label, value, nil
}

func (s *JpNumberSource) getBusinessInfo(businessContainer selenium.WebElement) (source.BusinessDetails, error) {

	_, err := businessContainer.FindElement(selenium.ByCSSSelector, "#result-0")
	if err == nil {
		return source.BusinessDetails{}, fmt.Errorf("no business details available")
	}
	var businessDetails source.BusinessDetails

	businessInfoElementContainer, err := businessContainer.FindElement(selenium.ByCSSSelector, "div:nth-child(2)")
	if err != nil {
		return source.BusinessDetails{}, err
	}
	rows, err := businessInfoElementContainer.FindElements(selenium.ByTagName, "tr")
	if err != nil {
		return businessDetails, err
	}
	for _, row := range rows {
		label, value, err := getTextFromTd(row)
		if err != nil {
			fmt.Printf("ROW ERROR: \n%v\nLABEL: %s\nVALUE: %s", err, label, value)
			continue
		}
		switch label {
		case "Name", "事業者名称":
			businessDetails.Name = value

		case "Industry", "業種":
			businessDetails.Industry = value
		case "Address", "住所":
			businessDetails.Address = value
		case "Official website", "公式サイト":
			businessDetails.Website = value
		case "Business", "事業紹介":
			businessDetails.CompanyOverview = value
		}

	}
	return businessDetails, nil
}

func (s *JpNumberSource) GetData(number string) (source.NumberInfo, error) {
	numberQuery := fmt.Sprintf("%s/searchnumber.do?number=%s", baseUrl, number)
	var data source.NumberInfo

	data.Number = number
	var siteInfo source.SiteInfo
	s.driver.GotoUrl(numberQuery)

	//Check line type exists and return line type
	// if exists := s.driver.CheckElementExists(lineTypeSelector); exists {
	lineType, err := s.getLineType()
	// if err != nil {
	// 	return source.PhoneData{}, err
	// }
	data.LineType = lineType
	// }

	// goto detailed page
	detailesPagesUrl, err := s.getDetailsPageURL(lineType, number)
	if err != nil {
		return source.NumberInfo{}, err
	}
	s.driver.GotoUrl(detailesPagesUrl)

	// businessName, err := s.driver.GetInnerText(businessSelector)
	businessContainer, err := s.driver.FindElement("div.frame-728-green-l:nth-child(4)")
	if err != nil {
		return source.NumberInfo{}, err
	}
	businessInfo, err := s.getBusinessInfo(businessContainer)
	if err != nil {
		if !strings.Contains(err.Error(), "no business details available") {
			return source.NumberInfo{}, err
		}
	}
	data.BusinessInfo = businessInfo

	reviewCount, err := s.driver.GetInnerText("span.red")
	if err != nil {
		return source.NumberInfo{}, err
	}
	if reviewCount != "" {
		i, err := strconv.Atoi(reviewCount)
		if err != nil {
			return source.NumberInfo{}, err
		}
		siteInfo.ReviewCount = i
	}

	if siteInfo.ReviewCount >= 1 {
		comments, err := s.getComments()
		if err != nil {
			fmt.Printf("\n\n\n\nCOMMENTS ERROR: %s\n\n\n\n", number)
			return source.NumberInfo{}, err
		}
		siteInfo.Comments = comments
	}

	data.SiteInfo = siteInfo

	return data, nil
}
