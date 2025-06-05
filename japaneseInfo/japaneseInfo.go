package japaneseinfo

import (
	"PhoneNumberCheck/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/extrame/xls"
)

type JapaneseInfo struct {
	validPrefectures map[string]struct{}
	validCities      map[string]struct{}
}

// TODO: Make better error handling and better error messages for missing file
// TEST:REMOVE EXPORT!!!!!
func GetWorkbookFile() (*xls.WorkBook, error) {
	const fileName = "japaneseRegionData.xls"

	if exists := utils.CheckIfFileExists(fileName); exists {
		file, err := xls.Open(fileName, "utf-8")
		if err != nil {
			// Try to fix by renaming or whatever your function does
			if err := checkDataFileExistsAndRename(); err != nil {
				return nil, fmt.Errorf("failed to rename data file: %w", err)
			}

			// Retry opening after rename
			file, err = xls.Open(fileName, "utf-8")
			if err != nil {
				return nil, fmt.Errorf("failed to open data file after rename: %w", err)
			}
		}
		return file, nil
	} else {
		// If file doesn't exist initially, try rename check immediately
		if err := checkDataFileExistsAndRename(); err != nil {
			return nil, fmt.Errorf("no data file found: %w", err)
		}

		// After rename, try to open the expected file
		file, err := xls.Open(fileName, "utf-8")
		if err != nil {
			return nil, fmt.Errorf("failed to open data file after rename: %w", err)
		}
		return file, nil
	}
}

func checkDataFileExistsAndRename() error {
	dir := "./"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if strings.EqualFold(ext, ".xls") {
			newName := "japaneseRegionData.xls"
			oldPath := filepath.Join(dir, name)
			newPath := filepath.Join(dir, newName)

			err := os.Rename(oldPath, newPath)
			if err != nil {
				return fmt.Errorf("failed to rename file %s to %s: %w", oldPath, newPath, err)
			}
			return nil
		}
	}
	return fmt.Errorf("no .xls data file found in directory")
}

//TODO: Find a way to reduce DRY HERE \/\/\/

type rawPrefectureData struct {
	iso    string
	en     string
	ja     string
	jaKana string `json:"ja-kana"`
}

func (info *JapaneseInfo) CheckIfPrefectureExists(rawPrefecture string) (string, error) {
	fmt.Printf("RAW: %q", rawPrefecture)
	for pref := range info.validPrefectures {
		fmt.Printf("PREFECTURE: %q", pref)
	}
	for pref := range info.validPrefectures {
		if strings.Contains(rawPrefecture, pref) {
			return pref, nil
		}
	}
	return "", fmt.Errorf("prefecture not found")
}

func (info *JapaneseInfo) CheckIfCityExists(rawCity string) (string, error) {
	for city := range info.validCities {
		if strings.Contains(rawCity, city) {
			return city, nil
		}
	}
	return "", fmt.Errorf("city not found")
}

//FIXME: Broken xls reader. does not work properly or I'm just being dumb

// TODO: Actually return errors rather than panic
func InitializeJapaneseInfo() (*JapaneseInfo, error) {
	ji := &JapaneseInfo{
		validPrefectures: make(map[string]struct{}),
		validCities:      make(map[string]struct{}),
	}
	// httpRequest, err := http.Get("https://raw.githubusercontent.com/pd-navi/japan-prefectures/refs/heads/master/data/prefectures.json")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	//
	// 	var rawJsonData []rawPrefectureData
	//
	// 	body, err := io.ReadAll(httpRequest.Body)
	// 	defer httpRequest.Body.Close()
	//
	// 	err = json.Unmarshal(body, &rawJsonData)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	file, err := GetWorkbookFile()
	if err != nil {
		panic(err)
	}
	sheet := file.GetSheet(2)
	if sheet == nil {
		panic("error getting sheet")
	}
	fmt.Printf("MAX: %d", sheet.MaxRow)
	for i := 1; i <= int(sheet.MaxRow); i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}
		fmt.Printf("Row %d: %q\n", i, row.Col(0))
		fmt.Printf("Row %d: %q\n", i, row.Col(1))
		fmt.Printf("Row %d: %q\n", i, row.Col(2))
		fmt.Printf("Row %d: %q\n", i, row.Col(3)) // current pref col

		pref := row.Col(4)
		if pref != "" {
			ji.validPrefectures[pref] = struct{}{}
		}

		city := row.Col(6)
		if city != "" {
			ji.validCities[city] = struct{}{}
		}
	}

	return ji, nil
}
