package ipqualityscore

import (
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	"PhoneNumberCheck/providers"
	"PhoneNumberCheck/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type IpqsSource struct {
	config           *providers.APIConfig
	vitalInfoChannel chan providers.VitalInfo
}

func (i *IpqsSource) VitalInfoChannel() <-chan providers.VitalInfo {
	return i.vitalInfoChannel
}

func (i *IpqsSource) CloseVitalInfoChannel() {
	close(i.vitalInfoChannel)
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
	config := providers.NewApiConfig(apiKey, baseUrl)
	return &IpqsSource{config: config, vitalInfoChannel: make(chan providers.VitalInfo)}, nil
}
func (i *IpqsSource) GetData(phoneNumber string) (providers.NumberDetails, error) {
	requestUrl := strings.Replace(i.config.BaseUrl, "<NUMBER>", phoneNumber, 1)
	res, err := i.config.HttpClient.Get(requestUrl)
	if err != nil {
		return providers.NumberDetails{}, fmt.Errorf("Error: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return providers.NumberDetails{}, err
	}

	var rawData rawApiData
	if err := json.Unmarshal(body, &rawData); err != nil {
		return providers.NumberDetails{}, err
	}

	if !rawData.Success {
		return providers.NumberDetails{}, fmt.Errorf("Error getting data from source:\n%v", rawData.Errors)
	}

	switch v := rawData.IdentityData.(type) {
	case string:
		fmt.Println("identityData is string", v)
		panic("identityData")
	case []any:
		fmt.Println("identityData is array", v)
		panic("identityData is array")
	}

	lineType, err := utils.GetLineType(rawData.LineType)
	if err != nil {
		return providers.NumberDetails{}, err
	}

	data := providers.NumberDetails{
		Number:  phoneNumber,
		Carrier: rawData.Carrier,
		VitalInfo: providers.VitalInfo{
			LineType: lineType,
			Name:     rawData.Name,
			FraudulentDetails: providers.FraudulentDetails{
				FraudScore:  rawData.FraudScore,
				RecentAbuse: rawData.RecentAbuse,
			},
		},
	}

	if err := japaneseinfo.GetAddressInfo(fmt.Sprintf("%s%s", rawData.Region, rawData.City), &data.BusinessDetails.LocationDetails); err != nil {
		return data, err
	}

	return data, nil
}
