package http

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type FaviconRequest struct {
	URL string `json:"url"`
}

type FaviconResponse struct {
	IconURL string `json:"iconUrl,omitempty"`
	Message string `json:"message,omitempty"`
}

type FaviconHandler struct{}

func NewFaviconHandler() *FaviconHandler {
	return &FaviconHandler{}
}

func (h *FaviconHandler) DetectFavicon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FaviconRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Validate URL
	baseURL, err := url.Parse(req.URL)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Fetch the webpage
	resp, err := http.Get(req.URL)
	if err != nil {
		http.Error(w, "Error fetching URL", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		http.Error(w, "Error parsing HTML", http.StatusInternalServerError)
		return
	}

	// Check for apple-touch-icon
	if iconURL := findAppleTouchIcon(doc, baseURL); iconURL != "" {
		json.NewEncoder(w).Encode(FaviconResponse{IconURL: iconURL})
		return
	}

	// Check for manifest.json
	if iconURL := findManifestIcon(doc, baseURL); iconURL != "" {
		json.NewEncoder(w).Encode(FaviconResponse{IconURL: iconURL})
		return
	}

	// Check for traditional favicon
	if iconURL := findTraditionalFavicon(doc, baseURL); iconURL != "" {
		json.NewEncoder(w).Encode(FaviconResponse{IconURL: iconURL})
		return
	}

	// Try default favicon location
	if iconURL := tryDefaultFavicon(baseURL); iconURL != "" {
		json.NewEncoder(w).Encode(FaviconResponse{IconURL: iconURL})
		return
	}

	http.Error(w, "No favicon found", http.StatusNotFound)
}

func findAppleTouchIcon(doc *goquery.Document, baseURL *url.URL) string {
	var iconURL string
	doc.Find("link[rel='apple-touch-icon']").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if resolvedURL, err := resolveURL(baseURL, href); err == nil {
				iconURL = resolvedURL
				return
			}
		}
	})
	return iconURL
}

func findManifestIcon(doc *goquery.Document, baseURL *url.URL) string {
	var iconURL string
	doc.Find("link[rel='manifest']").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			manifestURL, err := resolveURL(baseURL, href)
			if err != nil {
				return
			}

			resp, err := http.Get(manifestURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			var manifest struct {
				Icons []struct {
					Src   string `json:"src"`
					Sizes string `json:"sizes"`
				} `json:"icons"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
				return
			}

			if len(manifest.Icons) > 0 {
				// Find the largest icon
				largestIcon := manifest.Icons[0]
				for _, icon := range manifest.Icons[1:] {
					if icon.Sizes > largestIcon.Sizes {
						largestIcon = icon
					}
				}

				if resolvedURL, err := resolveURL(baseURL, largestIcon.Src); err == nil {
					iconURL = resolvedURL
				}
			}
		}
	})
	return iconURL
}

func findTraditionalFavicon(doc *goquery.Document, baseURL *url.URL) string {
	var iconURL string
	doc.Find("link[rel='icon'], link[rel='shortcut icon']").Each(func(_ int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if resolvedURL, err := resolveURL(baseURL, href); err == nil {
				iconURL = resolvedURL
				return
			}
		}
	})
	return iconURL
}

func tryDefaultFavicon(baseURL *url.URL) string {
	faviconURL := baseURL.ResolveReference(&url.URL{Path: "/favicon.ico"}).String()
	resp, err := http.Head(faviconURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		return faviconURL
	}
	return ""
}

func resolveURL(baseURL *url.URL, href string) (string, error) {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href, nil
	}
	resolvedURL := baseURL.ResolveReference(&url.URL{Path: href})
	return resolvedURL.String(), nil
}
