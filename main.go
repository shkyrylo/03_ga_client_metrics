package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	nbuAPIURL     = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json"
	measurementID = "G-H1CHMFR50J"
	apiSecret     = "L2CQFYUiT0eWrspiHP2lPw"
	clientID      = "505"
)

type ExchangeRate struct {
	CurrencyCode string  `json:"cc"`
	Rate         float64 `json:"rate"`
}

func getExchangeRate(currencyCode string) (float64, error) {
	resp, err := http.Get(nbuAPIURL)
	if err != nil {
		return 0, fmt.Errorf("не вдалося отримати курс: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("неочікуваний статус відповіді: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("не вдалося зчитати відповідь: %v", err)
	}

	var rates []ExchangeRate
	if err := json.Unmarshal(body, &rates); err != nil {
		return 0, fmt.Errorf("не вдалося розпарсити JSON: %v", err)
	}

	for _, rate := range rates {
		if rate.CurrencyCode == currencyCode {
			return rate.Rate, nil
		}
	}

	return 0, fmt.Errorf("курс для валюти %s не знайдено", currencyCode)
}

func sendEvent(exchangeRate float64) error {
	url := fmt.Sprintf("https://www.google-analytics.com/mp/collect?measurement_id=%s&api_secret=%s", measurementID, apiSecret)

	data := map[string]interface{}{
		"client_id": clientID,
		"events": []map[string]interface{}{
			{
				"name": "exchange_rate",
				"params": map[string]interface{}{
					"currency": "USD",
					"rate":     exchangeRate,
				},
			},
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("не вдалося створити JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("не вдалося надіслати запит: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API помилка: %s", resp.Status)
	}

	return nil
}

func main() {
	exchangeRate, err := getExchangeRate("USD")
	if err != nil {
		fmt.Printf("Сталася помилка: %v\n", err)
		return
	}

	fmt.Printf("Курс USD/UAH: %f\n", exchangeRate)

	if err := sendEvent(exchangeRate); err != nil {
		fmt.Printf("Сталася помилка: %v\n", err)
	} else {
		fmt.Println("Подію успішно відправлено!")
	}
}
