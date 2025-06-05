package main

import (
	"PhoneNumberCheck/config"
	"PhoneNumberCheck/providers/jpnumber"
	webscraping "PhoneNumberCheck/webScraping"
	"encoding/json"
	"fmt"
)

func main() {
	config.LoadEnv()
	// numberChan := make(chan string)
	// stopChan := make(chan struct{})
	// go webcamdetection.StartOCRScanner(numberChan, stopChan)
	// ipqsSource, err := ipqualityscore.Initialize()
	// if err != nil {
	// 	panic(err)
	// }
	// sources := []source.Source{ipqsSource}
	// ipqsSource.GetData("07091762683")
	// go webscraping.Main()
	driver, err := webscraping.InitializeDriver()
	if err != nil {
		panic(err)
	}
	defer driver.Close()
	//TODO: Send error here
	jpNumberProvider := jpnumber.Initialize(driver)

	// var datas []source.PhoneData
	testNums := []string{
		// "08003007299",
		// "08007770319",
		// "0366360855",
		// "05031075729",
		// "05811521308",
		"0752317111",
		// "0356641888",
		// "05054822807",
		// "0661675628",
		// "08005009120",
		// "09077097477",
		// "08087569409",
		// "09034998875",
		// "08020178530",
	}
	for _, val := range testNums {
		data, err := jpNumberProvider.GetData(val)
		if err != nil {
			panic(err)
		}
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonData))
		fmt.Printf("\n\n\n\n\n/////////////////////////////////////////////////////////////////")
	}
	// for _, val := range datas {
	// 	// fmt.Printf("Type: %s", val.LineType)
	// 	// fmt.Printf("Number: %s", val.Number)
	// }

}
