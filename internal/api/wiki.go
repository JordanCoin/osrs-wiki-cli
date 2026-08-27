package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the OSRS Wiki MediaWiki endpoint.
const DefaultBaseURL = "https://oldschool.runescape.wiki/api.php"

type Client struct {
	HTTPClient *http.Client
	// BaseURL is the MediaWiki api.php endpoint. Empty means the live wiki;
	// tests point it at an httptest server.
	BaseURL string
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		BaseURL:    DefaultBaseURL,
	}
}

func (c *Client) apiBase() string {
	if c.BaseURL == "" {
		return DefaultBaseURL
	}
	return c.BaseURL
}

// ImageResult holds the resolved image URL for an item.
type ImageResult struct {
	Title         string `json:"title"`
	ImageURL      string `json:"image_url"`
	FullURL       string `json:"full_url,omitempty"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	PageImageName string `json:"page_image,omitempty"`
}

// wikiCapitalize converts "Twisted Bow" → "Twisted bow" (only first letter uppercase).
func wikiCapitalize(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
}

// SearchResult holds an item from opensearch.
type SearchResult struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Search finds items by partial name using opensearch.
func (c *Client) Search(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	apiURL := fmt.Sprintf(
		"%s?action=opensearch&search=%s&limit=%d&format=json",
		c.apiBase(), url.QueryEscape(query), limit,
	)

	data, err := c.get(apiURL)
	if err != nil {
		return nil, err
	}

	// opensearch returns [query, [titles], [descriptions], [urls]]
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	if len(raw) < 4 {
		return nil, fmt.Errorf("unexpected response format")
	}

	var titles []string
	var urls []string
	json.Unmarshal(raw[1], &titles)
	json.Unmarshal(raw[3], &urls)

	var results []SearchResult
	for i := range titles {
		u := ""
		if i < len(urls) {
			u = urls[i]
		}
		results = append(results, SearchResult{Title: titles[i], URL: u})
	}
	return results, nil
}

// PriceResult holds GE price data.
type PriceResult struct {
	ItemID   int    `json:"item_id"`
	Name     string `json:"name"`
	High     int64  `json:"high"`
	Low      int64  `json:"low"`
	HighTime int64  `json:"high_time"`
	LowTime  int64  `json:"low_time"`
}

// GetPrice looks up the Grand Exchange price for an item by name.
// First resolves name → item ID via the mapping API, then gets price.
func (c *Client) GetPrice(itemName string) (*PriceResult, error) {
	// Load item mapping to find the ID
	mapping, err := c.getItemMapping()
	if err != nil {
		return nil, fmt.Errorf("failed to load item mapping: %w", err)
	}

	nameLower := strings.ToLower(itemName)
	var itemID int
	var foundName string
	for _, item := range mapping {
		if strings.ToLower(item.Name) == nameLower {
			itemID = item.ID
			foundName = item.Name
			break
		}
	}
	if itemID == 0 {
		return nil, fmt.Errorf("item '%s' not found in GE. Try 'osrs-wiki search %s'", itemName, itemName)
	}

	// Get price
	priceURL := fmt.Sprintf("https://prices.runescape.wiki/api/v1/osrs/latest?id=%d", itemID)
	data, err := c.getWithUserAgent(priceURL)
	if err != nil {
		return nil, err
	}

	var priceResp struct {
		Data map[string]struct {
			High     int64 `json:"high"`
			Low      int64 `json:"low"`
			HighTime int64 `json:"highTime"`
			LowTime  int64 `json:"lowTime"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &priceResp); err != nil {
		return nil, fmt.Errorf("invalid price response: %w", err)
	}

	idStr := fmt.Sprintf("%d", itemID)
	if p, ok := priceResp.Data[idStr]; ok {
		return &PriceResult{
			ItemID:   itemID,
			Name:     foundName,
			High:     p.High,
			Low:      p.Low,
			HighTime: p.HighTime,
			LowTime:  p.LowTime,
		}, nil
	}

	return nil, fmt.Errorf("no price data available for '%s'", foundName)
}

// ItemInfo holds basic item info from the mapping.
type ItemInfo struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Examine  string `json:"examine"`
	Members  bool   `json:"members"`
	HighAlch int    `json:"highalch"`
	LowAlch  int    `json:"lowalch"`
	Value    int    `json:"value"`
	Icon     string `json:"icon"`
}

// GetItem looks up item details from the mapping.
func (c *Client) GetItem(itemName string) (*ItemInfo, error) {
	mapping, err := c.getItemMapping()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(itemName)
	for _, item := range mapping {
		if strings.ToLower(item.Name) == nameLower {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("item '%s' not found in the Grand Exchange (only tradeable items). For bosses, quests, or game info, use: osrs-wiki wiki '%s'", itemName, itemName)
}

// Cached mapping
var cachedMapping []ItemInfo

func (c *Client) getItemMapping() ([]ItemInfo, error) {
	if cachedMapping != nil {
		return cachedMapping, nil
	}

	data, err := c.getWithUserAgent("https://prices.runescape.wiki/api/v1/osrs/mapping")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cachedMapping); err != nil {
		return nil, fmt.Errorf("invalid mapping response: %w", err)
	}
	return cachedMapping, nil
}

// PageContent holds wiki page text content.
type PageContent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// GetPage fetches the text content of a wiki page.
// Returns clean plaintext (no HTML). maxChars limits the response size.
func (c *Client) GetPage(title string, maxChars int) (*PageContent, error) {
	if maxChars <= 0 {
		maxChars = 4000
	}

	apiURL := fmt.Sprintf(
		"%s?action=query&titles=%s&prop=extracts&format=json&redirects=1&exchars=%d&explaintext=1",
		c.apiBase(), url.QueryEscape(title), maxChars,
	)

	data, err := c.get(apiURL)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Query struct {
			Pages map[string]struct {
				PageID  int    `json:"pageid"`
				Title   string `json:"title"`
				Extract string `json:"extract"`
				Missing bool   `json:"missing"`
			} `json:"pages"`
		} `json:"query"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	for _, page := range resp.Query.Pages {
		if page.Missing || page.Extract == "" {
			// Try wiki-capitalized version
			wikiTitle := wikiCapitalize(title)
			if wikiTitle != title {
				return c.getPageDirect(wikiTitle, maxChars)
			}
			return nil, fmt.Errorf("page '%s' not found on OSRS Wiki. Try 'osrs-wiki search %s'", title, title)
		}
		wikiURL := "https://oldschool.runescape.wiki/w/" + strings.ReplaceAll(page.Title, " ", "_")
		return &PageContent{
			Title:   page.Title,
			Content: page.Extract,
			URL:     wikiURL,
		}, nil
	}
	return nil, fmt.Errorf("page '%s' not found", title)
}

func (c *Client) getPageDirect(title string, maxChars int) (*PageContent, error) {
	apiURL := fmt.Sprintf(
		"%s?action=query&titles=%s&prop=extracts&format=json&redirects=1&exchars=%d&explaintext=1",
		c.apiBase(), url.QueryEscape(title), maxChars,
	)
	data, err := c.get(apiURL)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Query struct {
			Pages map[string]struct {
				Title   string `json:"title"`
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	for _, page := range resp.Query.Pages {
		if page.Extract == "" {
			return nil, fmt.Errorf("page '%s' not found", title)
		}
		wikiURL := "https://oldschool.runescape.wiki/w/" + strings.ReplaceAll(page.Title, " ", "_")
		return &PageContent{Title: page.Title, Content: page.Extract, URL: wikiURL}, nil
	}
	return nil, fmt.Errorf("page '%s' not found", title)
}

func (c *Client) get(url string) ([]byte, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func (c *Client) getWithUserAgent(reqURL string) ([]byte, error) {
	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("User-Agent", "osrs-wiki-cli (github.com/JordanCoin/osrs-wiki-cli)")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}
