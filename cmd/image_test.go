package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JordanCoin/osrs-wiki-cli/internal/api"
)

// wikiStub answers just enough of api.php to drive the image command.
// "Twisted bow" has art, "Champion's scroll" exists without art, everything
// else is missing. Opensearch only knows "Barrows".
func wikiStub(t *testing.T) *api.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		if q.Get("action") == "opensearch" {
			if q.Get("search") == "Barrows" {
				_, _ = w.Write([]byte(`["Barrows",["Barrows"],[""],["https://wiki.test/w/Barrows"]]`))
				return
			}
			_, _ = w.Write([]byte(`["x",[],[],[]]`))
			return
		}
		var out []string
		missing := -1
		for _, title := range strings.Split(q.Get("titles"), "|") {
			switch title {
			case "Twisted bow":
				out = append(out, `"82098":{"pageid":82098,"ns":0,"title":"Twisted bow","thumbnail":{"source":"https://wiki.test/150px-Twisted_bow_detail.png","width":150,"height":154},"pageimage":"Twisted_bow_detail.png"}`)
			case "Barrows":
				out = append(out, `"10064":{"pageid":10064,"ns":0,"title":"Barrows","thumbnail":{"source":"https://wiki.test/150px-Barrows.png","width":150,"height":104},"pageimage":"Barrows.png"}`)
			case "Champion's scroll":
				out = append(out, `"28395":{"pageid":28395,"ns":0,"title":"Champion's scroll"}`)
			default:
				out = append(out, `"`+itoa(missing)+`":{"ns":0,"title":`+quote(title)+`,"missing":""}`)
				missing--
			}
		}
		_, _ = w.Write([]byte(`{"batchcomplete":"","query":{"pages":{` + strings.Join(out, ",") + `}}}`))
	}))
	t.Cleanup(srv.Close)
	c := api.NewClient()
	c.BaseURL = srv.URL + "/api.php"
	return c
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func TestRunImage_JSONHitExitsZeroAndCarriesInputAndFound(t *testing.T) {
	var out, errb bytes.Buffer
	code := runImage(wikiStub(t), "Twisted bow", 150, false, true, &out, &errb)
	if code != 0 {
		t.Fatalf("exit = %d, want 0. stderr: %s", code, errb.String())
	}

	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if got["found"] != true || got["input"] != "Twisted bow" || got["title"] != "Twisted bow" {
		t.Errorf("hit JSON = %s", out.String())
	}
}

func TestRunImage_JSONMissExitsZeroWithReasonAndCandidates(t *testing.T) {
	var out, errb bytes.Buffer
	code := runImage(wikiStub(t), "Barrows piece", 150, false, true, &out, &errb)
	if code != 0 {
		t.Fatalf("exit = %d, want 0: with --json the JSON is the answer", code)
	}

	var got struct {
		Input      string `json:"input"`
		Found      bool   `json:"found"`
		Reason     string `json:"reason"`
		Candidates []struct {
			Title    string `json:"title"`
			ImageURL string `json:"image_url"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if got.Found || got.Input != "Barrows piece" || got.Reason != "not_found" {
		t.Errorf("miss JSON = %s", out.String())
	}
	if len(got.Candidates) != 1 || got.Candidates[0].Title != "Barrows" || got.Candidates[0].ImageURL == "" {
		t.Errorf("candidates = %+v, want Barrows with an image", got.Candidates)
	}
	if errb.Len() != 0 {
		t.Errorf("stderr = %q, want nothing: the JSON already said it", errb.String())
	}
}

func TestRunImage_PlainMissExitsThreeWithTheRightSentence(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"missing page", "Barrows piece", "no OSRS Wiki page for 'Barrows piece'"},
		{"page without art", "Champion's scroll", "exists but has no image"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, errb bytes.Buffer
			code := runImage(wikiStub(t), tc.input, 150, false, false, &out, &errb)
			if code != 3 {
				t.Fatalf("exit = %d, want 3", code)
			}
			if !strings.Contains(errb.String(), tc.want) {
				t.Errorf("stderr = %q, want it to contain %q", errb.String(), tc.want)
			}
			if strings.Contains(errb.String(), "no thumbnail") {
				t.Error("the old wording conflated a missing page with a missing image")
			}
			if out.Len() != 0 {
				t.Errorf("stdout = %q, want nothing on a miss", out.String())
			}
		})
	}
}

func TestRunImage_PlainHitPrintsOnlyTheURL(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runImage(wikiStub(t), "Twisted bow", 150, false, false, &out, &errb); code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.TrimSpace(out.String()) != "https://wiki.test/150px-Twisted_bow_detail.png" {
		t.Errorf("stdout = %q, want the bare thumbnail URL", out.String())
	}
}
