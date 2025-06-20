package providerdataprocessing

import (
	"PhoneNumberCheck/profanityAnalyzing"
	"PhoneNumberCheck/providers"
	"fmt"
	"math"

	"github.com/agnivade/levenshtein"
)

func calculateGraphScore(data []providers.GraphData, recentAbuse *bool) int {
	score := 0

	n := len(data)
	if n < 3 {
		return score
	}

	todayAccesses := data[n-1].Accesses
	last3AvgAccesses := (todayAccesses + data[n-2].Accesses + data[n-3].Accesses) / 3

	if todayAccesses > 10 {
		score += 2
	}

	if last3AvgAccesses > 30 {
		score += 2
	}

	if last3AvgAccesses > 0 && todayAccesses > last3AvgAccesses*3 {
		score += 3
	}

	if last3AvgAccesses > 10 {
		*recentAbuse = true
	}

	nonZeroDayAccesses := 0
	for _, d := range data {
		if d.Accesses > 0 {
			nonZeroDayAccesses++
		}
	}
	if nonZeroDayAccesses > 7 {
		score += 1
	}
	return score
}

func EvaluateSource(input NumberRiskInput) int {
	fmt.Println("Evaluating source...")
	score := 0

	if len(input.GraphData) > 0 {
		graphScore := calculateGraphScore(input.GraphData, input.RecentAbuse)
		score += int(math.Min(float64(graphScore)/10.0*30.0, 30))
		fmt.Println("Score for graph data: ", graphScore)
	}

	if input.FraudScore != nil {
		//FIXME: TEMPORARY
		prevScore := score
		score += int(math.Min(float64(*input.FraudScore)/100.0*25.0, 25))
		fmt.Println("Score for fraudScore: ", score-prevScore)
	}

	if input.RecentAbuse != nil && *input.RecentAbuse {
		score += 15
		fmt.Println("Score for recentAbuse: ", score-15)
	}

	if len(input.Comments) > 0 {
		commentScore := profanityAnalyzing.ScoreComments(input.Comments, 4)
		score += int(math.Min(float64(commentScore)/20.0*30.0, 30))
		fmt.Println("Score for comments: ", commentScore)
	}
	if score > 100 {
		return 100
	}

	return score
}

type FieldComparison struct {
	Values      []string
	SourceNames []string
}

type ConfidenceResult struct {
	NormalizedValue string
	Confidence      float32
	Supporters      []string
}

// TODO: Instead of continuing when value has already been used, keep it in the loop and check if it get's a better score in another group
func CalculateFieldConfidence(values []string, sourceNames []string) []ConfidenceResult {
	threshold := 10
	// threshold := profanityAnalyzing.LevenshteinThreshold
	groups := make([]ConfidenceResult, 0)
	used := make([]bool, len(values))

	for i, value := range values {
		if used[i] {
			continue
		}
		used[i] = true
		group := ConfidenceResult{
			NormalizedValue: value,
			Supporters:      []string{sourceNames[i]},
		}
		for j := i + 1; j < len(values); j++ {
			if used[j] {
				continue
			}
			if levenshtein.ComputeDistance(value, values[j]) <= threshold {
				used[j] = true
				group.Supporters = append(group.Supporters, sourceNames[j])
			}
		}
		group.Confidence = float32(len(group.Supporters)) / float32(len(values))
		groups = append(groups, group)
	}
	return groups
}
