package profanityAnalyzing

import (
	"fmt"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

//TODO: Change to json or yaml (fastest)

var greenLevel = map[string]struct{}{
	"営業電話":       {},
	"新入社員":       {},
	"代表者":        {},
	"資産運用":       {},
	"挨拶":         {},
	"電話":         {},
	"証券会社":       {},
	"着信履歴":       {},
	"担当者":        {},
	"担当変更":       {},
	"お世話になっています": {},
	"昼休憩":        {},
	"電話があった":     {},
	"資産":         {},
	"ご挨拶":        {},
	"留守電":        {},
	"営業":         {}}

var yellowLevel = map[string]struct{}{
	"迷惑":        {},
	"しつこい":      {},
	"営業妨害":      {},
	"時間の無駄":     {},
	"断っても断っても":  {},
	"何回も電話":     {},
	"電話しつこい":    {},
	"勝手に電話":     {},
	"話したことない":   {},
	"営業が強引":     {},
	"再度電話":      {},
	"営業トーク":     {},
	"昼休みに電話":    {},
	"研修で電話":     {},
	"電話マナー悪い":   {},
	"営業の電話":     {},
	"馬鹿な会社":     {},
	"最低な証券会社":   {},
	"定期的に電話":    {},
	"2ヶ月おきに電話":  {},
	"コロナ落ち着いたら": {},
	"何度も電話":     {},
}

var redLevel = map[string]struct{}{
	"着信拒否しても":       {},
	"無差別電話営業":       {},
	"ストーカー的":        {},
	"失礼極まりない":       {},
	"断っても繰り返し":      {},
	"怒鳴られた":         {},
	"人の不幸を飯のタネにするな": {},
	"ボケた母親":         {},
	"痴呆症の母":         {},
	"母親に営業":         {},
	"気持ち悪い証券会社":     {},
	"ケシカラン守銭奴":      {},
	"信用できない":        {},
	"無能な証券会社":       {},
	"最低なやつ":         {},
	"職場教育":          {},
	"しつこすぎる":        {},
	"着信拒否無視":        {},
	"何十回も電話":        {},
	"強引な勧誘":         {},
	"顧客のこと考えてない":    {},
	"営業停止依頼無視":      {},
	"嘘をつく":          {},
	"ウソつき":          {},
	"セコイ貧乏証券会社":     {},
}

func tokenizeJapanese(text string) []string {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(fmt.Sprintf("Tokenizing error: %v", err))
	}
	tokens := t.Tokenize(text)

	var words []string
	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		words = append(words, token.Surface)
	}
	return words
}

func checkCommentForHits(comment string) (greenHits, yellowHits, redHits []string) {
	words := tokenizeJapanese(comment)

	for _, word := range words {
		if _, ok := greenLevel[word]; ok {
			greenHits = append(greenHits, word)
		}
		if _, ok := yellowLevel[word]; ok {
			yellowHits = append(yellowHits, word)
		}
		if _, ok := redLevel[word]; ok {
			redHits = append(redHits, word)
		}
	}
	return
}

// func ScoreComments(comments []types.Comment) int {
// 	score := 0
// 	for _, comment := range comments {
// 		greenHits, yellowHits, redHits := checkCommentForHits(comment.Text)
//
// 		score += len(greenHits) * 1
// 		score += len(yellowHits) * 3
// 		score += len(redHits) * 6
//
// 		for key := range redLevel {
// 			if strings.Contains(comment.Text, key) {
// 				score += 3
// 			}
// 		}
// 		for key := range yellowLevel {
// 			if strings.Contains(comment.Text, key) {
// 				score += 1
// 			}
// 		}
//
// 	}
// 	return int(math.Min(float64(score), 20))
// }
