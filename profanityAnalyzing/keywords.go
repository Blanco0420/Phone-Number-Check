package profanityAnalyzing

import (
	"PhoneNumberCheck/config"
	"fmt"
	"strconv"
	"strings"

	"github.com/agnivade/levenshtein"
)

var badWords map[string]struct{}
var LevenshteinThreshold int

func init() {
	if levenshteinEnvValue, exists := config.GetEnvVariable("LEVENSHTEIN_THRESHOLD"); !exists {
		LevenshteinThreshold = 2
		return
	} else {
		parsed, err := strconv.Atoi(levenshteinEnvValue)
		if err != nil {
			LevenshteinThreshold = 2
			return
		}
		LevenshteinThreshold = parsed
	}

	fmt.Println("leven", LevenshteinThreshold)

}

func containsExactMatch(text string) bool {
	for word := range badWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
func ContainsFuzzyMatch(text string) bool {
	for word := range badWords {
		if levenshtein.ComputeDistance(text, word) <= LevenshteinThreshold {
			return true
		}
	}
	return false
}

func isBadByKeyword(text string) bool {
	text = strings.TrimSpace(text)
	return containsExactMatch(text) || ContainsFuzzyMatch(text)
}

//
// func classifyComment(text string) (bool, error) {
// 	if isBadByKeyword(text) {
// 		return true, nil
// 	}
// 	return false, nil
// 	//TODO: Implement this function
// 	// return isBadByMl(text)
// }
