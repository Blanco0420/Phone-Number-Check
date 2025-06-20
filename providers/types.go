package providers

import (
	"net/http"
	"time"

	"github.com/tebeka/selenium"
)

type TableEntry struct {
	Key     string
	Value   string
	Element selenium.WebElement
}

type GraphData struct {
	Date     time.Time
	Accesses int
}

type Comment struct {
	Text     string
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
	Website         string
	LocationDetails LocationDetails
	NameSuffixes    []string
}

type VitalInfo struct {
	Name              string
	Industry          string
	CompanyOverview   string
	LineType          LineType
	OverallFraudScore int
	FraudulentDetails FraudulentDetails
}

type NumberDetails struct {
	Number          string
	Carrier         string
	VitalInfo       VitalInfo
	BusinessDetails BusinessDetails
	SiteInfo        SiteInfo
}

type FraudulentDetails struct {
	FraudScore  int
	RecentAbuse bool
}

type LineType string

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
