package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Why a lookup returned nothing.
const (
	// ReasonNotFound means no wiki page answers to that title at all.
	ReasonNotFound = "not_found"
	// ReasonNoImage means the page is real but carries no artwork.
	ReasonNoImage = "no_image"
)

const (
	maxCandidates = 5  // how many alternatives a caller gets
	searchLimit   = 5  // titles requested per opensearch call
	collectCap    = 12 // stop collecting titles once we have this many
	enoughTitles  = 8  // stop running further searches once we have this many
)

// ImageCandidate is an alternative page that does have artwork.
type ImageCandidate struct {
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
}

// ImageLookup is the whole answer to `osrs-wiki image`: either the artwork
// for the page the input resolved to, or the reason there is none plus a
// short list of pages the caller could use instead.
//
// It marshals to one of two shapes. A hit keeps every field the command has
// always emitted and adds `input` and `found`. A miss carries `reason` and
// `candidates` and nothing else, so a consumer can branch on `found` alone.
type ImageLookup struct {
	Input  string
	Found  bool
	Result *ImageResult

	// Set only on a miss.
	Reason        string
	Candidates    []ImageCandidate
	ResolvedTitle string // the page that exists but has no image
}

func (l ImageLookup) MarshalJSON() ([]byte, error) {
	if l.Found && l.Result != nil {
		return json.Marshal(struct {
			Input         string `json:"input"`
			Found         bool   `json:"found"`
			Title         string `json:"title"`
			ImageURL      string `json:"image_url"`
			FullURL       string `json:"full_url,omitempty"`
			Width         int    `json:"width"`
			Height        int    `json:"height"`
			PageImageName string `json:"page_image,omitempty"`
		}{
			Input: l.Input, Found: true,
			Title: l.Result.Title, ImageURL: l.Result.ImageURL, FullURL: l.Result.FullURL,
			Width: l.Result.Width, Height: l.Result.Height, PageImageName: l.Result.PageImageName,
		})
	}

	// An empty list is an answer; it must never serialise as null.
	candidates := l.Candidates
	if candidates == nil {
		candidates = []ImageCandidate{}
	}
	return json.Marshal(struct {
		Input      string           `json:"input"`
		Found      bool             `json:"found"`
		Reason     string           `json:"reason"`
		Candidates []ImageCandidate `json:"candidates"`
	}{Input: l.Input, Found: false, Reason: l.Reason, Candidates: candidates})
}

// Message is the one-line explanation for a human reading stderr.
func (l *ImageLookup) Message() string {
	var b strings.Builder
	if l.Reason == ReasonNoImage {
		title := l.ResolvedTitle
		if title == "" {
			title = l.Input
		}
		fmt.Fprintf(&b, "the OSRS Wiki page '%s' exists but has no image.", title)
	} else {
		fmt.Fprintf(&b, "no OSRS Wiki page for '%s'.", l.Input)
	}

	if len(l.Candidates) == 0 {
		fmt.Fprintf(&b, " Try 'osrs-wiki search %s'", l.Input)
		return b.String()
	}
	titles := make([]string, len(l.Candidates))
	for i, c := range l.Candidates {
		titles[i] = c.Title
	}
	fmt.Fprintf(&b, " Try one of: %s", strings.Join(titles, ", "))
	return b.String()
}

// LookupImage resolves an item name to wiki artwork.
//
// A transport failure is an error. "The wiki has nothing for this" is not: it
// comes back as a miss carrying the reason and a few alternatives, because a
// caller building a bingo board can act on that and cannot act on an error
// string.
func (c *Client) LookupImage(input string, size int) (*ImageLookup, error) {
	if size <= 0 {
		size = 150
	}

	// Both spellings go out in one request. The wiki title casing rule is
	// "Twisted bow", not "Twisted Bow", and asking for both costs nothing
	// extra when they travel in the same batch.
	spellings := []string{input}
	if wc := wikiCapitalize(input); wc != input {
		spellings = append(spellings, wc)
	}

	pages, err := c.pageImages(spellings, size)
	if err != nil {
		return nil, err
	}

	resolved := ""
	for _, s := range spellings {
		p := pages[s]
		if p == nil || p.Missing {
			continue
		}
		if p.ThumbURL != "" {
			return &ImageLookup{Input: input, Found: true, Result: p.result()}, nil
		}
		if resolved == "" {
			resolved = p.Title
		}
	}

	reason := ReasonNotFound
	if resolved != "" {
		reason = ReasonNoImage
	}
	return &ImageLookup{
		Input:         input,
		Found:         false,
		Reason:        reason,
		ResolvedTitle: resolved,
		Candidates:    c.findCandidates(input, resolved, size),
	}, nil
}

// ── Candidates ───────────────────────────────────────────────────────

// Words that describe a category of thing rather than name one. A bingo tile
// reads "Any Zulrah unique"; the wiki has a page called "Zulrah".
var categoryWords = map[string]bool{
	"any": true, "a": true, "an": true,
	"piece": true, "pieces": true,
	"unique": true, "uniques": true,
	"drop": true, "drops": true,
	"item": true, "items": true,
	"x": true,
}

// A leading count: "3x", "5".
var leadingCount = regexp.MustCompile(`^\d+x?$`)

// stripCategoryWords turns "Any 3x Ancestral robe piece" into "Ancestral robe".
func stripCategoryWords(input string) string {
	kept := make([]string, 0, 4)
	for _, field := range strings.Fields(input) {
		word := strings.ToLower(strings.Trim(field, `.,;:"()[]`))
		if categoryWords[word] {
			continue
		}
		if len(kept) == 0 && leadingCount.MatchString(word) {
			continue
		}
		kept = append(kept, field)
	}
	return strings.Join(kept, " ")
}

// candidateQueries lists the searches to try, best bet first: the phrase as
// typed, then with the category words gone, then the individual words of what
// is left.
func candidateQueries(input string) []string {
	queries := []string{input}

	stripped := stripCategoryWords(input)
	if stripped != "" && !strings.EqualFold(stripped, input) {
		queries = append(queries, stripped)
	}

	base := stripped
	if base == "" {
		base = input
	}
	if words := strings.Fields(base); len(words) > 1 {
		queries = append(queries, words[len(words)-1], words[0])
	}

	seen := map[string]bool{}
	out := make([]string, 0, len(queries))
	for _, q := range queries {
		k := normKey(q)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, q)
	}
	return out
}

// findCandidates collects title suggestions, then verifies in a single
// batched request which of them actually have artwork. A candidate with no
// image is worse than no candidate: it renders as a blank tile.
func (c *Client) findCandidates(input, resolved string, size int) []ImageCandidate {
	skip := map[string]bool{normKey(input): true}
	if resolved != "" {
		skip[normKey(resolved)] = true
	}

	var ordered []string
	for _, query := range candidateQueries(input) {
		hits, err := c.Search(query, searchLimit)
		if err != nil {
			continue // a failed suggestion is not a failed lookup
		}
		for _, h := range hits {
			k := normKey(h.Title)
			if skip[k] || !usableCandidate(h.Title) {
				continue
			}
			skip[k] = true
			ordered = append(ordered, h.Title)
		}
		if len(ordered) >= enoughTitles {
			break
		}
	}

	if len(ordered) == 0 {
		return []ImageCandidate{}
	}
	if len(ordered) > collectCap {
		ordered = ordered[:collectCap]
	}

	pages, err := c.pageImages(ordered, size)
	if err != nil {
		return []ImageCandidate{}
	}

	out := make([]ImageCandidate, 0, maxCandidates)
	emitted := map[string]bool{}
	for _, title := range ordered {
		p := pages[title]
		if !p.hasImage() || emitted[normKey(p.Title)] {
			continue
		}
		emitted[normKey(p.Title)] = true
		out = append(out, ImageCandidate{Title: p.Title, ImageURL: p.ThumbURL})
		if len(out) == maxCandidates {
			break
		}
	}
	return out
}

// usableCandidate rejects other namespaces ("File:...") and subpages
// ("Zulrah/Strategies"), neither of which is ever the artwork for a tile.
func usableCandidate(title string) bool {
	return !strings.Contains(title, ":") && !strings.Contains(title, "/")
}

func normKey(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// ── The pageimages call ──────────────────────────────────────────────

type wikiPage struct {
	Title       string
	Missing     bool
	ThumbURL    string
	ThumbWidth  int
	ThumbHeight int
	PageImage   string
}

func (p *wikiPage) hasImage() bool {
	return p != nil && !p.Missing && p.ThumbURL != ""
}

func (p *wikiPage) result() *ImageResult {
	return &ImageResult{
		Title:         p.Title,
		ImageURL:      p.ThumbURL,
		FullURL:       fullSizeURL(p.ThumbURL),
		Width:         p.ThumbWidth,
		Height:        p.ThumbHeight,
		PageImageName: p.PageImage,
	}
}

// fullSizeURL strips the /thumb/ segment and the size prefix off a thumbnail
// URL to get the original upload.
func fullSizeURL(thumbURL string) string {
	if !strings.Contains(thumbURL, "/thumb/") {
		return thumbURL
	}
	withoutThumb := strings.Replace(thumbURL, "/thumb/", "/", 1)
	if lastSlash := strings.LastIndex(withoutThumb, "/"); lastSlash != -1 {
		return withoutThumb[:lastSlash]
	}
	return withoutThumb
}

// pageImages asks for the thumbnail of every given title in one request and
// returns the pages keyed by the title that was asked for, having followed
// MediaWiki's normalization and redirects. So a caller can ask for "Tbow" and
// read back the "Twisted bow" page under that key.
func (c *Client) pageImages(titles []string, size int) (map[string]*wikiPage, error) {
	if len(titles) == 0 {
		return map[string]*wikiPage{}, nil
	}

	apiURL := fmt.Sprintf(
		"%s?action=query&titles=%s&prop=pageimages&format=json&pithumbsize=%d&redirects=1",
		c.apiBase(), url.QueryEscape(strings.Join(titles, "|")), size,
	)
	data, err := c.get(apiURL)
	if err != nil {
		return nil, err
	}

	type titleMap struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	var resp struct {
		Query struct {
			Normalized []titleMap `json:"normalized"`
			Redirects  []titleMap `json:"redirects"`
			Pages      map[string]struct {
				PageID    int             `json:"pageid"`
				Title     string          `json:"title"`
				Missing   json.RawMessage `json:"missing"`
				PageImage string          `json:"pageimage"`
				Thumbnail struct {
					Source string `json:"source"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"thumbnail"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}

	byTitle := make(map[string]*wikiPage, len(resp.Query.Pages))
	for _, p := range resp.Query.Pages {
		byTitle[p.Title] = &wikiPage{
			Title:       p.Title,
			Missing:     pageIsMissing(p.Missing),
			ThumbURL:    p.Thumbnail.Source,
			ThumbWidth:  p.Thumbnail.Width,
			ThumbHeight: p.Thumbnail.Height,
			PageImage:   p.PageImage,
		}
	}

	hops := make(map[string]string, len(resp.Query.Normalized)+len(resp.Query.Redirects))
	for _, n := range resp.Query.Normalized {
		hops[n.From] = n.To
	}
	redirects := make(map[string]string, len(resp.Query.Redirects))
	for _, r := range resp.Query.Redirects {
		redirects[r.From] = r.To
	}

	out := make(map[string]*wikiPage, len(titles))
	for _, asked := range titles {
		key := asked
		if to, ok := hops[key]; ok {
			key = to
		}
		if to, ok := redirects[key]; ok {
			key = to
		}
		if p, ok := byTitle[key]; ok {
			out[asked] = p
		}
	}
	return out, nil
}

// pageIsMissing reads MediaWiki's flag for a title that has no page. The API
// sends `"missing": ""` (and a negative page id); formatversion=2 sends
// `true`. Without this check a nonexistent page is indistinguishable from a
// real page that simply has no artwork, which is the bug this replaces.
func pageIsMissing(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s != "" && s != "null" && s != "false"
}
