package main

import (
	"backend/db"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

var db_con *sql.DB

func main() {

	db_con = db.Establish_db_conn()

	// Specify the path to Chrome/Chromium executable
	const CHROME_EXEC_PATH = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	var currentUser, _ = user.Current()
	var username = currentUser.Username
	var USER_DATA_DIR = fmt.Sprintf("/Users/%s/Library/Application Support/Google/Chrome/", username)
	const PROFILE_DIRECTORY = "Default"
	// options
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:0], // No default options to provent chrome account login problems.
		chromedp.ExecPath(CHROME_EXEC_PATH),
		chromedp.DisableGPU,
		chromedp.UserDataDir(USER_DATA_DIR),
		chromedp.Flag("profile-directory", PROFILE_DIRECTORY),
		chromedp.Flag("headless", false),
		chromedp.Flag("flag-switches-begin", true),
		chromedp.Flag("flag-switches-end", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("new-window", true),
	)
	// Create a custom Chrome allocator with the specified path
	allocatorCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create a browser context
	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	const urlBaseSrting = "https://www.finanzfluss.de/informer/etf/suche?page=%d&per=100"

	log.Println("Starting browser...")

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

	for currPage <= maxPage {
		var url = fmt.Sprintf(urlBaseSrting, currPage)
		log.Println("Scraping url ", url)

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			closePopup(),
			awaitTableLoad(),
			scrapeData(),
		)
		if err != nil {
			log.Printf("Failed to execute chromedp tasks: %v", err)
		}
		currPage++
	}
}

func getMaxPage(maxPage *int) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.Evaluate(`Math.ceil(parseInt(document.querySelector(".result-number").innerText.replaceAll('.', '')) /100)`, &maxPage).Do(ctx)
			if err != nil {
				return fmt.Errorf("error evaluating maxPage: %w", err)
			}
			log.Println("Max page:", maxPage)
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

			timeout := 30 * time.Second

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

func closePopup() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to close the popup if it appears
			var popupExists bool
			err := chromedp.EvaluateAsDevTools(`!!document.querySelector('#CybotCookiebotDialogBodyButtonDecline')`, &popupExists).Do(ctx)
			if err != nil || !popupExists {
				fmt.Println("Popup didn't appear.")
				return nil
			}
			fmt.Println("Popup appeared! Closing...")
			return chromedp.Click(`#CybotCookiebotDialogBodyButtonDecline`, chromedp.ByID).Do(ctx)
		}),
	}
}

func scrapeData() chromedp.Tasks {
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
				db.InsertOrUpdateEtf(db_con, result["id"], result["name"], result["fundVolume"], isDistributing, releaseDate, result["replicationMethod"], result["shareClassVolume"], float32(totalExpenseRatioFloat))
				insertedCount++
			}

			log.Println("Inserted/Updated Count:", insertedCount)

			return nil
		}),
	}
}
