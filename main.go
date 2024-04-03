package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type GoldData struct {
	Gold        float64 `json:"gold"`
	GoldTenGram float64 `json:"goldTenGram"`
	SellGold    float64 `json:"sellGold"`
}

func main() {
	buyGoldURL := buildURL("com_midas_todays_price_MidasTodaysPricePortlet_INSTANCE_kpwp", "2", "normal", "view", "/serveBuyLivePrice")
	sellGoldURL := buildURL("com_midas_todays_price_sell_MidasTodaysPriceSellApiPortlet_INSTANCE_kpwp", "2", "normal", "view", "/serveSellLivePrice")

	var wg sync.WaitGroup
	wg.Add(2)

	buyGoldChan := make(chan GoldData)
	sellGoldChan := make(chan GoldData)

	go fetchData(buyGoldURL, buyGoldChan, &wg)
	go fetchData(sellGoldURL, sellGoldChan, &wg)

	go func() {
		wg.Wait()
		close(buyGoldChan)
		close(sellGoldChan)
	}()

	buyGoldData := <-buyGoldChan
	sellGoldData := <-sellGoldChan

	printHeader()

	formattedRates := formatGoldRates(buyGoldData, sellGoldData)
	fmt.Println(formattedRates)
}

func buildURL(p_p_id, p_p_lifecycle, p_p_state, p_p_mode, p_p_resource_id string) string {
	u := url.URL{
		Scheme: "https",
		Host:   "www.mmtcpamp.com",
		Path:   "/gold-silver-rate-today",
	}
	q := u.Query()
	q.Set("p_p_id", p_p_id)
	q.Set("p_p_lifecycle", p_p_lifecycle)
	q.Set("p_p_state", p_p_state)
	q.Set("p_p_mode", p_p_mode)
	q.Set("p_p_resource_id", p_p_resource_id)
	u.RawQuery = q.Encode()
	return u.String()
}

func fetchData(url string, ch chan<- GoldData, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}
	defer resp.Body.Close()
	var data GoldData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	ch <- data
}

func printHeader() {
	fmt.Println("MMTC-PAMP Gold Rates in INR")
	fmt.Println("Last update:", time.Now().Format("02/Jan/2006 15:04:05 PM"))
	fmt.Println()
}

func formatGoldRates(buyGoldData GoldData, sellGoldData GoldData) string {
	buyPriceGram := fmt.Sprintf("₹%.2f", buyGoldData.Gold)
	buyPriceTola := fmt.Sprintf("₹%.2f", buyGoldData.Gold*11.6638038)
	sellPriceGram := fmt.Sprintf("₹%.2f", sellGoldData.SellGold)
	sellPriceTola := fmt.Sprintf("₹%.2f", sellGoldData.SellGold*11.6638038)

	gramUnit := "/gm"
	tolaUnit := "/tola"

	buyingRate := buyPriceGram + gramUnit + "\t" + buyPriceTola + tolaUnit + "\n"
	sellingRate := sellPriceGram + gramUnit + "\t" + sellPriceTola + tolaUnit + "\n"

	return "Buying Rate:\t" + buyingRate + "Selling Rate:\t" + sellingRate
}
