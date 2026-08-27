package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// ── A stand-in for the OSRS Wiki ─────────────────────────────────────
//
// The real api.php answers two shapes we care about: opensearch (title
// suggestions) and query+prop=pageimages (does this page exist, and does it
// have artwork). The fake answers both, and records every request so a test
// can assert how many calls a lookup costs.

type fakePage struct {
	id    int
	thumb string // empty means the page exists but carries no image
}

type fakeWiki struct {
	pages     map[string]fakePage // title -> page
	redirects map[string]string   // requested title -> real title
	searches  map[string][]string // opensearch query -> titles

	mu            sync.Mutex
	imageRequests [][]string // the titles asked for, per pageimages call
	thumbSizes    []string   // pithumbsize, per pageimages call
	searchQueries []string
}

func (f *fakeWiki) imageCalls() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]string, len(f.imageRequests))
	copy(out, f.imageRequests)
	return out
}

func (f *fakeWiki) searchCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.searchQueries...)
}

func (f *fakeWiki) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	w.Header().Set("Content-Type", "application/json")

	switch q.Get("action") {
	case "opensearch":
		term := q.Get("search")
		f.mu.Lock()
		f.searchQueries = append(f.searchQueries, term)
		titles := append([]string(nil), f.searches[term]...)
		f.mu.Unlock()

		urls := make([]string, len(titles))
		for i, t := range titles {
			urls[i] = "https://wiki.test/w/" + strings.ReplaceAll(t, " ", "_")
		}
		body, _ := json.Marshal([]any{term, titles, make([]string, len(titles)), urls})
		_, _ = w.Write(body)

	case "query":
		asked := strings.Split(q.Get("titles"), "|")
		f.mu.Lock()
		f.imageRequests = append(f.imageRequests, asked)
		f.thumbSizes = append(f.thumbSizes, q.Get("pithumbsize"))
		f.mu.Unlock()
		_, _ = w.Write(f.pageImagesBody(asked))

	default:
		http.Error(w, "unexpected action", http.StatusBadRequest)
	}
}

func (f *fakeWiki) pageImagesBody(asked []string) []byte {
	type thumb struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	type page struct {
		PageID    int     `json:"pageid,omitempty"`
		NS        int     `json:"ns"`
		Title     string  `json:"title"`
		Missing   *string `json:"missing,omitempty"`
		Thumbnail *thumb  `json:"thumbnail,omitempty"`
		PageImage string  `json:"pageimage,omitempty"`
	}
	type redirect struct {
		From string `json:"from"`
		To   string `json:"to"`
	}

	pages := map[string]page{}
	var redirects []redirect
	nextMissing := -1
	empty := ""

	for _, title := range asked {
		target := title
		if to, ok := f.redirects[title]; ok {
			target = to
			redirects = append(redirects, redirect{From: title, To: to})
		}
		p, ok := f.pages[target]
		if !ok {
			pages[strconv.Itoa(nextMissing)] = page{NS: 0, Title: title, Missing: &empty}
			nextMissing--
			continue
		}
		entry := page{PageID: p.id, NS: 0, Title: target}
		if p.thumb != "" {
			entry.Thumbnail = &thumb{Source: p.thumb, Width: 150, Height: 150}
			entry.PageImage = strings.ReplaceAll(target, " ", "_") + ".png"
		}
		pages[strconv.Itoa(p.id)] = entry
	}

	body := map[string]any{
		"batchcomplete": "",
		"query":         map[string]any{"pages": pages},
	}
	if len(redirects) > 0 {
		body["query"].(map[string]any)["redirects"] = redirects
	}
	out, _ := json.Marshal(body)
	return out
}

func newFakeClient(t *testing.T, f *fakeWiki) *Client {
	t.Helper()
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	c := NewClient()
	c.BaseURL = srv.URL + "/api.php"
	return c
}

// ── Hits ─────────────────────────────────────────────────────────────

func TestLookupImage_HitFollowsRedirectAndEchoesInput(t *testing.T) {
	f := &fakeWiki{
		pages:     map[string]fakePage{"Twisted bow": {id: 82098, thumb: "https://wiki.test/150px-Twisted_bow_detail.png"}},
		redirects: map[string]string{"Tbow": "Twisted bow"},
	}
	c := newFakeClient(t, f)

	got, err := c.LookupImage("Tbow", 150)
	if err != nil {
		t.Fatalf("LookupImage: %v", err)
	}
	if !got.Found {
		t.Fatalf("want found, got %+v", got)
	}
	if got.Input != "Tbow" {
		t.Errorf("input = %q, want the string as typed", got.Input)
	}
	if got.Result.Title != "Twisted bow" {
		t.Errorf("title = %q, want the resolved page title", got.Result.Title)
	}

	// A hit must not go shopping for alternatives.
	if n := len(f.searchCalls()); n != 0 {
		t.Errorf("a hit ran %d searches, want 0", n)
	}

	var wire map[string]any
	raw, _ := json.Marshal(got)
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if wire["found"] != true {
		t.Errorf("found = %v, want true", wire["found"])
	}
	for _, k := range []string{"input", "title", "image_url", "width", "height", "page_image"} {
		if _, ok := wire[k]; !ok {
			t.Errorf("hit JSON is missing %q: %s", k, raw)
		}
	}
}

// ── Misses ───────────────────────────────────────────────────────────

func TestLookupImage_MissingPageIsNotFoundWithStrippedCandidates(t *testing.T) {
	f := &fakeWiki{
		pages: map[string]fakePage{
			"Barrows":           {id: 10064, thumb: "https://wiki.test/150px-Barrows_minigame.png"},
			"Barrows equipment": {id: 10065, thumb: "https://wiki.test/150px-Barrows_equipment.png"},
		},
		// As typed, opensearch finds nothing. That is the whole problem.
		searches: map[string][]string{
			"Barrows piece": {},
			"Barrows":       {"Barrows", "Barrows equipment"},
		},
	}
	c := newFakeClient(t, f)

	got, err := c.LookupImage("Barrows piece", 150)
	if err != nil {
		t.Fatalf("LookupImage: %v", err)
	}
	if got.Found {
		t.Fatal("a page with no wiki entry must not be a hit")
	}
	if got.Reason != ReasonNotFound {
		t.Errorf("reason = %q, want %q", got.Reason, ReasonNotFound)
	}
	if len(got.Candidates) != 2 {
		t.Fatalf("candidates = %+v, want the two Barrows pages", got.Candidates)
	}
	if got.Candidates[0].Title != "Barrows" {
		t.Errorf("candidates[0] = %q, want the best match first", got.Candidates[0].Title)
	}
	if got.Candidates[0].ImageURL == "" {
		t.Error("a candidate must carry the image URL the caller will render")
	}

	// The stripped search is the one that saved it.
	if q := f.searchCalls(); len(q) < 2 || q[1] != "Barrows" {
		t.Errorf("searches = %v, want the category words stripped on the second try", q)
	}
}

func TestLookupImage_ExistingPageWithNoArtworkIsNoImage(t *testing.T) {
	f := &fakeWiki{
		pages: map[string]fakePage{
			"Champion's scroll": {id: 28395}, // exists, no thumbnail
			"Champion's cape":   {id: 28396, thumb: "https://wiki.test/150px-Champions_cape.png"},
		},
		searches: map[string][]string{
			"Champion's scroll": {"Champion's scroll", "Champion's cape"},
		},
	}
	c := newFakeClient(t, f)

	got, err := c.LookupImage("Champion's scroll", 150)
	if err != nil {
		t.Fatalf("LookupImage: %v", err)
	}
	if got.Found {
		t.Fatal("a page with no pageimage is not a hit")
	}
	if got.Reason != ReasonNoImage {
		t.Errorf("reason = %q, want %q (the page does exist)", got.Reason, ReasonNoImage)
	}
	// The input's own page must not be offered back as an alternative.
	for _, cand := range got.Candidates {
		if cand.Title == "Champion's scroll" {
			t.Error("the input's own page was offered as a candidate")
		}
	}
	if len(got.Candidates) != 1 || got.Candidates[0].Title != "Champion's cape" {
		t.Errorf("candidates = %+v, want just Champion's cape", got.Candidates)
	}
}

func TestLookupImage_CandidatesWithoutArtworkAreDroppedEntirely(t *testing.T) {
	// Every suggestion is a disambiguation-style page with no image, so the
	// honest answer is an empty list, not a list of blank tiles.
	f := &fakeWiki{
		pages: map[string]fakePage{
			"Zulrah (disambiguation)": {id: 1},
			"Zulrah lore":             {id: 2},
		},
		searches: map[string][]string{
			"Any Zulrah unique": {},
			"Zulrah":            {"Zulrah (disambiguation)", "Zulrah lore"},
		},
	}
	c := newFakeClient(t, f)

	got, err := c.LookupImage("Any Zulrah unique", 150)
	if err != nil {
		t.Fatalf("LookupImage: %v", err)
	}
	if got.Found {
		t.Fatal("want a miss")
	}
	if len(got.Candidates) != 0 {
		t.Fatalf("candidates = %+v, want none: not one of them has an image", got.Candidates)
	}

	// An empty list must still serialise as [], never as null.
	raw, _ := json.Marshal(got)
	if !strings.Contains(string(raw), `"candidates":[]`) {
		t.Errorf("miss JSON = %s, want an empty candidates array", raw)
	}
}

func TestLookupImage_CandidatesAreCappedAtFive(t *testing.T) {
	pages := map[string]fakePage{}
	var found []string
	for _, name := range []string{"Zulrah", "Zulrah's scales", "Tanzanite fang", "Magic fang", "Serpentine visage", "Uncut onyx", "Toxic blowpipe"} {
		pages[name] = fakePage{id: len(pages) + 100, thumb: "https://wiki.test/150px-" + strings.ReplaceAll(name, " ", "_") + ".png"}
		found = append(found, name)
	}
	f := &fakeWiki{
		pages:    pages,
		searches: map[string][]string{"Zulrah drop": {}, "Zulrah": found},
	}
	c := newFakeClient(t, f)

	got, err := c.LookupImage("Zulrah drop", 150)
	if err != nil {
		t.Fatalf("LookupImage: %v", err)
	}
	if len(got.Candidates) != 5 {
		t.Fatalf("candidates = %d, want the list capped at 5", len(got.Candidates))
	}
	if got.Candidates[0].Title != "Zulrah" {
		t.Errorf("candidates[0] = %q, want best first", got.Candidates[0].Title)
	}
}

func TestLookupImage_VerifiesEveryCandidateInOneBatchedCall(t *testing.T) {
	names := []string{"Zulrah", "Zulrah's scales", "Tanzanite fang", "Magic fang", "Serpentine visage"}
	pages := map[string]fakePage{}
	for i, n := range names {
		pages[n] = fakePage{id: 200 + i, thumb: "https://wiki.test/150px-x.png"}
	}
	f := &fakeWiki{
		pages:    pages,
		searches: map[string][]string{"Any Zulrah unique": {}, "Zulrah": names},
	}
	c := newFakeClient(t, f)

	if _, err := c.LookupImage("Any Zulrah unique", 200); err != nil {
		t.Fatalf("LookupImage: %v", err)
	}

	calls := f.imageCalls()
	// One call resolves the input, one verifies the whole candidate list.
	if len(calls) != 2 {
		t.Fatalf("made %d pageimages calls, want 2 (one lookup, one batched verify): %v", len(calls), calls)
	}
	if len(calls[1]) != len(names) {
		t.Errorf("the verify call asked for %d titles, want all %d in one request", len(calls[1]), len(names))
	}

	// --size must reach the candidates too, or a board mixes thumbnail sizes.
	for i, s := range f.thumbSizes {
		if s != "200" {
			t.Errorf("pageimages call %d used pithumbsize=%q, want 200", i, s)
		}
	}
}

func TestLookupImage_StaysUnderTheCallBudget(t *testing.T) {
	f := &fakeWiki{pages: map[string]fakePage{}, searches: map[string][]string{}}
	c := newFakeClient(t, f)

	if _, err := c.LookupImage("Any 3x Ancestral robe piece", 150); err != nil {
		t.Fatalf("LookupImage: %v", err)
	}
	total := len(f.imageCalls()) + len(f.searchCalls())
	if total > 6 {
		t.Errorf("a total miss cost %d HTTP calls, want at most 6", total)
	}
}

// ── Query generation ─────────────────────────────────────────────────

func TestCandidateQueries(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"Any Zulrah unique", []string{"Any Zulrah unique", "Zulrah"}},
		{"Barrows piece", []string{"Barrows piece", "Barrows"}},
		{"3x Ancestral robe pieces", []string{"3x Ancestral robe pieces", "Ancestral robe", "robe", "Ancestral"}},
		{"Twisted bow", []string{"Twisted bow", "bow", "Twisted"}},
	}
	for _, tc := range cases {
		got := candidateQueries(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("candidateQueries(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("candidateQueries(%q) = %v, want %v", tc.in, got, tc.want)
				break
			}
		}
	}
}
