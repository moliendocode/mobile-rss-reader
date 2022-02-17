package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/mmcdole/gofeed"
)

type Feed struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type rssRequest struct {
	Links []string `json:"links"`
}

type dataHandlers struct {
	sync.Mutex
	Feeds []Feed `json:"feeds"`
}

func (h *dataHandlers) get(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var feeds rssRequest
	err = json.Unmarshal(bodyBytes, &feeds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Lock()

	for i := range feeds.Links {
		result := getData(feeds.Links[i])
		if result != nil {
			h.Feeds = append(h.Feeds, *result...)
		}
		i++
	}
	defer h.Unlock()

	jsonBytes, err := json.Marshal(h.Feeds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func newData() *dataHandlers {
	return &dataHandlers{
		Feeds: []Feed{},
	}
}

func main() {
	dataHandlers := newData()
	http.HandleFunc("/", dataHandlers.get)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}

func getData(feedUrl string) *[]Feed {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedUrl)
	if err != nil {
		return nil
	}

	var feeds []Feed
	for i := range feed.Items {
		feeds = append(feeds, Feed{
			Title:       feed.Items[i].Title,
			Link:        feed.Items[i].Link,
			Description: feed.Items[i].Description,
			Image:       feed.Items[i].Image.URL,
		})
	}

	return &feeds
}
