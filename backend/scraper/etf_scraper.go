package scraper

import (
	"backend/db"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func ScrapeEtf() {

	const urlBaseSrting = "https://www.finanzfluss.de/informer/etf/%s"

	log.Println("Starting etf scraper ...")

	var ctx = getChromdpCtx()

	var id = "ie00b5bmr087"

	var url = fmt.Sprintf(urlBaseSrting, id)
	log.Println("Scraping url ", url)

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		closePopup(),
		scrapeEtf(),
	)
	if err != nil {
		log.Printf("Failed to execute chromedp tasks: %v", err)
	}
}

func scrapeEtf() chromedp.Tasks { // TODO:Implement. This is copied
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			var pageNr string
			err := chromedp.Evaluate(`document.querySelector("button.pagination-number.current-number").innerText`, &pageNr).Do(ctx)
			if err != nil {
				return err
			}
			log.Println("Scraping page nr", pageNr)
			var results []map[string]string
			err = chromedp.Evaluate(`Array.from(document.querySelectorAll('#main > div.page-content > div > div.informer-search > div.main-content > div.results-table > div > div.table-container > table > tbody > tr')).map(el => {
        url = el.querySelector('.name a')?.href;
        if (url.endsWith('/')) {
            url = url.substring(0, url.length-1)
        };
        id = url.split('/').pop();
				return {
          id: id,
					name: el.querySelector('.name')?.innerText,
					totalExpenseRatio: el.querySelector('.totalExpenseRatio')?.innerText,
					isDistributing: el.querySelector('.isDistributing')?.innerText || "false",
					replicationMethod: el.querySelector('.replicationMethod')?.innerText,
					fundVolume: el.querySelector('.fundVolume')?.innerText,
					shareClassVolume: el.querySelector('.shareClassVolume')?.innerText,
					releaseDate: el.querySelector('.releaseDate')?.innerText,
				};
			})`, &results).Do(ctx)
			if err != nil {
				return err
			}

			var insertedCount = 0

			for _, result := range results {
				var isDistributing, err_isDistributing = strconv.ParseBool(result["isDistributing"])
				if err_isDistributing != nil {
					fmt.Println("Error parsing isDistributing:", err_isDistributing)
				}
				var releaseDate, err_releaseDate = time.Parse("02.01.06", result["releaseDate"]) // Layout for DD.MM.YY
				if err_releaseDate != nil {
					fmt.Println("Error parsing releaseDate:", err_releaseDate)
				}
				var totalExpenseRatio = strings.TrimSpace(result["totalExpenseRatio"])
				totalExpenseRatio = strings.TrimSuffix(totalExpenseRatio, "%")
				totalExpenseRatio = strings.ReplaceAll(totalExpenseRatio, "\u00a0", "")
				totalExpenseRatio = strings.ReplaceAll(totalExpenseRatio, ",", ".")
				if totalExpenseRatio == "—" {
					totalExpenseRatio = "0"
				}
				var totalExpenseRatioFloat, err_totalExpenseRatio = strconv.ParseFloat(totalExpenseRatio, 32)
				totalExpenseRatioFloat = totalExpenseRatioFloat / 100
				if err_totalExpenseRatio != nil {
					fmt.Println("Error parsing totalExpenseRatio:", err_totalExpenseRatio)
				}
				db.InsertOrUpdateEtf(result["id"], result["name"], result["fundVolume"], isDistributing, releaseDate, result["replicationMethod"], result["shareClassVolume"], float32(totalExpenseRatioFloat))
				insertedCount++
			}

			log.Println("Inserted/Updated Count:", insertedCount)

			return nil
		}),
	}
}
