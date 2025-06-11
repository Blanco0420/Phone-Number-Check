package japaneseinfo

import (
	"PhoneNumberCheck/source"
	"fmt"
	"regexp"

	jpostcode "github.com/syumai/go-jpostcode"
)

func getJapaneseInfoFromPostcode(postcode string) (*jpostcode.Address, error) {
	address, err := jpostcode.Find(postcode)
	if err != nil {
		return address, err
	}
	return address, nil
}

func findPostcodeInText(text string) (string, bool) {
	re := regexp.MustCompile(`\b\d{3}-\d{4}\b`)
	matches := re.FindAllString(text, -1)
	if len(matches) < 1 {
		return "", false
	} else if len(matches) > 1 {
		panic(fmt.Errorf("MULTIPLE POSTCODES FOUND: %s", text))
	}
	return matches[0], true
}

func GetAddressInfo(address string, locationDetails *source.LocationDetails) error {
	if postcode, exists := findPostcodeInText(address); exists {
		addressInfo, err := getJapaneseInfoFromPostcode(postcode)
		if err != nil {
			return err
		}
		locationDetails.Prefecture = addressInfo.Prefecture
		locationDetails.City = addressInfo.City
		//TODO: Also parse and send back here
		locationDetails.Address = address

	} else {
		//TODO: Parse address info
	}
	return nil
}
