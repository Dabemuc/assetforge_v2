package scraper

import (
	"context"
	"fmt"
	"os/user"

	"github.com/chromedp/chromedp"
)

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

func getChromdpCtx() context.Context {
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

	return ctx
}
