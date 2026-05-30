package tests

import (
	"context"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const baseURL = "http://localhost:3000"

func newBrowser(t *testing.T) *rod.Browser {
	t.Helper()
	l := launcher.
		New().
		Headless(true).
		Set("no-sandbox", "").
		Set("disable-dev-shm-usage", "")
	url := l.MustLaunch()
	browser := rod.New().
		ControlURL(url).
		Timeout(30 * time.Second).
		MustConnect()
	t.Cleanup(func() {
		browser.MustClose()
		l.Cleanup()
	})
	return browser
}

func loginAs(t *testing.T, browser *rod.Browser, email, password string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page := browser.Context(ctx).MustPage(baseURL + "/login")
	page.MustWaitLoad()
	page.MustElement("input[name='email']").MustInput(email)
	page.MustElement("input[name='password']").MustInput(password)
	page.MustElement("button[type='submit']").MustClick()
	page.MustWaitNavigation()
}

func TestSessionCreateRendersCorrectly(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	browser := newBrowser(t)
	loginAs(t, browser, "mosquito_42@backpack.dev", "password123")

	page := browser.Context(ctx).MustPage(baseURL + "/sessions/new")
	page.MustWaitLoad()

	page.MustElement("select[name='job_id']")
	page.MustElement("input[name='start_time']")
	page.MustElement("textarea[name='notes']")
	page.MustElement("button[type='submit']")
}

func TestSessionCreateValidatesEmptyJobID(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	browser := newBrowser(t)
	loginAs(t, browser, "mosquito_42@backpack.dev", "password123")

	page := browser.Context(ctx).MustPage(baseURL + "/sessions/new")
	page.MustWaitLoad()
	page.MustElement("button[type='submit']").MustClick()
	time.Sleep(500 * time.Millisecond)

	if page.MustInfo().URL != baseURL+"/sessions/new" {
		t.Fatal("expected to stay on session create page after empty submission")
	}

	page.MustElement("[data-error='job_id']")
}

func TestSessionCreateValidatesEmptyStartTime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	browser := newBrowser(t)
	loginAs(t, browser, "mosquito_42@backpack.dev", "password123")

	page := browser.Context(ctx).MustPage(baseURL + "/sessions/new")
	page.MustWaitLoad()
	page.MustElement("button[type='submit']").MustClick()
	time.Sleep(500 * time.Millisecond)

	page.MustElement("[data-error='start_time']")
}
