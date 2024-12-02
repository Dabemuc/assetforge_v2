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

func ScrapeList() {

	const urlBaseSrting = "https://www.finanzfluss.de/informer/etf/suche?page=%d&per=100"

	log.Println("Starting list scraper ...")

	ctx, cancel := getChromdpCtx()
	defer cancel() // Make sure to clean up when done.

	var maxPage int
	err := chromedp.Run(ctx,
		chromedp.Navigate(fmt.Sprintf(urlBaseSrting, 1)),
		getMaxPage(&maxPage),
	)
	if err != nil {
		log.Printf("Failed to execute chromedp tasks: %v", err)
	}

	var currPage = 1
	log.Println("Maxpage:", maxPage)
	var renderedPageNr int

	for currPage <= maxPage {
		var url = fmt.Sprintf(urlBaseSrting, currPage)
		log.Println("##### Scraping url ", url)

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			closePopup(),
			awaitTableLoad(),
			getRenderedPage(&renderedPageNr),
		)
		if err != nil {
			log.Printf("Failed to execute chromedp tasks: %v", err)
		}
		if renderedPageNr == currPage {
			err := chromedp.Run(ctx,
				scrapeList(),
			)
			if err != nil {
				log.Printf("Failed to execute chromedp tasks: %v", err)
			}
			log.Println("Page", currPage, "scraped successfully!")
			currPage++
		} else {
			log.Println("renderedPageNr", renderedPageNr, "does not equal targeted pagenr", currPage, "- Redoing this page")
		}
	}
}

func getRenderedPage(renderedPageNr *int) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`parseInt(document.querySelector("button.pagination-number.current-number").innerText)`, &renderedPageNr).Do(ctx)
			if err != nil {
				return fmt.Errorf("Error evaluating renderedPageNr: %w", err)
			}
			return nil
		}),
	}

}

func getMaxPage(maxPage *int) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`Math.ceil(parseInt(document.querySelector(".result-number").innerText.replaceAll('.', '')) /100)`, &maxPage).Do(ctx)
			if err != nil {
				return fmt.Errorf("error evaluating maxPage: %w", err)
			}
			return nil
		}),
	}
}

func awaitTableLoad() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("Waiting for table to load 100 rows")
			var numRows int
			start := time.Now()

			timeout := 10 * time.Second

			for {
				// Evaluate the number of rows in the table
				err := chromedp.Evaluate(`document.querySelectorAll(".results-table tbody tr").length`, &numRows).Do(ctx)
				if err != nil {
					return fmt.Errorf("error evaluating row count: %w", err)
				}

				// Check if we have 100 rows (or if it's the last page with fewer)
				if numRows >= 100 {
					log.Println("Table completed loading 100 rows")
					break
				}

				// Timeout check (if it takes too long, stop waiting)
				if time.Since(start) > timeout {
					log.Println("Timeout reached, table may not have 100 rows")
					break
				}

				// Optionally, you can wait a little before trying again
				time.Sleep(200 * time.Millisecond)
			}

			return nil
		}),
	}
}

func scrapeList() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			var results []map[string]string
			err := chromedp.Evaluate(`Array.from(document.querySelectorAll('#main > div.page-content > div > div.informer-search > div.main-content > div.results-table > div > div.table-container > table > tbody > tr')).map(el => {
        url = el.querySelector('.name a')?.href;
        if (url.endsWith('/')) {
            url = url.substring(0, url.length-1)
        };
        id = url.split('/').pop();
        return {
          id: id,
					name: el.querySelector('.name')?.innerText,
					totalExpenseRatio: el.querySelector('.totalExpenseRatio')?.innerText,
					isDistributing: (!!el.querySelector('svg path[d^="M21.8371"]')).toString() || "false",
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
