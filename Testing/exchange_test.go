package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"base":"USD","target":"EUR","rate":0.9}`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	rate, err := service.GetRate("USD", "EUR")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if rate != 0.9 {
		t.Errorf("got %f, want 0.9", rate)
	}
}

func TestGetRate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, `{"error":"invalid currency pair"}`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("AAA", "BBB")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestGetRate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `invalid json`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestGetRate_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestGetRate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{ "error": "server panic" }`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestGetRate_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second)
		fmt.Fprintln(w, `{"base":"USD","target":"EUR","rate":0.9}`)
	}))
	defer server.Close()

	service := NewExchangeService(server.URL)

	_, err := service.GetRate("USD", "EUR")
	if err == nil {
		t.Errorf("expected error")
	}
}
