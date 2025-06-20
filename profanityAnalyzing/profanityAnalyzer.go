package profanityAnalyzing

import (
	"PhoneNumberCheck/providers"
	"fmt"
	"strings"
	"sync"

	"github.com/agnivade/levenshtein"
)

type commentResult struct {
	index int
	text  string
	score int
}

func scoreComment(comment string) int {
	score := 0
	for word := range badWords {
		if strings.Contains(comment, word) {
			score += 2
		} else if levenshtein.ComputeDistance(comment, word) <= LevenshteinThreshold {
			score += 1
		}
	}
	return score
}

// TODO: Maybe remove concurrency as paramater and make an env var
func ScoreComments(comments []providers.Comment, concurrency int) int {
	fmt.Println("Scoring comments...")
	in := make(chan int, len(comments))
	out := make(chan int, len(comments))

	var wg sync.WaitGroup
	wg.Add(concurrency)
	//TODO: Add the comment score to final data for ml use
	for range concurrency {
		go func() {
			defer wg.Done()
			for comment := range in {
				score := scoreComment(comments[comment].Text)
				out <- score
			}
		}()
	}

	go func() {
		for i := range comments {
			in <- i
		}
		close(in)
	}()

	go func() {
		wg.Wait()
		close(out)
	}()

	total := 0
	for s := range out {
		total += s
	}
	return total
}
