package source

import (
	"net/http"
	"time"
)

type LineType string

type Comment struct {
	Comment  string
	PostDate time.Time
}

type SiteInfo struct {
	ReviewCount int
	Comments    []Comment
}

type LocationDetails struct {
	Region   string
	City     string
	Address  string
	PostCode string
}

type BusinessDetails struct {
	Name            string
	Industry        string
	Website         string
	LocationDetails LocationDetails
	CompanyOverview string
}

type NumberDetails struct {
	Number          string
	Carrier         string
	LineType        LineType
	FraudScore      int
	RecentAbuse     bool
	SiteInfo        SiteInfo
	BusinessDetails BusinessDetails
}

const (
	LineTypeMobile   LineType = "mobile"
	LineTypeTollFree LineType = "dialfree"
	LineTypeLandline LineType = "landline"
	LineTypeVOIP     LineType = "ipphone"
	LineTypeUnknown  LineType = "unknown"
)

type APIConfig struct {
	APIKey     string
	BaseUrl    string
	Timeout    time.Duration
	HttpClient *http.Client
	Headers    map[string]string
}

func NewApiConfig(apiKey, baseUrl string) *APIConfig {
	return &APIConfig{
		APIKey:     apiKey,
		BaseUrl:    baseUrl,
		Timeout:    5 * time.Second,
		HttpClient: &http.Client{Timeout: 5 * time.Second},
		Headers:    make(map[string]string),
	}
}

type Source interface {
	GetData(phoneNumber string) (NumberDetails, error)
}
