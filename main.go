package main

import (
	japaneseinfo "PhoneNumberCheck/japaneseInfo"
	"fmt"
)

func main() {
	// 	japaneseinfo, err := japaneseinfo.InitializeJapaneseInfo()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// pref, err := japaneseinfo.CheckIfPrefectureExists("東京都")
	// if err != nil {
	// 		panic(err)
	// 	}
	// fmt.Print("NEWPREF: ", pref)
	workbook, err := japaneseinfo.GetWorkbookFile()
	if err != nil {
		panic(err)
	}
	sheet := workbook.GetSheet(0)
	for i := 0; i <= int(sheet.MaxRow-1); i++ {
		row := sheet.Row(i)
		// col0 := row.Col(0)
		// col1 := row.Col(1)
		fmt.Printf("%s", row)
		// fmt.Printf("%v", row.Col(6))
	}

	// config.LoadEnv()
	//
	//	// numberChan := make(chan string)
	//	// stopChan := make(chan struct{})
	//	// go webcamdetection.StartOCRScanner(numberChan, stopChan)
	//	// ipqsSource, err := ipqualityscore.Initialize()
	//	// if err != nil {
	//	// 	panic(err)
	//	// }
	//	// sources := []source.Source{ipqsSource}
	//	// ipqsSource.GetData("07091762683")
	//	// go webscraping.Main()
	//	// driver, err := webscraping.InitializeDriver()
	//	// if err != nil {
	//	// 	panic(err)
	//	// }
	//	// defer driver.Close()
	//	// //TODO: Send error here
	//	// jpNumberProvider := jpnumber.Initialize(driver)
	//
	//	// var datas []source.PhoneData
	//	testNums := []string{
	//		// "08003007299",
	//		// "08007770319",
	//		// "0366360855",
	//		// "05031075729",
	//		// "05811521308",
	//		"0752317111",
	//		// "0356641888",
	//		// "05054822807",
	//		// "0661675628",
	//		// "08005009120",
	//		// "09077097477",
	//		// "08087569409",
	//		// "09034998875",
	//		// "08020178530",
	//	}
	//	ipQualitySource, err := ipqualityscore.Initialize()
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	for _, val := range testNums {
	//		data, err := ipQualitySource.GetData(val)
	//		if err != nil {
	//			panic(err)
	//		}
	//		jsonData, err := json.MarshalIndent(data, "", "  ")
	//		if err != nil {
	//			panic(err)
	//		}
	//		fmt.Println(string(jsonData))
	//		fmt.Printf("\n\n\n\n\n/////////////////////////////////////////////////////////////////")
	//	}
	//	// for _, val := range datas {
	//	// 	// fmt.Printf("Type: %s", val.LineType)
	//	// 	// fmt.Printf("Number: %s", val.Number)
	//	// }
}
