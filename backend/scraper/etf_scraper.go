package scraper

import (
	"backend/db"
	"context"
	"fmt"
	"log"

	"encoding/json"

	"github.com/chromedp/chromedp"
)

func ScrapeEtf(id string) {

	const urlBaseSrting = "https://www.finanzfluss.de/informer/etf/%s"

	log.Println("Starting etf scraper ...")

	ctx, cancel := getChromdpCtx()
	defer cancel() // Make sure to clean up when done.

	var url = fmt.Sprintf(urlBaseSrting, id)
	log.Println("Scraping url ", url)

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		closePopup(),
		scrapeEtf(id),
	)
	if err != nil {
		log.Printf("Failed to execute chromedp tasks: %v", err)
	}
}

func scrapeEtf(id string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			var results db.EtfDetailsData
			results.Id = id
			log.Println("Scraping..........")
			err := chromedp.Evaluate(`(function() {
        // Expand activity_distribution
        document.querySelector("#main > div.page-content > div > div.mx-auto.my-0.max-w-\\[960px\\].space-y-\\[40px\\].md\\:space-y-\\[64px\\] > div:nth-child(4) > div:nth-child(2) > div > div > div.show-more-less").click()
				
        // Gather data
        return {
          //scrape_date_details: TODO: Put in db method,
          isin: document.querySelector("#Copy-ISIN-Matomo .value").innerText,
          wkn: document.querySelector("#Copy-WKN-Matomo .value").innerText,
          nr_positions: document.querySelector(".etf-data-wrapper > :nth-child(5) .value").innerText,
          base_index: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(1) .right-side").innerText,
          share_class_volume: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(6) .right-side").innerText,
          fund_domicile: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(8) .right-side").innerText,
          fund_currency: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(9) .right-side").innerText,
          securities_lending_permitted: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(10) .right-side").innerText == "Ja",
          trade_currency: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(11) .right-side").innerText,
          has_currency_hedging: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(12) .right-side").innerText == "Ja",
          has_special_assets: document.querySelector("div.etf-profile-page-basic-informations > div.show-more-cards-wrapper > div.show-more-card.base-data > div > div.content > div:nth-child(6) .right-side").innerText == "Ja",
          fund_provider: document.querySelector("div.show-more-card.legal-structure > div > div.content > div:nth-child(1) .right-side").innerText,
          legal_structure: document.querySelector("div.show-more-card.legal-structure > div > div.content > div:nth-child(2) .right-side").innerText,
          fund_structure: document.querySelector("div.show-more-card.legal-structure > div > div.content > div:nth-child(4) .right-side").innerText,
          administrator: document.querySelector("div.show-more-card.legal-structure > div > div.content > div:nth-child(5) .right-side").innerText,
          depotbank: document.querySelector("div.show-more-card.legal-structure > div > div.content > div:nth-child(6) .right-side").innerText,
          auditor: document.querySelector("div.show-more-card.legal-structure > div > div.content > div:nth-child(7) .right-side").innerText,
          country_composition: Array.from(document.querySelectorAll(".countries-card .content .country-item-wrapper")).map(wrapper => {
            return {
                country: wrapper.querySelector(".country-name").innerText,
                percentile: wrapper.querySelector(".progress-bar").firstChild.data,
            }
          }),
          region_composition: Array.from(document.querySelectorAll(".regions-card .region")).map(wrapper => {
            return {
                country: wrapper.querySelector(".name").innerText,
                percentile: wrapper.querySelector(".progress-bar").firstChild.data,
            }
          }),
          currency_distribution: Array.from(document.querySelectorAll(".currency-card .currency-item")).map(wrapper => {
            return {
                country: wrapper.querySelector(".currency-wrapper :nth-child(2)").innerText,
                percentile: wrapper.querySelector(".percentage").firstChild.data,
            }
          }),
          weight_top_10: document.querySelector(".diversification-card .row:nth-child(1) .percentage").innerText,
          nr_stock_positions: document.querySelector(".diversification-card .row:nth-child(3) .right-label").innerText,
          nr_bond_positions: document.querySelector(".diversification-card .row:nth-child(4) .right-label").innerText,
          nr_cash_and_other_positions: document.querySelector(".diversification-card .row:nth-child(5) .right-label").innerText,
          top_10_holdings: Array.from(document.querySelectorAll(".top-10-holdings-card-row")).map(wrapper => {
            return {
                name: wrapper.querySelector(".icon-name").innerText,
                percentile: wrapper.querySelector(".percentage").innerText
            }
          }),
          industry_distribution: Array.from(document.querySelectorAll(".sector-card .sector-card-row")).map(wrapper => {
            return {
                name: wrapper.querySelector(".label").innerText,
                percentile: wrapper.querySelector(".percentage").innerText
            }
          }),
          activity_distribution: Array.from(document.querySelectorAll("#main > div.page-content > div > div.mx-auto.my-0.max-w-\\[960px\\].space-y-\\[40px\\].md\\:space-y-\\[64px\\] > div:nth-child(4) > div:nth-child(2) > div > div > div.grid")).map(wrapper => {
            return {
                name: wrapper.querySelector(".label").innerText,
                percentiles: {
                    min: wrapper.querySelector("input").min,
                    value: wrapper.querySelector("input").value,
                    max: wrapper.querySelector("input").max   
                }
            }
          }),
          historical_performance: Array.from(document.querySelectorAll(".performance-card-table tr:has(td)")).map(wrapper => {
            return {
                timespan: wrapper.cells[0].firstChild.data,
                performance: wrapper.cells[1].innerText,
                return: wrapper.cells[2].innerText
            }
          }),
          historical_volatility: Array.from(document.querySelectorAll(".risk-metrics-item")[0].querySelectorAll(".risk-value-box")).map(wrapper => {
            return {
                period: wrapper.querySelector(".period").innerText,
                value: wrapper.querySelector(".value").innerText
            }
          }),
          historical_max_drawdown: Array.from(document.querySelectorAll(".risk-metrics-item")[1].querySelectorAll(".risk-value-box")).map(wrapper => {
            return {
                period: wrapper.querySelector(".period").innerText,
                value: wrapper.querySelector(".value").innerText
            }
          }),
          historical_sharpe_ratio: Array.from(document.querySelectorAll(".risk-metrics-item")[2].querySelectorAll(".risk-value-box")).map(wrapper => {
            return {
                period: wrapper.querySelector(".period").innerText,
                value: wrapper.querySelector(".value").innerText
            }
          }),
          exchanges: Array.from(document.querySelectorAll(".available-exchanges-data-table tr:has(td)")).map(wrapper => {
            return {
                name: wrapper.querySelector(".exchange-name").innerText,
                currency: wrapper.querySelector(".currency").innerText,
                ticker: wrapper.querySelector(".ticker").innerText
            }
          })
        };
      })();`, &results).Do(ctx)
			if err != nil {
				return err
			}
			log.Println("Raw result:")
			output, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(output))

			//parse and insert into db
			db.UpdateEtfDetails(results)
			return nil
		}),
	}
}
