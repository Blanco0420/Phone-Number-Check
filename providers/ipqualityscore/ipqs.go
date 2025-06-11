package ipqualityscore

import (
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	"PhoneNumberCheck/source"
	"PhoneNumberCheck/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type IpqsSource struct {
	config *source.APIConfig
}

type rawApiData struct {
	Success      bool
	Valid        bool
	FraudScore   int  `json:"fraud_score"`
	RecentAbuse  bool `json:"recent_abuse"`
	Risky        bool
	Active       bool
	Carrier      string
	LineType     string `json:"line_type"`
	City         string
	PostCode     string `json:"zip_code"`
	Region       string
	Name         string
	IdentityData any `json:"identity_data"`
	Spammer      bool
	ActiveStatus string `json:"active_status"`
	Errors       []string
}

func Initialize() (*IpqsSource, error) {
	apiKey, exists := os.LookupEnv("PROVIDERS__IPQS_API_KEY")
	if !exists {
		return &IpqsSource{}, fmt.Errorf("Error, apiKey environment variable not set")
	}
	baseUrl := "https://www.ipqualityscore.com/api/json/phone/" + apiKey + "/<NUMBER>?country[]=JP"
	config := source.NewApiConfig(apiKey, baseUrl)
	return &IpqsSource{config: config}, nil
}
func (i *IpqsSource) GetData(phoneNumber string) (source.NumberDetails, error) {
	requestUrl := strings.Replace(i.config.BaseUrl, "<NUMBER>", phoneNumber, 1)
	res, err := i.config.HttpClient.Get(requestUrl)
	if err != nil {
		return source.NumberDetails{}, fmt.Errorf("Error: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return source.NumberDetails{}, err
	}

	var rawData rawApiData
	if err := json.Unmarshal(body, &rawData); err != nil {
		return source.NumberDetails{}, err
	}

	if !rawData.Success {
		return source.NumberDetails{}, fmt.Errorf("Error getting data from source:\n%v", rawData.Errors)
	}

	switch v := rawData.IdentityData.(type) {
	case string:
		fmt.Println("identityData is string", v)
		panic("identityData")
	case []any:
		fmt.Println("identityData is array", v)
		panic("identityData is array")
	}

	pref, _ := japaneseinfo.FindPrefectureByCityName(rawData.City, 2)
	locationDetails := source.LocationDetails{
		City:       rawData.City,
		Prefecture: pref,
	}

	businessDetails := source.BusinessDetails{
		Name:            rawData.Name,
		LocationDetails: locationDetails,
	}

	lineType, err := utils.GetLineType(rawData.LineType)
	if err != nil {
		return source.NumberDetails{}, err
	}

	data := source.NumberDetails{
		Number:          phoneNumber,
		Carrier:         rawData.Carrier,
		LineType:        lineType,
		FraudScore:      rawData.FraudScore,
		BusinessDetails: businessDetails,
		RecentAbuse:     rawData.RecentAbuse,
	}

	return data, nil
}
