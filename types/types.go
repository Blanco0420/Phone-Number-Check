package types

import (
	"time"

	"github.com/tebeka/selenium"
)

type TableEntry struct {
	Key     string
	Value   string
	Element selenium.WebElement
}

type GraphData struct {
	Date     time.Time
	Accesses int
}
