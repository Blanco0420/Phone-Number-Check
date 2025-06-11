package numverify

import (
	"PhoneNumberCheck/source"
	"PhoneNumberCheck/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type NumverifySource struct {
	config *source.APIConfig
}

/*
{"valid":true,"number":"810752317111","local_format":"0752317111","international_format":"+81752317111","country_prefix":"+81","country_code":"JP","country_name":"Japan","location":"Kyoto","carrier":"","line_type":"landline"}
*/

type apiResponse struct {
	Valid   *bool     `json:"valid,omitempty"`
	Success *bool     `json:"success,omitempty"`
	Error   *apiError `json:"error,omitempty"`
}

type apiError struct {
	code int
	Type string
	info string
}

type successResponse struct {
	valid    bool
	location string
	carrier  string
	LineType string `json:"line_type"`
}

func Initialize() (*NumverifySource, error) {
	apiKey, exists := os.LookupEnv("PROVIDERS__NUMVERIFY_API_KEY")
	if !exists {
		return &NumverifySource{}, fmt.Errorf("Error, apiKey environment variable not set")
	}
	baseUrl := "https://apilayer.net/api/validate?access_key=" + apiKey + "&number=<NUMBER>&country_code=JP"
	config := source.NewApiConfig(apiKey, baseUrl)
	return &NumverifySource{config: config}, nil
}

func (s *NumverifySource) GetData(phoneNumber string) (source.NumberDetails, error) {
	var data source.NumberDetails
	requestUrl := strings.Replace(s.config.BaseUrl, "<NUMBER>", phoneNumber, 1)
	res, err := s.config.HttpClient.Get(requestUrl)
	if err != nil {
		return data, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return data, err
	}
	var apiResponse apiResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return data, err
	}

	if apiResponse.Success != nil && !*apiResponse.Success {
		return data, errors.New(apiResponse.Error.info)
	}

	var successResponse successResponse
	if err := json.Unmarshal(body, &successResponse); err != nil {
		return data, err
	}

	lineType, err := utils.GetLineType(successResponse.LineType)
	if err != nil {
		return data, err
	}
	data.Number = phoneNumber
	data.LineType = lineType
	data.Carrier = successResponse.carrier
	data.BusinessDetails.LocationDetails.City = successResponse.location

	return data, nil
}
