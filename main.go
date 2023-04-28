// Based on https://gist.github.com/inotnako/c4a82f6723f6ccea5d83c5d3689373dd

package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
    "errors"
)

type Page struct {
	URL string `json:"url"`
}

func checkURL(URL string, w http.ResponseWriter, r *http.Request) (*url.URL, error) {
        // Check: https://.com
        fmt.Printf("BRRRRRRRRRRRRRRRRRRRRR %s", URL)
	URLParse, err := url.Parse(URL);
    if err != nil || URLParse.Scheme == "" || URLParse.Host == "" {
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

	fmt.Printf("Link: %s", page.URL)

	w.Header().Set("Content-Type", "application/json")

	ParseType := r.FormValue(`parse_type`)
	switch ParseType {
	case "browser":
		_, err = checkURL(page.URL, w, r)
		if err != nil {
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(page.URL))
	case "simple":
		_, err = checkURL(page.URL, w, r)
		if err != nil {
			fmt.Println("ASDDDDDDDDDDDDDDDDDDDDDDDDD")
    		w.WriteHeader(http.StatusBadRequest)
	    	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
		}

		resp, err := http.Get(page.URL)
		if err != nil {
			//proxy status and err
			w.WriteHeader(resp.StatusCode)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		meta := extract(resp.Body)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meta)

	default:
		ErrorMessage := fmt.Sprintf(`Unknown parse_type value: %s`, ParseType)
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]string{"error": ErrorMessage}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonResponse)
		return
	}

}

type HTMLMeta struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	SiteName    string `json:"site_name"`
}

func extractMetaProperty(t html.Token, prop string) (content string, ok bool) {
	for _, attr := range t.Attr {
		if attr.Key == "property" && attr.Val == prop {
			ok = true
		}

		if attr.Key == "content" {
			content = attr.Val
		}
	}

	return
}

func extract(resp io.Reader) *HTMLMeta {
	z := html.NewTokenizer(resp)

	titleFound := false

	hm := new(HTMLMeta)

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return hm
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == `body` {
				return hm
			}
			if t.Data == "title" {
				titleFound = true
			}
			if t.Data == "meta" {
				desc, ok := extractMetaProperty(t, "description")
				if ok {
					hm.Description = desc
				}

				ogTitle, ok := extractMetaProperty(t, "og:title")
				if ok {
					hm.Title = ogTitle
				}

				ogDesc, ok := extractMetaProperty(t, "og:description")
				if ok {
					hm.Description = ogDesc
				}

				ogImage, ok := extractMetaProperty(t, "og:image")
				if ok {
					hm.Image = ogImage
				}

				ogSiteName, ok := extractMetaProperty(t, "og:site_name")
				if ok {
					hm.SiteName = ogSiteName
				}
			}
		case html.TextToken:
			if titleFound {
				t := z.Token()
				hm.Title = t.Data
				titleFound = false
			}
		}
	}
	return hm
}

func main() {
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
