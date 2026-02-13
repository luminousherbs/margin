package api

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"

	"margin.at/internal/db"
)

//go:embed fonts/Inter-Regular.ttf
var interRegularTTF []byte

//go:embed fonts/Inter-Bold.ttf
var interBoldTTF []byte

//go:embed fonts/DroidSansFallback.ttf
var droidSansFallbackTTF []byte

//go:embed assets/logo.png
var logoPNG []byte

var (
	fontRegular  *opentype.Font
	fontBold     *opentype.Font
	fontFallback *opentype.Font
	logoImage    image.Image
)

func init() {
	var err error
	fontRegular, err = opentype.Parse(interRegularTTF)
	if err != nil {
		log.Printf("Warning: failed to parse Inter-Regular font: %v", err)
	}
	fontBold, err = opentype.Parse(interBoldTTF)
	if err != nil {
		log.Printf("Warning: failed to parse Inter-Bold font: %v", err)
	}
	fontFallback, err = opentype.Parse(droidSansFallbackTTF)
	if err != nil {
		log.Printf("Warning: failed to parse DroidSansFallback font: %v", err)
	}

	if len(logoPNG) > 0 {
		img, _, err := image.Decode(bytes.NewReader(logoPNG))
		if err != nil {
			log.Printf("Warning: failed to decode logo PNG: %v", err)
		} else {
			logoImage = img
		}
	}
}

func drawText(img *image.RGBA, text string, x, y int, c color.Color, size float64, bold bool) {
	if fontRegular == nil || fontBold == nil {
		return
	}

	primaryFont := fontRegular
	if bold {
		primaryFont = fontBold
	}

	opts := &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	}

	facePrimary, _ := opentype.NewFace(primaryFont, opts)
	defer facePrimary.Close()

	var faceFallback font.Face
	if fontFallback != nil {
		faceFallback, _ = opentype.NewFace(fontFallback, opts)
		defer faceFallback.Close()
	}

	dPrimary := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: facePrimary,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}

	var dFallback *font.Drawer
	if faceFallback != nil {
		dFallback = &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(c),
			Face: faceFallback,
			Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
		}
	}

	var buf sfnt.Buffer
	for _, r := range text {
		useFallback := false
		if fontFallback != nil {
			idx, err := primaryFont.GlyphIndex(&buf, r)
			if err != nil || idx == 0 {
				useFallback = true
			}
		}

		if useFallback {
			dFallback.Dot = dPrimary.Dot

			dFallback.DrawString(string(r))

			dPrimary.Dot = dFallback.Dot
		} else {
			dPrimary.DrawString(string(r))
		}
	}
}

type OGHandler struct {
	db        *db.DB
	baseURL   string
	staticDir string
}

func NewOGHandler(database *db.DB) *OGHandler {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://margin.at"
	}
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "../web/dist"
	}
	return &OGHandler{
		db:        database,
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		staticDir: staticDir,
	}
}

var crawlerUserAgents = []string{
	"facebookexternalhit",
	"Facebot",
	"Twitterbot",
	"LinkedInBot",
	"WhatsApp",
	"Slackbot",
	"TelegramBot",
	"Discordbot",
	"applebot",
	"bot",
	"crawler",
	"spider",
	"preview",
	"Cardyb",
	"Bluesky",
}

var lucideToEmoji = map[string]string{
	"folder":    "📁",
	"star":      "⭐",
	"heart":     "❤️",
	"bookmark":  "🔖",
	"lightbulb": "💡",
	"zap":       "⚡",
	"coffee":    "☕",
	"music":     "🎵",
	"camera":    "📷",
	"code":      "💻",
	"globe":     "🌍",
	"flag":      "🚩",
	"tag":       "🏷️",
	"box":       "📦",
	"archive":   "🗄️",
	"file":      "📄",
	"image":     "🖼️",
	"video":     "🎬",
	"mail":      "✉️",
	"pin":       "📍",
	"calendar":  "📅",
	"clock":     "🕐",
	"search":    "🔍",
	"settings":  "⚙️",
	"user":      "👤",
	"users":     "👥",
	"home":      "🏠",
	"briefcase": "💼",
	"gift":      "🎁",
	"award":     "🏆",
	"target":    "🎯",
	"trending":  "📈",
	"activity":  "📊",
	"cpu":       "🔲",
	"database":  "🗃️",
	"cloud":     "☁️",
	"sun":       "☀️",
	"moon":      "🌙",
	"flame":     "🔥",
	"leaf":      "🍃",
}

func iconToEmoji(icon string) string {
	if strings.HasPrefix(icon, "icon:") {
		name := strings.TrimPrefix(icon, "icon:")
		if emoji, ok := lucideToEmoji[name]; ok {
			return emoji
		}
		return "📁"
	}
	return icon
}

func isCrawler(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	for _, bot := range crawlerUserAgents {
		if strings.Contains(ua, strings.ToLower(bot)) {
			return true
		}
	}
	return false
}

func (h *OGHandler) resolveHandle(handle string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://public.api.bsky.app/xrpc/com.atproto.identity.resolveHandle?handle=%s", url.QueryEscape(handle)))
	if err == nil && resp.StatusCode == http.StatusOK {
		var result struct {
			Did string `json:"did"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.Did != "" {
			return result.Did, nil
		}
	}
	defer resp.Body.Close()

	return "", fmt.Errorf("failed to resolve handle")
}

func (h *OGHandler) HandleAnnotationPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var did, rkey, collectionType string

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		firstPart, _ := url.QueryUnescape(parts[0])

		if firstPart == "at" || firstPart == "annotation" {
			if len(parts) >= 3 {
				did, _ = url.QueryUnescape(parts[1])
				rkey = parts[2]
			}
		} else {
			if len(parts) >= 3 {
				var err error
				did, err = h.resolveHandle(firstPart)
				if err != nil {
					h.serveIndexHTML(w, r)
					return
				}

				switch parts[1] {
				case "highlight":
					collectionType = "at.margin.highlight"
				case "bookmark":
					collectionType = "at.margin.bookmark"
				case "annotation":
					collectionType = "at.margin.annotation"
				}
				rkey = parts[2]
			}
		}
	}

	if did == "" || rkey == "" {
		h.serveIndexHTML(w, r)
		return
	}

	if !isCrawler(r.UserAgent()) {
		h.serveIndexHTML(w, r)
		return
	}

	if collectionType != "" {
		uri := fmt.Sprintf("at://%s/%s/%s", did, collectionType, rkey)
		if h.tryServeType(w, uri, collectionType) {
			return
		}
	} else {
		types := []string{
			"at.margin.annotation",
			"at.margin.bookmark",
			"at.margin.highlight",
		}
		for _, t := range types {
			uri := fmt.Sprintf("at://%s/%s/%s", did, t, rkey)
			if h.tryServeType(w, uri, t) {
				return
			}
		}

		colURI := fmt.Sprintf("at://%s/at.margin.collection/%s", did, rkey)
		if h.tryServeType(w, colURI, "at.margin.collection") {
			return
		}
	}

	h.serveIndexHTML(w, r)
}

func (h *OGHandler) tryServeType(w http.ResponseWriter, uri, colType string) bool {
	switch colType {
	case "at.margin.annotation":
		if item, err := h.db.GetAnnotationByURI(uri); err == nil && item != nil {
			h.serveAnnotationOG(w, item)
			return true
		}
	case "at.margin.highlight":
		if item, err := h.db.GetHighlightByURI(uri); err == nil && item != nil {
			h.serveHighlightOG(w, item)
			return true
		}
	case "at.margin.bookmark":
		if item, err := h.db.GetBookmarkByURI(uri); err == nil && item != nil {
			h.serveBookmarkOG(w, item)
			return true
		}
	case "at.margin.collection":
		if item, err := h.db.GetCollectionByURI(uri); err == nil && item != nil {
			h.serveCollectionOG(w, item)
			return true
		}
	}
	return false
}

func (h *OGHandler) HandleCollectionPage(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	var did, rkey string

	if strings.Contains(path, "/collection/") {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 3 && parts[1] == "collection" {
			handle, _ := url.QueryUnescape(parts[0])
			rkey = parts[2]
			var err error
			did, err = h.resolveHandle(handle)
			if err != nil {
				h.serveIndexHTML(w, r)
				return
			}
		} else if strings.HasPrefix(path, "/collection/") {
			uriParam := strings.TrimPrefix(path, "/collection/")
			if uriParam != "" {
				uri, err := url.QueryUnescape(uriParam)
				if err == nil {
					parts := strings.Split(uri, "/")
					if len(parts) >= 3 && strings.HasPrefix(uri, "at://") {
						did = parts[2]
						rkey = parts[len(parts)-1]
					}
				}
			}
		}
	}

	if did == "" && rkey == "" {
		h.serveIndexHTML(w, r)
		return
	} else if did != "" && rkey != "" {
		uri := fmt.Sprintf("at://%s/at.margin.collection/%s", did, rkey)

		if !isCrawler(r.UserAgent()) {
			h.serveIndexHTML(w, r)
			return
		}

		collection, err := h.db.GetCollectionByURI(uri)
		if err == nil && collection != nil {
			h.serveCollectionOG(w, collection)
			return
		}
	}

	h.serveIndexHTML(w, r)
}

func (h *OGHandler) serveBookmarkOG(w http.ResponseWriter, bookmark *db.Bookmark) {
	title := "Bookmark on Margin"
	if bookmark.Title != nil && *bookmark.Title != "" {
		title = *bookmark.Title
	}

	description := ""
	if bookmark.Description != nil && *bookmark.Description != "" {
		description = *bookmark.Description
	} else {
		description = "A saved bookmark on Margin"
	}

	sourceDomain := ""
	if bookmark.Source != "" {
		if parsed, err := url.Parse(bookmark.Source); err == nil {
			sourceDomain = parsed.Host
		}
	}

	if sourceDomain != "" {
		description += " from " + sourceDomain
	}

	authorHandle := bookmark.AuthorDID
	profiles := fetchProfilesForDIDs(h.db, []string{bookmark.AuthorDID})
	if profile, ok := profiles[bookmark.AuthorDID]; ok && profile.Handle != "" {
		authorHandle = "@" + profile.Handle
	}

	pageURL := fmt.Sprintf("%s/at/%s", h.baseURL, url.PathEscape(bookmark.URI[5:]))
	ogImageURL := fmt.Sprintf("%s/og-image?uri=%s", h.baseURL, url.QueryEscape(bookmark.URI))

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Margin</title>
    <meta name="description" content="%s">
    
    <!-- Open Graph -->
    <meta property="og:type" content="article">
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:url" content="%s">
    <meta property="og:image" content="%s">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta property="og:site_name" content="Margin">
    
    <!-- Twitter Card -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="%s">
    <meta name="twitter:description" content="%s">
    <meta name="twitter:image" content="%s">
    
    <!-- Author -->
    <meta property="article:author" content="%s">
    
    <meta http-equiv="refresh" content="0; url=%s">
</head>
<body>
    <p>Redirecting to <a href="%s">%s</a>...</p>
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(pageURL),
		html.EscapeString(ogImageURL),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(ogImageURL),
		html.EscapeString(authorHandle),
		html.EscapeString(pageURL),
		html.EscapeString(pageURL),
		html.EscapeString(title),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

func (h *OGHandler) serveHighlightOG(w http.ResponseWriter, highlight *db.Highlight) {
	title := "Highlight on Margin"
	description := ""

	if highlight.SelectorJSON != nil && *highlight.SelectorJSON != "" {
		var selector struct {
			Exact string `json:"exact"`
		}
		if err := json.Unmarshal([]byte(*highlight.SelectorJSON), &selector); err == nil && selector.Exact != "" {
			description = fmt.Sprintf("\"%s\"", selector.Exact)
			if len(description) > 200 {
				description = description[:197] + "...\""
			}
		}
	}

	if highlight.TargetTitle != nil && *highlight.TargetTitle != "" {
		title = fmt.Sprintf("Highlight on: %s", *highlight.TargetTitle)
		if len(title) > 60 {
			title = title[:57] + "..."
		}
	}

	sourceDomain := ""
	if highlight.TargetSource != "" {
		if parsed, err := url.Parse(highlight.TargetSource); err == nil {
			sourceDomain = parsed.Host
		}
	}

	authorHandle := highlight.AuthorDID
	profiles := fetchProfilesForDIDs(h.db, []string{highlight.AuthorDID})
	if profile, ok := profiles[highlight.AuthorDID]; ok && profile.Handle != "" {
		authorHandle = "@" + profile.Handle
	}

	if description == "" {
		description = fmt.Sprintf("A highlight by %s", authorHandle)
		if sourceDomain != "" {
			description += fmt.Sprintf(" on %s", sourceDomain)
		}
	}

	pageURL := fmt.Sprintf("%s/at/%s", h.baseURL, url.PathEscape(highlight.URI[5:]))
	ogImageURL := fmt.Sprintf("%s/og-image?uri=%s", h.baseURL, url.QueryEscape(highlight.URI))

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Margin</title>
    <meta name="description" content="%s">
    
    <!-- Open Graph -->
    <meta property="og:type" content="article">
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:url" content="%s">
    <meta property="og:image" content="%s">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta property="og:site_name" content="Margin">
    
    <!-- Twitter Card -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="%s">
    <meta name="twitter:description" content="%s">
    <meta name="twitter:image" content="%s">
    
    <!-- Author -->
    <meta property="article:author" content="%s">
    
    <meta http-equiv="refresh" content="0; url=%s">
</head>
<body>
    <p>Redirecting to <a href="%s">%s</a>...</p>
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(pageURL),
		html.EscapeString(ogImageURL),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(ogImageURL),
		html.EscapeString(authorHandle),
		html.EscapeString(pageURL),
		html.EscapeString(pageURL),
		html.EscapeString(title),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

func (h *OGHandler) serveCollectionOG(w http.ResponseWriter, collection *db.Collection) {
	icon := "📁"
	if collection.Icon != nil && *collection.Icon != "" {
		icon = iconToEmoji(*collection.Icon)
	}

	title := fmt.Sprintf("%s %s", icon, collection.Name)
	description := ""
	if collection.Description != nil && *collection.Description != "" {
		description = *collection.Description
		if len(description) > 200 {
			description = description[:197] + "..."
		}
	}

	authorHandle := collection.AuthorDID
	var avatarURL string
	profiles := fetchProfilesForDIDs(h.db, []string{collection.AuthorDID})
	if profile, ok := profiles[collection.AuthorDID]; ok {
		if profile.Handle != "" {
			authorHandle = "@" + profile.Handle
		}
		if profile.Avatar != "" {
			avatarURL = profile.Avatar
		}
	}

	if description == "" {
		description = fmt.Sprintf("A collection by %s", authorHandle)
	} else {
		description = fmt.Sprintf("By %s • %s", authorHandle, description)
	}

	pageURL := fmt.Sprintf("%s/collection/%s", h.baseURL, url.PathEscape(collection.URI))
	ogImageURL := fmt.Sprintf("%s/og-image?uri=%s", h.baseURL, url.QueryEscape(collection.URI))

	_ = avatarURL

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Margin</title>
    <meta name="description" content="%s">
    
    <!-- Open Graph -->
    <meta property="og:type" content="article">
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:url" content="%s">
    <meta property="og:image" content="%s">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta property="og:site_name" content="Margin">
    
    <!-- Twitter Card -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="%s">
    <meta name="twitter:description" content="%s">
    <meta name="twitter:image" content="%s">
    
    <!-- Author -->
    <meta property="article:author" content="%s">
    
    <meta http-equiv="refresh" content="0; url=%s">
</head>
<body>
    <p>Redirecting to <a href="%s">%s</a>...</p>
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(pageURL),
		html.EscapeString(ogImageURL),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(ogImageURL),
		html.EscapeString(authorHandle),
		html.EscapeString(pageURL),
		html.EscapeString(pageURL),
		html.EscapeString(title),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

func (h *OGHandler) serveAnnotationOG(w http.ResponseWriter, annotation *db.Annotation) {
	title := "Annotation on Margin"
	description := ""

	if annotation.BodyValue != nil && *annotation.BodyValue != "" {
		description = *annotation.BodyValue
		if len(description) > 200 {
			description = description[:197] + "..."
		}
	}

	if annotation.TargetTitle != nil && *annotation.TargetTitle != "" {
		title = fmt.Sprintf("Comment on: %s", *annotation.TargetTitle)
		if len(title) > 60 {
			title = title[:57] + "..."
		}
	}

	sourceDomain := ""
	if annotation.TargetSource != "" {
		if parsed, err := url.Parse(annotation.TargetSource); err == nil {
			sourceDomain = parsed.Host
		}
	}

	authorHandle := annotation.AuthorDID
	profiles := fetchProfilesForDIDs(h.db, []string{annotation.AuthorDID})
	if profile, ok := profiles[annotation.AuthorDID]; ok && profile.Handle != "" {
		authorHandle = "@" + profile.Handle
	}

	pageURL := fmt.Sprintf("%s/at/%s", h.baseURL, url.PathEscape(annotation.URI[5:]))

	var selectorText string
	if annotation.SelectorJSON != nil && *annotation.SelectorJSON != "" {
		var selector struct {
			Exact string `json:"exact"`
		}
		if err := json.Unmarshal([]byte(*annotation.SelectorJSON), &selector); err == nil && selector.Exact != "" {
			selectorText = selector.Exact
			if len(selectorText) > 100 {
				selectorText = selectorText[:97] + "..."
			}
		}
	}

	if selectorText != "" && description != "" {
		description = fmt.Sprintf("\"%s\"\n\n%s", selectorText, description)
	} else if selectorText != "" {
		description = fmt.Sprintf("Highlighted: \"%s\"", selectorText)
	}

	if description == "" {
		description = fmt.Sprintf("An annotation by %s", authorHandle)
		if sourceDomain != "" {
			description += fmt.Sprintf(" on %s", sourceDomain)
		}
	}

	ogImageURL := fmt.Sprintf("%s/og-image?uri=%s", h.baseURL, url.QueryEscape(annotation.URI))

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Margin</title>
    <meta name="description" content="%s">
    
    <!-- Open Graph -->
    <meta property="og:type" content="article">
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:url" content="%s">
    <meta property="og:image" content="%s">
    <meta property="og:image:width" content="1200">
    <meta property="og:image:height" content="630">
    <meta property="og:site_name" content="Margin">
    
    <!-- Twitter Card -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="%s">
    <meta name="twitter:description" content="%s">
    <meta name="twitter:image" content="%s">
    
    <!-- Author -->
    <meta property="article:author" content="%s">
    
    <meta http-equiv="refresh" content="0; url=%s">
</head>
<body>
    <p>Redirecting to <a href="%s">%s</a>...</p>
</body>
</html>`,
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(pageURL),
		html.EscapeString(ogImageURL),
		html.EscapeString(title),
		html.EscapeString(description),
		html.EscapeString(ogImageURL),
		html.EscapeString(authorHandle),
		html.EscapeString(pageURL),
		html.EscapeString(pageURL),
		html.EscapeString(title),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

func (h *OGHandler) serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.staticDir+"/index.html")
}

func (h *OGHandler) HandleOGImage(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri parameter required", http.StatusBadRequest)
		return
	}

	var authorHandle, text, quote, sourceDomain, avatarURL string

	annotation, err := h.db.GetAnnotationByURI(uri)
	if err == nil && annotation != nil {
		authorHandle = annotation.AuthorDID
		profiles := fetchProfilesForDIDs(h.db, []string{annotation.AuthorDID})
		if profile, ok := profiles[annotation.AuthorDID]; ok {
			if profile.Handle != "" {
				authorHandle = "@" + profile.Handle
			}
			if profile.Avatar != "" {
				avatarURL = profile.Avatar
			}
		}

		if annotation.BodyValue != nil {
			text = *annotation.BodyValue
		}

		if annotation.SelectorJSON != nil && *annotation.SelectorJSON != "" {
			var selector struct {
				Exact string `json:"exact"`
			}
			if err := json.Unmarshal([]byte(*annotation.SelectorJSON), &selector); err == nil {
				quote = selector.Exact
			}
		}

		if annotation.TargetSource != "" {
			if parsed, err := url.Parse(annotation.TargetSource); err == nil {
				sourceDomain = parsed.Host
			}
		}
	} else {
		bookmark, err := h.db.GetBookmarkByURI(uri)
		if err == nil && bookmark != nil {
			authorHandle = bookmark.AuthorDID
			profiles := fetchProfilesForDIDs(h.db, []string{bookmark.AuthorDID})
			if profile, ok := profiles[bookmark.AuthorDID]; ok {
				if profile.Handle != "" {
					authorHandle = "@" + profile.Handle
				}
				if profile.Avatar != "" {
					avatarURL = profile.Avatar
				}
			}

			text = "Bookmark"
			if bookmark.Description != nil {
				quote = *bookmark.Description
			}
			if bookmark.Title != nil {
				text = *bookmark.Title
			}

			if bookmark.Source != "" {
				if parsed, err := url.Parse(bookmark.Source); err == nil {
					sourceDomain = parsed.Host
				}
			}
		} else {
			highlight, err := h.db.GetHighlightByURI(uri)
			if err == nil && highlight != nil {
				authorHandle = highlight.AuthorDID
				profiles := fetchProfilesForDIDs(h.db, []string{highlight.AuthorDID})
				if profile, ok := profiles[highlight.AuthorDID]; ok {
					if profile.Handle != "" {
						authorHandle = "@" + profile.Handle
					}
					if profile.Avatar != "" {
						avatarURL = profile.Avatar
					}
				}

				targetTitle := ""
				if highlight.TargetTitle != nil {
					targetTitle = *highlight.TargetTitle
				}

				if highlight.SelectorJSON != nil && *highlight.SelectorJSON != "" {
					var selector struct {
						Exact string `json:"exact"`
					}
					if err := json.Unmarshal([]byte(*highlight.SelectorJSON), &selector); err == nil && selector.Exact != "" {
						quote = selector.Exact
					}
				}

				if highlight.TargetSource != "" {
					if parsed, err := url.Parse(highlight.TargetSource); err == nil {
						sourceDomain = parsed.Host
					}
				}

				img := generateHighlightOGImagePNG(authorHandle, targetTitle, quote, sourceDomain, avatarURL)

				w.Header().Set("Content-Type", "image/png")
				w.Header().Set("Cache-Control", "public, max-age=86400")
				png.Encode(w, img)
				return
			} else {
				collection, err := h.db.GetCollectionByURI(uri)
				if err == nil && collection != nil {
					authorHandle = collection.AuthorDID
					profiles := fetchProfilesForDIDs(h.db, []string{collection.AuthorDID})
					if profile, ok := profiles[collection.AuthorDID]; ok {
						if profile.Handle != "" {
							authorHandle = "@" + profile.Handle
						}
						if profile.Avatar != "" {
							avatarURL = profile.Avatar
						}
					}

					icon := "📁"
					if collection.Icon != nil && *collection.Icon != "" {
						icon = iconToEmoji(*collection.Icon)
					}

					description := ""
					if collection.Description != nil && *collection.Description != "" {
						description = *collection.Description
					}

					img := generateCollectionOGImagePNG(authorHandle, collection.Name, description, icon, avatarURL)

					w.Header().Set("Content-Type", "image/png")
					w.Header().Set("Cache-Control", "public, max-age=86400")
					png.Encode(w, img)
					return
				} else {
					http.Error(w, "Record not found", http.StatusNotFound)
					return
				}
			}
		}
	}

	img := generateOGImagePNG(authorHandle, text, quote, sourceDomain, avatarURL)

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	png.Encode(w, img)
}

func generateOGImagePNG(author, text, quote, source, avatarURL string) image.Image {
	width := 1200
	height := 630
	padding := 100

	bgColor := color.RGBA{9, 9, 11, 255}
	primaryColor := color.RGBA{59, 130, 246, 255}
	primaryLight := color.RGBA{96, 165, 250, 255}
	textPrimary := color.RGBA{250, 250, 250, 255}
	textSecondary := color.RGBA{161, 161, 170, 255}
	borderColor := color.RGBA{63, 63, 70, 255}
	cardBg := color.RGBA{24, 24, 27, 255}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 0, width, 6), &image.Uniform{primaryColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, width-60, height-50), &image.Uniform{cardBg}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, width-60, 51), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, height-51, width-60, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, 61, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(width-61, 50, width-60, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)

	avatarSize := 64
	avatarX := padding
	avatarY := padding

	avatarImg := fetchAvatarImage(avatarURL)
	if avatarImg != nil {
		drawCircularAvatar(img, avatarImg, avatarX, avatarY, avatarSize)
	} else {
		drawDefaultAvatar(img, author, avatarX, avatarY, avatarSize, primaryColor)
	}
	drawText(img, author, avatarX+avatarSize+24, avatarY+42, textSecondary, 28, false)

	contentWidth := width - (padding * 2)
	yPos := 220

	if text != "" {
		textLen := len(text)
		textSize := 32.0
		textLineHeight := 42
		maxTextLines := 5

		if textLen > 200 {
			textSize = 28.0
			textLineHeight = 36
			maxTextLines = 6
		}

		lines := wrapTextToWidth(text, contentWidth, int(textSize))
		numLines := min(len(lines), maxTextLines)

		for i := 0; i < numLines; i++ {
			line := lines[i]
			if i == numLines-1 && len(lines) > numLines {
				line += "..."
			}
			drawText(img, line, padding, yPos+(i*textLineHeight), textPrimary, textSize, false)
		}
		yPos += (numLines * textLineHeight) + 40
	}

	if quote != "" {
		quoteLen := len(quote)
		quoteSize := 24.0
		quoteLineHeight := 32
		maxQuoteLines := 3

		if quoteLen > 150 {
			quoteSize = 20.0
			quoteLineHeight = 28
			maxQuoteLines = 4
		}

		lines := wrapTextToWidth(quote, contentWidth-30, int(quoteSize))
		numLines := min(len(lines), maxQuoteLines)
		barHeight := numLines * quoteLineHeight

		draw.Draw(img, image.Rect(padding, yPos, padding+6, yPos+barHeight), &image.Uniform{primaryLight}, image.Point{}, draw.Src)

		for i := 0; i < numLines; i++ {
			line := lines[i]
			if i == numLines-1 && len(lines) > numLines {
				line += "..."
			}
			drawText(img, line, padding+24, yPos+24+(i*quoteLineHeight), textSecondary, quoteSize, true)
		}
		yPos += barHeight + 40
	}

	draw.Draw(img, image.Rect(padding, yPos, width-padding, yPos+1), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	yPos += 40
	drawText(img, source, padding, yPos+32, textSecondary, 24, false)

	if logoImage != nil {
		logoSize := 32
		drawScaledImage(img, logoImage, width-padding-logoSize, height-90, logoSize, logoSize)
	}

	return img
}

func drawScaledImage(dst *image.RGBA, src image.Image, x, y, w, h int) {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			srcX := bounds.Min.X + (dx * srcW / w)
			srcY := bounds.Min.Y + (dy * srcH / h)
			c := src.At(srcX, srcY)
			_, _, _, a := c.RGBA()
			if a > 0 {
				dst.Set(x+dx, y+dy, c)
			}
		}
	}
}

func fetchAvatarImage(avatarURL string) image.Image {
	if avatarURL == "" {
		return nil
	}

	resp, err := http.Get(avatarURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil
	}

	return img
}

func drawCircularAvatar(dst *image.RGBA, src image.Image, x, y, size int) {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	centerX := size / 2
	centerY := size / 2
	radius := size / 2

	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			distX := dx - centerX
			distY := dy - centerY
			if distX*distX+distY*distY <= radius*radius {
				srcX := bounds.Min.X + (dx * srcW / size)
				srcY := bounds.Min.Y + (dy * srcH / size)
				dst.Set(x+dx, y+dy, src.At(srcX, srcY))
			}
		}
	}
}

func drawDefaultAvatar(dst *image.RGBA, author string, x, y, size int, accentColor color.RGBA) {
	centerX := size / 2
	centerY := size / 2
	radius := size / 2

	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			distX := dx - centerX
			distY := dy - centerY
			if distX*distX+distY*distY <= radius*radius {
				dst.Set(x+dx, y+dy, accentColor)
			}
		}
	}

	initial := "?"
	if len(author) > 1 {
		if author[0] == '@' && len(author) > 1 {
			initial = strings.ToUpper(string(author[1]))
		} else {
			initial = strings.ToUpper(string(author[0]))
		}
	}
	drawText(dst, initial, x+size/2-10, y+size/2+12, color.RGBA{255, 255, 255, 255}, 32, true)
}

func wrapTextToWidth(text string, maxWidth int, fontSize int) []string {
	words := strings.Fields(text)
	var lines []string
	var currentLine string

	charWidth := fontSize * 6 / 10

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine)*charWidth > maxWidth && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

func generateCollectionOGImagePNG(author, collectionName, description, icon, avatarURL string) image.Image {
	width := 1200
	height := 630
	padding := 120

	bgColor := color.RGBA{9, 9, 11, 255}
	primaryColor := color.RGBA{59, 130, 246, 255}
	textPrimary := color.RGBA{250, 250, 250, 255}
	textSecondary := color.RGBA{161, 161, 170, 255}
	textTertiary := color.RGBA{113, 113, 122, 255}
	borderColor := color.RGBA{63, 63, 70, 255}
	cardBg := color.RGBA{24, 24, 27, 255}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 0, width, 6), &image.Uniform{primaryColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, width-60, height-50), &image.Uniform{cardBg}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, width-60, 51), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, height-51, width-60, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, 61, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(width-61, 50, width-60, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)

	iconY := 120
	var iconWidth int
	if icon != "" {
		// Render emoji using the fallback font directly
		drawText(img, icon, padding, iconY+70, textPrimary, 80, true)
		iconWidth = 100
	}

	drawText(img, collectionName, padding+iconWidth, iconY+65, textPrimary, 64, true)

	yPos := 280
	contentWidth := width - (padding * 2)

	if description != "" {
		if len(description) > 200 {
			description = description[:197] + "..."
		}
		lines := wrapTextToWidth(description, contentWidth, 32)
		for i, line := range lines {
			if i >= 4 {
				break
			}
			drawText(img, line, padding, yPos+(i*42), textSecondary, 32, false)
		}
	} else {
		drawText(img, "A collection on Margin", padding, yPos, textTertiary, 32, false)
	}

	yPos = 480
	draw.Draw(img, image.Rect(padding, yPos, width-padding, yPos+1), &image.Uniform{borderColor}, image.Point{}, draw.Src)

	avatarSize := 64
	avatarX := padding
	avatarY := yPos + 40

	avatarImg := fetchAvatarImage(avatarURL)
	if avatarImg != nil {
		drawCircularAvatar(img, avatarImg, avatarX, avatarY, avatarSize)
	} else {
		drawDefaultAvatar(img, author, avatarX, avatarY, avatarSize, primaryColor)
	}

	handleX := avatarX + avatarSize + 24
	drawText(img, author, handleX, avatarY+42, textTertiary, 28, false)

	if logoImage != nil {
		logoSize := 32
		drawScaledImage(img, logoImage, width-padding-logoSize, height-90, logoSize, logoSize)
	}

	return img
}

func generateHighlightOGImagePNG(author, pageTitle, quote, source, avatarURL string) image.Image {
	width := 1200
	height := 630
	padding := 100

	bgColor := color.RGBA{9, 9, 11, 255}
	primaryColor := color.RGBA{59, 130, 246, 255}
	primaryLight := color.RGBA{96, 165, 250, 255}
	textPrimary := color.RGBA{250, 250, 250, 255}
	textSecondary := color.RGBA{161, 161, 170, 255}
	borderColor := color.RGBA{63, 63, 70, 255}
	cardBg := color.RGBA{24, 24, 27, 255}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 0, width, 6), &image.Uniform{primaryColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, width-60, height-50), &image.Uniform{cardBg}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, width-60, 51), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, height-51, width-60, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(60, 50, 61, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(width-61, 50, width-60, height-50), &image.Uniform{borderColor}, image.Point{}, draw.Src)

	avatarSize := 64
	avatarX := padding
	avatarY := padding

	avatarImg := fetchAvatarImage(avatarURL)
	if avatarImg != nil {
		drawCircularAvatar(img, avatarImg, avatarX, avatarY, avatarSize)
	} else {
		drawDefaultAvatar(img, author, avatarX, avatarY, avatarSize, primaryColor)
	}
	drawText(img, author, avatarX+avatarSize+24, avatarY+42, textSecondary, 28, false)

	contentWidth := width - (padding * 2)
	yPos := 220
	if quote != "" {
		quoteLen := len(quote)
		fontSize := 42.0
		lineHeight := 56
		maxLines := 4

		if quoteLen > 200 {
			fontSize = 32.0
			lineHeight = 44
			maxLines = 6
		} else if quoteLen > 100 {
			fontSize = 36.0
			lineHeight = 48
			maxLines = 5
		}

		lines := wrapTextToWidth(quote, contentWidth-40, int(fontSize))
		numLines := min(len(lines), maxLines)
		barHeight := numLines * lineHeight

		draw.Draw(img, image.Rect(padding, yPos, padding+8, yPos+barHeight), &image.Uniform{primaryLight}, image.Point{}, draw.Src)

		for i := 0; i < numLines; i++ {
			line := lines[i]
			if i == numLines-1 && len(lines) > numLines {
				line += "..."
			}
			drawText(img, line, padding+40, yPos+42+(i*lineHeight), textPrimary, fontSize, false)
		}
		yPos += barHeight + 40
	}

	draw.Draw(img, image.Rect(padding, yPos, width-padding, yPos+1), &image.Uniform{borderColor}, image.Point{}, draw.Src)
	yPos += 40

	if pageTitle != "" {
		if len(pageTitle) > 60 {
			pageTitle = pageTitle[:57] + "..."
		}
		drawText(img, pageTitle, padding, yPos+32, textSecondary, 32, true)
	}

	if source != "" {
		drawText(img, source, padding, yPos+80, textSecondary, 24, false)
	}

	if logoImage != nil {
		logoSize := 32
		drawScaledImage(img, logoImage, width-padding-logoSize, height-90, logoSize, logoSize)
	}

	return img
}
