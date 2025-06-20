package main

import (
	"PhoneNumberCheck/config"
	providerdataprocessing "PhoneNumberCheck/providerDataProcessing"
	"PhoneNumberCheck/providers"
	"PhoneNumberCheck/providers/jpnumber"
	"PhoneNumberCheck/providers/telnavi"
	"PhoneNumberCheck/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/trace"
	"sync"
)

var (
	testNums = []string{
		// "08003007299",
		// "08007770319",
		// "0366360855",
		// "05031075729",
		// "0648642471",
		// "07091762683",
		"05031595686",
		// "05811521308",
		// "0648642471",
		// "0368935962",
		// "0120830068",
		// "0752317111",
		// "0356641888",
		// "05054822807",
		// "0661675628",
		// "08005009120",
		// "09077097477",
		// "08087569409",
		// "09034998875",
		// "08020178530",
	}
)

func buildFinalDisplayData(data map[string]providers.NumberDetails) providerdataprocessing.FinalDisplayData {
	var businessNames, lineTypes, industries, businessOverviews []string
	var businessSources, lineTypeSources, industrySources, overviewSources []string
	var fraudScores []int
	var recentAbuseCount int
	var abuseSeen int

	for sourceName, details := range data {
		if details.VitalInfo.Name != "" {
			businessNames = append(businessNames, details.VitalInfo.Name)
			businessSources = append(businessSources, sourceName)
		}
		if details.VitalInfo.LineType != "" {
			lineTypes = append(lineTypes, string(details.VitalInfo.LineType))
			lineTypeSources = append(lineTypeSources, sourceName)
		}
		if details.VitalInfo.Industry != "" {
			industries = append(industries, details.VitalInfo.Industry)
			industrySources = append(industrySources, sourceName)
		}
		if details.VitalInfo.OverallFraudScore != 0 {
			fraudScores = append(fraudScores, details.VitalInfo.OverallFraudScore)
		}
		if details.VitalInfo.FraudulentDetails.RecentAbuse {
			//TODO: Maybe fix:
			// if not nil: abuseSeen++ ; if is true: recentAbuseCount++
			recentAbuseCount++
			abuseSeen++
		}
	}
	return providerdataprocessing.FinalDisplayData{
		BusinessName:    providerdataprocessing.CalculateFieldConfidence(businessNames, businessSources),
		LineType:        providerdataprocessing.CalculateFieldConfidence(lineTypes, lineTypeSources),
		Industry:        providerdataprocessing.CalculateFieldConfidence(industries, industrySources),
		CompanyOverview: providerdataprocessing.CalculateFieldConfidence(businessOverviews, overviewSources),
		FinalFraudScore: utils.AverageIntSlice(fraudScores),
		FinalRecentAbuse: func() bool {
			if abuseSeen == 0 {
				return false
			}
			return recentAbuseCount >= (abuseSeen / 2)
		}(),
	}
}

func testingProviders(data *map[string]providers.NumberDetails, sources map[string]providers.Source) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// UI goroutine to refresh screen
	// go func() {
	// 	for {
	// 		// time.Sleep(500 * time.Millisecond)
	// 		// fmt.Print("\033[H\033[2J") // Clear screen
	// 		// fmt.Println("Live Source Output:")
	// 		// fmt.Println("====================")
	// 		// outputsMu.Lock()
	// 		// for _, name := range orderedSourceNames {
	// 		// 	if out, ok := outputs[name]; ok {
	// 		// 		fmt.Printf("--- %s ---\n%s\n\n", name, out)
	// 		// 	} else {
	// 		// 		fmt.Printf("--- %s ---\n(waiting for data...)\n\n", name)
	// 		// 	}
	// 		// }
	// 		// outputsMu.Unlock()
	// 	}
	// }()
	for localSourceName, localSource := range sources {
		wg.Add(1)
		//TODO: actually do something with the channel instead of leaving it
		go func() {
			for _ = range localSource.VitalInfoChannel() {
			}
		}()
		go func(srcName string, src providers.Source) {
			defer wg.Done()
			for _, number := range testNums {
				fmt.Printf("[%s] calling getData\n", srcName)
				sourceData, err := src.GetData(number)
				if err != nil {
					panic(err)
				}
				mu.Lock()
				(*data)[srcName] = sourceData
				mu.Unlock()
			}
			src.CloseVitalInfoChannel()
			fmt.Printf("[%s] finished GetData and closed channel\n", srcName)
		}(localSourceName, localSource)
	}

	wg.Wait()
	return nil
}

func testingOutput(data *map[string][]byte, sources []providers.Source) {

	for sourceIndex, localSource := range sources {
		go func(src providers.Source) {
			for v := range src.VitalInfoChannel() {
				json, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					panic(err)
				}
				fmt.Printf("\n\n\n\n\n%s\n\n\n\n\n", string(json))
			}
		}(localSource)
		go func() {
			for _, number := range testNums {
				sourceData, err := localSource.GetData(number)
				if err != nil {
					panic(err)
				}
				sourceString := fmt.Sprintf("source%d", sourceIndex)
				jsonData, err := json.MarshalIndent(sourceData, "", "  ")
				if err != nil {
					panic(err)
				}
				(*data)[sourceString] = jsonData
			}
			localSource.CloseVitalInfoChannel()
		}()
	}
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	trace.Start(f)
	defer trace.Stop()
	config.LoadEnv()

	// TODO: Send error here
	// jpNumberProvider := jpnumber.Initialize(driver)

	// numverify, err := numverify.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	jpNumber, err := jpnumber.Initialize()
	if err != nil {
		panic(err)
	}
	defer jpNumber.Close()

	// ipqsSource, err := ipqualityscore.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	telnavi, err := telnavi.Initialize()
	if err != nil {
		panic(err)
	}
	defer telnavi.Close()

	data := map[string]providers.NumberDetails{}
	sources := map[string]providers.Source{
		"jpNumber": jpNumber,
		// ipqsSource,
		// numverify,
		"telnavi": telnavi,
	}
	// data = testingOutput(data, sources)
	go func() {
		log.Println("Starting pprof server at :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	err = testingProviders(&data, sources)
	if err != nil {
		panic(err)
	}
	finalData := buildFinalDisplayData(data)
	jsonBytes, err := json.MarshalIndent(finalData, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonBytes))
	// // for key, val := range data {
	// // 	json, err := json.MarshalIndent(val, "", "  ")
	// // 	if err != nil {
	// // 		panic(err)
	// // 	}
	// // }
	// localData, err := json.MarshalIndent(data, "", "  ")
	// if err != nil {
	// 	panic(err)
	// }
	// file, err := os.OpenFile("output.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	// defer file.Close()
	// _, err = file.WriteString("\n")
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = file.Write(localData)
	// if err != nil {
	// 	panic(err)
	// }

	// numberChan := make(chan string)
	// stopChan := make(chan struct{})
	// go webcamdetection.StartOCRScanner(numberChan, stopChan)

	// ipqsSource, err := ipqualityscore.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	// pref, exists := japaneseinfo.FindPrefectureByCityName("台東区", 1)
	// fmt.Println(pref, exists)

}
