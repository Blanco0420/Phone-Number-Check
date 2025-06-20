package providerdataprocessing

import (
	"PhoneNumberCheck/providers"
)

type NumberRiskInput struct {
	SourceName  string
	GraphData   []providers.GraphData
	FraudScore  *int
	RecentAbuse *bool
	Comments    []providers.Comment
}

type FinalDisplayData struct {
	BusinessName     []ConfidenceResult
	LineType         []ConfidenceResult
	Industry         []ConfidenceResult
	CompanyOverview  []ConfidenceResult
	FinalFraudScore  int
	FinalRecentAbuse bool
}
