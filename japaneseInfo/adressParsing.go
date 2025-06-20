package japaneseinfo

import (
	"PhoneNumberCheck/providers"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	jpostcode "github.com/syumai/go-jpostcode"
)

type jsonAddress struct {
	success     bool
	prefecture  string
	city        string
	town        string
	chome       string
	ban         string
	GoField     string `json:"go"`
	left        string
	errorString string
}

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

func parseAddress(address string) (jsonAddress, error) {
	//TODO: when docker container, add binary to path
	var jsonAddress jsonAddress
	cmd := exec.Command("./parseAddress", address)
	out, err := cmd.Output()
	if err != nil {
		return jsonAddress, err
	}
	if err := json.Unmarshal(out, &jsonAddress); err != nil {
		return jsonAddress, err
	}
	if !jsonAddress.success {
		return jsonAddress, fmt.Errorf("error parsing address %s: %v", address, jsonAddress.errorString)
	}

	return jsonAddress, nil
}

func GetAddressInfo(address string, locationDetails *providers.LocationDetails) error {
	if strings.TrimSpace(address) == "" {
		return nil
	}
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

		addressInfo, err := parseAddress(address)
		if err != nil {
			return err
		}
		fmt.Printf("Address: %v", addressInfo)
		locationDetails.Prefecture = addressInfo.prefecture
		locationDetails.City = addressInfo.city
		locationDetails.Address = address
	}
	return nil
}
