package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ultima-pista/ultima-pista/internal/api"
	"github.com/ultima-pista/ultima-pista/internal/store"
)

func setup(t *testing.T) (*httptest.Server, *http.Client, *http.Client) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	srv := api.New(st, false, "")
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	jar1, _ := cookiejar.New(nil)
	jar2, _ := cookiejar.New(nil)
	c1 := &http.Client{Jar: jar1}
	c2 := &http.Client{Jar: jar2}
	return ts, c1, c2
}

func postJSON(t *testing.T, c *http.Client, url string, body any) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return res.StatusCode, out
}

func getJSON(t *testing.T, c *http.Client, url string) (int, map[string]any) {
	t.Helper()
	res, err := c.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return res.StatusCode, out
}

func TestRemoteFlowAccessAndReveal(t *testing.T) {
	ts, narr, det := setup(t)

	st, body := postJSON(t, narr, ts.URL+"/api/rooms", map[string]string{
		"nickname": "Narrador", "mode": "remoto",
	})
	if st != http.StatusCreated {
		t.Fatalf("create: %d %v", st, body)
	}
	code := body["code"].(string)

	st, body = postJSON(t, det, ts.URL+"/api/rooms/join", map[string]string{
		"code": code, "nickname": "Lia",
	})
	if st != http.StatusOK {
		t.Fatalf("join: %d %v", st, body)
	}

	// Detective cannot read private
	st, body = getJSON(t, det, ts.URL+"/api/rooms/"+code+"/private")
	if st != http.StatusForbidden {
		t.Fatalf("detective private: %d %v", st, body)
	}

	// Telão snapshot has no you / no secrets
	st, telao := getJSON(t, narr, ts.URL+"/api/rooms/"+code+"/telao")
	if st != http.StatusOK {
		t.Fatalf("telao: %d", st)
	}
	if _, ok := telao["you"]; ok && telao["you"] != nil {
		t.Fatalf("telao should not expose you: %v", telao["you"])
	}

	st, body = postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/start", map[string]string{
		"caseId": "biblioteca-meia-noite",
	})
	if st != http.StatusOK {
		t.Fatalf("start: %d %v", st, body)
	}

	// Private for narrator works, no solution in public
	st, priv := getJSON(t, narr, ts.URL+"/api/rooms/"+code+"/private")
	if st != http.StatusOK {
		t.Fatalf("private: %d %v", st, priv)
	}
	if priv["solution"] == nil || priv["solution"] == "" {
		t.Fatal("expected solution for narrator")
	}
	st, pub := getJSON(t, det, ts.URL+"/api/rooms/"+code)
	if st != http.StatusOK {
		t.Fatal(st)
	}
	round := pub["round"].(map[string]any)
	if round["revealedSolution"] != nil {
		t.Fatal("solution leaked before reveal")
	}
	caseObj := round["case"].(map[string]any)
	raw, _ := json.Marshal(caseObj)
	if bytes.Contains(bytes.ToLower(raw), []byte("gato")) && bytes.Contains(raw, []byte("pegadas")) {
		// solution mentions these; public text should not have full solution
	}
	sol := fmt.Sprint(priv["solution"])
	pubRaw, _ := json.Marshal(pub)
	if bytes.Contains(pubRaw, []byte(sol[:40])) {
		t.Fatal("solution fragment in public snapshot")
	}

	// Detective asks
	st, q := postJSON(t, det, ts.URL+"/api/rooms/"+code+"/questions", map[string]string{
		"text": "Havia um animal?", "idempotencyKey": "q1",
	})
	if st != http.StatusCreated {
		t.Fatalf("ask: %d %v", st, q)
	}
	qid := q["id"].(string)

	// Detective cannot answer
	st, _ = postJSON(t, det, ts.URL+"/api/rooms/"+code+"/questions/"+qid+"/answer", map[string]string{
		"answer": "sim",
	})
	if st != http.StatusForbidden {
		t.Fatalf("detective answer should be forbidden: %d", st)
	}

	st, _ = postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/questions/"+qid+"/answer", map[string]string{
		"answer": "sim",
	})
	if st != http.StatusOK {
		t.Fatalf("answer: %d", st)
	}

	st, th := postJSON(t, det, ts.URL+"/api/rooms/"+code+"/theories", map[string]string{
		"text": "O gato da bibliotecária entrou e um invasor fugiu.", "idempotencyKey": "t1",
	})
	if st != http.StatusCreated {
		t.Fatalf("theory: %d %v", st, th)
	}
	tid := th["id"].(string)

	st, _ = postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/theories/"+tid+"/confirm", map[string]any{})
	if st != http.StatusOK {
		t.Fatalf("confirm: %d", st)
	}

	st, pub = getJSON(t, det, ts.URL+"/api/rooms/"+code)
	round = pub["round"].(map[string]any)
	if round["status"] != "resolvida" {
		t.Fatalf("expected resolvida: %v", round["status"])
	}
	if round["revealedSolution"] == nil {
		t.Fatal("expected revealed solution")
	}

	// Telão also gets solution after reveal
	st, telao = getJSON(t, det, ts.URL+"/api/rooms/"+code+"/telao")
	if st != http.StatusOK {
		t.Fatal(st)
	}
	tr := telao["round"].(map[string]any)
	if tr["revealedSolution"] == nil {
		t.Fatal("telao missing revealed solution")
	}
}

func TestIdempotencyAndConcurrentAnswer(t *testing.T) {
	ts, narr, det := setup(t)
	_, body := postJSON(t, narr, ts.URL+"/api/rooms", map[string]string{"nickname": "N", "mode": "remoto"})
	code := body["code"].(string)
	postJSON(t, det, ts.URL+"/api/rooms/join", map[string]string{"code": code, "nickname": "D"})
	postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/start", map[string]string{"caseId": "carta-sem-remetente"})

	st1, q1 := postJSON(t, det, ts.URL+"/api/rooms/"+code+"/questions", map[string]string{
		"text": "O carteiro está envolvido?", "idempotencyKey": "same-key",
	})
	st2, q2 := postJSON(t, det, ts.URL+"/api/rooms/"+code+"/questions", map[string]string{
		"text": "O carteiro está envolvido?", "idempotencyKey": "same-key",
	})
	if st1 != http.StatusCreated || st2 != http.StatusCreated {
		t.Fatalf("idem create %d %d", st1, st2)
	}
	if q1["id"] != q2["id"] {
		t.Fatalf("idempotency failed: %v vs %v", q1["id"], q2["id"])
	}
	qid := q1["id"].(string)

	jar3, _ := cookiejar.New(nil)
	narr2 := &http.Client{Jar: jar3}
	// Copy narrator session by creating room won't work — reuse narr jar via cloning cookies
	// Use same narr client twice concurrently instead
	var okCount int32
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			st, _ := postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/questions/"+qid+"/answer", map[string]string{"answer": "sim"})
			if st == http.StatusOK {
				atomic.AddInt32(&okCount, 1)
			}
		}()
	}
	wg.Wait()
	if okCount != 1 {
		t.Fatalf("expected exactly one successful answer, got %d", okCount)
	}
	_ = narr2
}

func TestPresencialOralAndOldRoundRejected(t *testing.T) {
	ts, narr, _ := setup(t)
	_, body := postJSON(t, narr, ts.URL+"/api/rooms", map[string]string{"nickname": "N", "mode": "presencial"})
	code := body["code"].(string)
	postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/start", map[string]string{"caseId": "jardim-congelado"})

	st, q := postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/oral", map[string]string{
		"text": "As esculturas eram de gelo?", "answer": "sim", "idempotencyKey": "oral1",
	})
	if st != http.StatusCreated {
		t.Fatalf("oral: %d %v", st, q)
	}
	oldQ := q["id"].(string)

	st, _ = postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/confirm-solution", map[string]string{
		"summary": "Festival de gelo adulterado",
	})
	if st != http.StatusOK {
		t.Fatalf("confirm spoken: %d", st)
	}
	postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/lobby", map[string]any{})
	postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/start", map[string]string{"caseId": "foto-do-casamento"})

	st, body = postJSON(t, narr, ts.URL+"/api/rooms/"+code+"/questions/"+oldQ+"/answer", map[string]string{"answer": "nao"})
	if st == http.StatusOK {
		t.Fatal("answering old round question should fail")
	}
}

func TestCasesHaveNoSolutionInList(t *testing.T) {
	ts, c, _ := setup(t)
	res, err := c.Get(ts.URL + "/api/cases")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if bytes.Contains(bytes.ToLower(raw), []byte("solution")) {
		t.Fatal("solution key in public cases")
	}
	var list []map[string]any
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatal(err)
	}
	if len(list) < 5 {
		t.Fatalf("need >=5 cases, got %d", len(list))
	}
}
