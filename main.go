// Based on https://gist.github.com/inotnako/c4a82f6723f6ccea5d83c5d3689373dd
// https://medium.com/@dave-jaydeep/golang-parse-extract-html-2599db904e7d

package main

import (
	"encoding/json"
	"errors"
	// "fmt"
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
	URLParse, err := url.Parse(URL);
    if err != nil || URLParse.Host == "" {  // URLParse.Scheme == "" ||
        return URLParse, errors.New("Invalid URL!")
	}
    return URLParse, nil
}

func home(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))

	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	var page Page
	err = json.Unmarshal(body, &page)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")

    // ParseType := r.FormValue(`parseType`)
	// switch ParseType {
	// case "browser":
	//	_, err = checkURL(page.URL, w, r)
	//	if err != nil {
	//		return
	//	}
	//	w.WriteHeader(http.StatusOK)
	//	w.Write([]byte(page.URL))
	// case "simple":
	_, err = checkURL(page.URL)
    if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp, err := http.Get(page.URL)
	if err != nil {
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
	}

	meta := extract(resp.Body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(meta)
}


type HTMLMeta struct {
	Title       string `json:"title"`
    /*
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
    */
}

// meta property extracter is broken
func extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
	for _, attr := range t.Attr {
		if attr.Key == "content" && attr.Val == prop {
            log.Println("All is OK~! 111111")
			ok = true
		}

		if attr.Key == "content" {
                        log.Println("All is OK~! 222222")
			content = attr.Val
		}
	}

	return
}

func extract(resp io.Reader) *HTMLMeta {
	tokenizer := html.NewTokenizer(resp)
	hm := new(HTMLMeta)

	for {
		tokenType := tokenizer.Next()
        token := tokenizer.Token()
		switch tokenType {
		case html.ErrorToken:
			return hm
		case html.StartTagToken, html.SelfClosingTagToken:
            //fmt.Println(tokenType)
            //fmt.Println(token.Data)
            //fmt.Println(token.String())

			if token.Data == `body` {
				return hm
			}
            // log.Println(t.Data)
			if token.Data == "title" {
                    tokenizer.Next()
                    token := tokenizer.Token()
                    hm.Title = token.String()
			}
            /*
			if token.Data == "meta" {
				desc, ok := extractMetaProperty(token, "description")
				if ok {
					hm.Description = desc
				}

				ogTitle, ok := extractMetaProperty(token, "og:title")
				if ok {
					hm.Title = ogTitle
				}

				ogDesc, ok := extractMetaProperty(token, "og:description")
				if ok {
					hm.Description = ogDesc
				}

				ogImage, ok := extractMetaProperty(token, "og:image")
				if ok {
					hm.Image = ogImage
				}

				ogSiteName, ok := extractMetaProperty(token, "og:site_name")
				if ok {
					hm.SiteName = ogSiteName
				}
			}
            */
		}
	}
	return hm
}

func main() {
	http.HandleFunc("/", home)
    log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
