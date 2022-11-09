package internal

import (
	u "net/url"
	"testing"
)

func TestDistinct(t *testing.T) {
	dc := NewDistinctChannel[u.URL](10)
	url, err := u.Parse("https://example.com")
	url2, err := u.Parse("https://example.com")
	if err != nil {
		t.Errorf("%v", err)
	}
	distinct := dc.Add(*url)
	if !distinct {
		t.Errorf("first insert must always be distinct")
	}
	distinct = dc.Add(*url2)
	if distinct {
		t.Errorf("second insert must never be distinct")
	}
}
