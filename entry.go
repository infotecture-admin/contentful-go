package contentful

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// EntriesService servıce
type EntriesService service

//Entry model
type Entry struct {
	locale string
	Sys    *Sys                   `json:"sys"`
	Fields map[string]interface{} `json:"fields"`
}

// GetVersion returns entity version
func (entry *Entry) GetVersion() int {
	version := 1
	if entry.Sys != nil {
		version = entry.Sys.Version
	}

	return version
}

// GetKey returns the entry's keys
func (service *EntriesService) GetEntryKey(entry *Entry, key string) (*EntryField, error) {
	ef := EntryField{
		value: entry.Fields[key],
	}

	col, err := service.c.ContentTypes.List(entry.Sys.Space.Sys.ID).Next()
	if err != nil {
		return nil, err
	}

	for _, ct := range col.ToContentType() {
		if ct.Sys.ID != entry.Sys.ContentType.Sys.ID {
			continue
		}

		for _, field := range ct.Fields {
			if field.ID != key {
				continue
			}

			ef.dataType = field.Type
		}
	}

	return &ef, nil
}

// List returns entries collection
func (service *EntriesService) List(spaceID string) *Collection {
	path := fmt.Sprintf("/spaces/%s/entries", spaceID)
	method := "GET"

	req, err := service.c.newRequest(method, path, nil, nil)
	if err != nil {
		return &Collection{}
	}

	col := NewCollection(&CollectionOptions{})
	col.c = service.c
	col.req = req

	return col
}

// Get returns a single entry
func (service *EntriesService) Get(spaceID, entryID string) (*Entry, error) {
	path := fmt.Sprintf("/spaces/%s/entries/%s", spaceID, entryID)
	query := url.Values{}

	req, err := service.c.newRequest("GET", path, query, nil)
	if err != nil {
		return &Entry{}, err
	}

	var entry Entry
	if ok := service.c.do(req, &entry); ok != nil {
		return nil, err
	}

	return &entry, err
}

func (service *EntriesService) Upsert(spaceID, contentType string, entry *Entry) error {
	bytesArray, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	var path string
	var method string

	if entry.Sys != nil && entry.Sys.CreatedAt != "" {
		path = fmt.Sprintf("/spaces/%s/entries/%s", spaceID, entry.Sys.ID)
		method = "PUT"
	} else {
		path = fmt.Sprintf("/spaces/%s/entries", spaceID)
		method = "POST"
	}

	req, err := service.c.newRequest(method, path, nil, bytes.NewReader(bytesArray))
	if err != nil {
		return err
	}

	req.Header.Set("X-Contentful-Version", strconv.Itoa(entry.GetVersion()))
	req.Header.Set("X-Contentful-Content-Type", contentType)

	if err = service.c.do(req, entry); err != nil {
		return err
	}

	return nil
}
