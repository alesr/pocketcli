package pocketcli

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type httpClientMock struct {
	DoFunc func(*http.Request) (*http.Response, error)
}

func (m *httpClientMock) Do(r *http.Request) (*http.Response, error) {
	return m.DoFunc(r)
}

func TestNew(t *testing.T) {
	t.Run("should return a new client with all fields correctly initialized", func(t *testing.T) {
		// Arrange

		var (
			httpCli     = &http.Client{}
			host        = "http://localhost:8080"
			consumerKey = "consumerKey"
			accessToken = "accessToken"
			username    = "username"
		)

		// Act

		client, err := New(httpCli, host, consumerKey, accessToken, username)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Assert

		if client.host != host {
			t.Errorf("expected host to be %s, got %s", host, client.host)
		}

		if client.consumerKey != consumerKey {
			t.Errorf("expected consumerKey to be %s, got %s", consumerKey, client.consumerKey)
		}

		if client.accessToken != accessToken {
			t.Errorf("expected accessToken to be %s, got %s", accessToken, client.accessToken)
		}

		if client.username != username {
			t.Errorf("expected username to be %s, got %s", username, client.username)
		}

		if client.httpCli != httpCli {
			t.Errorf("expected httpCli to be %v, got %s", httpCli, client.httpCli)
		}

		if client.retrieveOpts == nil {
			t.Error("expected retrieveOpts to not be nil")
		}

		if client.retrieveOpts.String() != "{\"tag\":\"rmk\"}" {
			t.Errorf("expected retrieveOpts to be %s, got %s", "{\"tag\":\"rmk\"}", client.retrieveOpts.String())
		}
	})
}

func TestDo(t *testing.T) {
	t.Run("succesfully performs a request using the mock http client", func(t *testing.T) {
		// Arrange

		var (
			httpCli = httpClientMock{
				DoFunc: func(r *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			}
			host        = "http://localhost:8080"
			consumerKey = "consumerKey"
			accessToken = "accessToken"
			username    = "username"
		)

		client, err := New(&httpCli, host, consumerKey, accessToken, username)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Act

		resp, err := client.do(context.Background(), req, accessToken, http.StatusOK)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Assert

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status code to be %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("sets consumer key, access token as url params and content type header", func(t *testing.T) {

		var (
			host        = "http://localhost:8080"
			consumerKey = "consumerKey"
			accessToken = "accessToken"
			username    = "username"
			contentType = "application/json charset=utf-8"
		)

		var (
			httpCli = httpClientMock{
				DoFunc: func(r *http.Request) (*http.Response, error) {
					// Check if the consumer key and access token are set as url params
					if r.URL.Query().Get("consumer_key") != consumerKey {
						t.Errorf("expected consumer_key to be %s, got %s", consumerKey, r.URL.Query().Get("consumer_key"))
					}

					if r.URL.Query().Get("access_token") != accessToken {
						t.Errorf("expected access_token to be %s, got %s", accessToken, r.URL.Query().Get("access_token"))
					}

					// Check if the content type header is set
					if r.Header.Get("Content-Type") != contentType {
						t.Errorf("expected Content-Type to be %s, got %s", contentType, r.Header.Get("Content-Type"))
					}

					return &http.Response{
						StatusCode: http.StatusOK,
					}, nil
				},
			}
		)

		client, err := New(&httpCli, host, consumerKey, accessToken, username)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		_, err = client.do(context.Background(), req, accessToken, http.StatusOK)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("returns an error if the http client returns an error", func(t *testing.T) {
		// Arrange

		var givenErr = errors.New("given error")

		var (
			httpCli = httpClientMock{
				DoFunc: func(r *http.Request) (*http.Response, error) {
					return nil, givenErr
				},
			}
			host        = "http://localhost:8080"
			consumerKey = "consumerKey"
			accessToken = "accessToken"
			username    = "username"
		)

		client, err := New(&httpCli, host, consumerKey, accessToken, username)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// Act

		_, err = client.do(context.Background(), req, accessToken, http.StatusOK)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		// Assert

		if errors.Is(err, givenErr) == false {
			t.Errorf("expected error to be %s, got %s", givenErr, err)
		}

	})
}
