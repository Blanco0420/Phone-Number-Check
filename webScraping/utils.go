package webscraping

import (
	"fmt"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

type WebDriverWrapper struct {
	driver  selenium.WebDriver
	service *selenium.Service
}

func InitializeDriver() (*WebDriverWrapper, error) {
	const port = 4444
	service, err := selenium.NewGeckoDriverService("geckodriver", port)
	if err != nil {
		return &WebDriverWrapper{}, fmt.Errorf("Error starting geckodriver service: %v", err)
	}

	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{
		Args: []string{
			// "--headless",
		},
	}
	caps.AddFirefox(firefoxCaps)

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return &WebDriverWrapper{}, fmt.Errorf("Error connecting to remote server: %v", err)
	}

	return &WebDriverWrapper{
		driver:  driver,
		service: service,
	}, nil
}

func (w *WebDriverWrapper) GotoUrl(url string) error {
	return w.driver.Get(url)
}

func (w *WebDriverWrapper) EnterText(selector, text string) error {
	elem, err := w.driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return err
	}
	return elem.SendKeys(text)
}

func (w *WebDriverWrapper) FindElement(selector string) (selenium.WebElement, error) {
	elem, err := w.driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return nil, err
	}
	return elem, nil
}

func (w *WebDriverWrapper) FindElements(selector string) ([]selenium.WebElement, error) {
	elems, err := w.driver.FindElements(selenium.ByCSSSelector, selector)
	if err != nil {
		return nil, err
	}
	return elems, nil
}

func (w *WebDriverWrapper) CheckElementExists(selector string) bool {
	_, err := w.driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return false
	}
	return true
}

func (w *WebDriverWrapper) GetInnerText(selector string) (string, error) {
	elem, err := w.driver.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return "", fmt.Errorf("Error finding element with selector %s: %v", selector, err)
	}
	text, err := w.GetInnerTextFromElement(elem)
	if err != nil {
		return "", err
	}
	return text, nil
}

func (w *WebDriverWrapper) GetInnerTextFromElement(elem selenium.WebElement) (string, error) {
	text, err := elem.Text()
	if err != nil {
		return "", fmt.Errorf("Error getting text on element %v: %v", elem, err)
	}
	return text, nil
}

func (w *WebDriverWrapper) Close() {
	if w.driver != nil {
		w.driver.Quit()
	}
	if w.service != nil {
		w.service.Stop()
	}
}

// func Main() {
// 	const (
// 		serverUrl = "http://localhost:4444"
// 	)
//
// 	caps := selenium.Capabilities{"browserName": "firefox"}
//
// 	wd, err := selenium.NewRemote(caps, serverUrl)
// 	if err != nil {
// 		log.Fatal("Error starting remote: ", err)
// 	}
//
// 	defer wd.Quit()
//
// 	err = wd.Get("https://google.com")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	title, err := wd.Title()
// 	if err != nil {
// 		log.Fatal(err)
//
// 	}
//
// 	fmt.Println(title)
// }
