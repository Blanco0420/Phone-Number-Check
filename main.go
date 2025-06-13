package main

import (
	"PhoneNumberCheck/config"
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	"PhoneNumberCheck/providers/ipqualityscore"
	"PhoneNumberCheck/providers/jpnumber"
	"PhoneNumberCheck/providers/numverify"
	"PhoneNumberCheck/providers/telnavi"
	"PhoneNumberCheck/source"
	"encoding/json"
	"fmt"
	"os"
)

var (
	testNums = []string{
		// "08003007299",
		// "08007770319",
		// "0366360855",
		// "05031075729",
		// "0648642471",
		"07091762683",
		// "05031595686",
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

func testingProvider(data map[string][]byte, source source.Source) map[string][]byte {
	for _, number := range testNums {
		sourceData, err := source.GetData(number)
		if err != nil {
			panic(err)
		}
		jsonData, err := json.MarshalIndent(sourceData, "", "  ")
		if err != nil {
			panic(err)
		}
		data[number] = jsonData
	}
	return data
}

func testingOutput(data map[string][]byte, sources []source.Source) map[string][]byte {

	for sourceIndex, localSource := range sources {
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
			data[sourceString] = jsonData
		}
	}
	return data
}

func main() {

	config.LoadEnv()
	japaneseinfo.Initialize()

	// TODO: Send error here
	// jpNumberProvider := jpnumber.Initialize(driver)

	numverify, err := numverify.Initialize()
	if err != nil {
		fmt.Println("here")
		panic(err)
	}

	jpNumber, err := jpnumber.Initialize()
	if err != nil {
		panic(err)
	}
	defer jpNumber.Close()

	ipqsSource, err := ipqualityscore.Initialize()
	if err != nil {
		panic(err)
	}

	telnavi, err := telnavi.Initialize()
	if err != nil {
		panic(err)
	}
	defer telnavi.Close()

	data := map[string][]byte{}
	sources := []source.Source{
		jpNumber,
		ipqsSource,
		numverify,
		telnavi,
	}
	//NOTE: Temporary used variables
	_ = sources

	// data = testingOutput(data, sources)
	data = testingProvider(data, jpNumber)
	for key, val := range data {
		if err := os.WriteFile(fmt.Sprintf("%s-output.json", key), val, 0644); err != nil {
			panic(err)
		}
	}

	// numberChan := make(chan string)
	// stopChan := make(chan struct{})
	// go webcamdetection.StartOCRScanner(numberChan, stopChan)

	// ipqsSource, err := ipqualityscore.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	japaneseinfo.Initialize()
	// pref, exists := japaneseinfo.FindPrefectureByCityName("台東区", 1)
	// fmt.Println(pref, exists)

}
