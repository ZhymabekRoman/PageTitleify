package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type Page struct {
	URL string `json:"url"`
}

func checkURL(URL string) (*url.URL, error) {
	URLParse, err := url.Parse(URL)
	if err != nil || URLParse.Host == "" {
		return nil, errors.New("invalid URL")
	}
	return URLParse, nil
}

func home(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	log.Println(string(body))

	var page Page
	err = json.Unmarshal(body, &page)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_, err = checkURL(page.URL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp, err := http.Get(page.URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			w.WriteHeader(resp.StatusCode)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	meta := extract(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(meta)
}

type HTMLMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func extract(resp io.Reader) *HTMLMeta {
	tokenizer := html.NewTokenizer(resp)
	hm := new(HTMLMeta)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return hm
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "title" {
				tokenType = tokenizer.Next()
				if tokenType == html.TextToken {
					token = tokenizer.Token()
					hm.Title = token.Data
				}
			} else if token.Data == "meta" {
				for _, attr := range token.Attr {
					if attr.Key == "name" && attr.Val == "description" {
						for _, attr := range token.Attr {
							if attr.Key == "content" {
								hm.Description = attr.Val
							}
						}
					}
				}
			}
		}
	}
}

func main() {
	http.HandleFunc("/", home)
	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
