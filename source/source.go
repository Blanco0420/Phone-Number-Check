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
	AccessCount int
	ReviewCount int
	UserRating  float32
	Comments    []Comment
}

type LocationDetails struct {
	Prefecture string
	City       string
	Address    string
	PostCode   string
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
	BusinessDetails BusinessDetails
	SiteInfo        SiteInfo
}

const (
	LineTypeMobile   LineType = "mobile"
	LineTypeTollFree LineType = "dialfree"
	LineTypeLandline LineType = "landline"
	LineTypeVOIP     LineType = "voip"
	LineTypeUnknown  LineType = "unknown"
	LineTypeOther    LineType = "other"
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
		Timeout:    10 * time.Second,
		HttpClient: &http.Client{Timeout: 10 * time.Second},
		Headers:    make(map[string]string),
	}
}

type Source interface {
	GetData(phoneNumber string) (NumberDetails, error)
}
