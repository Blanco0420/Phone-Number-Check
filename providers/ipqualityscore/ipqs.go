package ipqualityscore

import (
	"PhoneNumberCheck/source"
	"fmt"
	"os"
)

type IpqsSource struct {
	config *source.APIConfig
}

// TODO: Raw data type here
type RawMessage struct {
}

func Initialize() (*IpqsSource, error) {
	apiKey, exists := os.LookupEnv("PROVIDERS__IPQS_API_KEY")
	if !exists {
		return &IpqsSource{}, fmt.Errorf("Error, apiKey environment variable not set")
	}
	baseUrl := "https://www.ipqualityscore.com/api/json/phone/" + apiKey
	config := source.NewApiConfig(apiKey, baseUrl)
	return &IpqsSource{config: config}, nil
}
func (i *IpqsSource) GetData(phoneNumber string) (*source.NumberInfo, error) {
	res, err := i.config.HttpClient.Get(i.config.BaseUrl + phoneNumber)
	if err != nil {
		return &source.NumberInfo{}, fmt.Errorf("Error: %v", err)
	}
	fmt.Println(res)
	return &source.NumberInfo{}, nil
}
