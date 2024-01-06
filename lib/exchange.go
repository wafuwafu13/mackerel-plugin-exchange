package exchange

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ExchangePlugin mackerel plugin
type ExchangePlugin struct {
	Prefix string
}

// MetricKeyPrefix interface for PluginWithPrefix
func (u ExchangePlugin) MetricKeyPrefix() string {
	if u.Prefix == "" {
		u.Prefix = "exchange"
	}
	return u.Prefix
}

// GraphDefinition interface for mackerelplugin
func (u ExchangePlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := cases.Title(language.Und, cases.NoLower).String(u.Prefix)
	return map[string]mp.Graphs{
		"": {
			Label: labelPrefix,
			Unit:  mp.UnitFloat,
			Metrics: []mp.Metrics{
				{Name: "USD", Label: "USD"},
				{Name: "EUR", Label: "EUR"},
			},
		},
	}
}

type ExchangeAPIResponse struct {
	Success   bool               `json:"success"`
	Timestamp int                `json:"timestamp"`
	Base      string             `json:"base"`
	Date      string             `json:"date"`
	Rates     map[string]float64 `json:"rates"`
}

// FetchMetrics interface for mackerelplugin
func (u ExchangePlugin) FetchMetrics() (map[string]float64, error) {
	const path = "/Users/wafuwafu13/Desktop/mackerel-plugin-exchange/timestamp.txt"
	const interval = time.Hour * 2
	const errorValue = 200.0

	if !isOverTime(path, interval) {
		return map[string]float64{"USD": errorValue, "EUR": errorValue}, fmt.Errorf("Intentionally not sent")
	}

	writeTimestamp(path)

	err := godotenv.Load("/Users/wafuwafu13/Desktop/mackerel-plugin-exchange/.env")
	if err != nil {
		return map[string]float64{"USD": errorValue, "EUR": errorValue}, fmt.Errorf("godotenv.Load Error: %s", err)
	}

	apiKey := os.Getenv("EXCHANGE_API_KEY")

	url := fmt.Sprintf("http://api.exchangeratesapi.io/v1/latest?access_key=%s", apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return map[string]float64{"USD": errorValue, "EUR": errorValue}, fmt.Errorf("http.Get Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]float64{"USD": errorValue, "EUR": errorValue}, fmt.Errorf("io.ReadAll Error: %s", err)
	}

	var exchangeResponse ExchangeAPIResponse
	if err := json.Unmarshal(body, &exchangeResponse); err != nil {
		return map[string]float64{"USD": errorValue, "EUR": errorValue}, fmt.Errorf("json.Unmarshal Error: %s", err)
	}

	EURToJPY, ok1 := exchangeResponse.Rates["JPY"]
	EURToUSD, ok2 := exchangeResponse.Rates["USD"]
	if !ok1 || !ok2 {
		return map[string]float64{"USD": errorValue, "EUR": errorValue}, fmt.Errorf("ok Error")
	}
	USDToJPY := EURToJPY / EURToUSD

	return map[string]float64{"USD": USDToJPY, "EUR": EURToJPY}, nil
}

// isOverTime checks if the last timestamp in the file is more than interval
func isOverTime(path string, interval time.Duration) bool {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if lastLine == "" {
		return true
	}

	lastTimestamp, err := time.Parse("2006-01-02 15:04:05", lastLine)
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
		return false
	}

	return time.Now().UTC().Sub(lastTimestamp.UTC()) >= interval
}

// writeTimestamp writes the current timestamp to the file
func writeTimestamp(path string) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(time.Now().UTC().Format("2006-01-02 15:04:05") + "\n")
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

// Do the plugin
func Do() {
	u := ExchangePlugin{}
	helper := mp.NewMackerelPlugin(u)
	helper.Run()
}
