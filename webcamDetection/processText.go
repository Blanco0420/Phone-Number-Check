package webcamdetection

import (
	"fmt"

	"github.com/otiai10/gosseract/v2"
)

func ProcessText(bytes []byte) {

	client := gosseract.NewClient()
	client.SetPageSegMode(gosseract.PSM_SINGLE_LINE)
	client.SetLanguage("eng")
	// client.SetWhitelist("0123456789")
	defer client.Close()
	client.SetImageFromBytes(bytes)
	text, err := client.Text()
	if err != nil {
		fmt.Println("error getting text: ", err)
	}
	fmt.Println(text)
	fmt.Println("####################################")
	fmt.Print("\n\n\n\n\n\n")

}
