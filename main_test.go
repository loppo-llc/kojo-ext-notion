package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestChildEnv(t *testing.T) {
	got := childEnv([]string{"PATH=/bin", "KOJO_EXT_TOKEN=x", "Kojo_Ext_Secrets={}", "NOTION_TOKEN=old", "KOJO_API_BASE=b"}, "ntn_1")
	want := []string{"PATH=/bin", "KOJO_API_BASE=b", "NOTION_TOKEN=ntn_1"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestFetchSecrets(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" || r.URL.Query().Get("agentId") != "ag_1" {
			http.Error(w, "nope", 403)
			return
		}
		w.Write([]byte(`{"secrets":{"token":"ntn_x"}}`))
	}))
	defer ts.Close()
	s, err := fetchSecrets(context.Background(), ts.URL, "tok", "ag_1")
	if err != nil || s["token"] != "ntn_x" {
		t.Fatalf("%v %v", s, err)
	}
	if _, err := fetchSecrets(context.Background(), ts.URL, "bad", "ag_1"); err == nil {
		t.Fatal("want error")
	}
}
