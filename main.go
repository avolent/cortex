package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const (
	debounceWindow  = 120 * time.Millisecond
	ssePingInterval = 30 * time.Second
	readHeaderLimit = 10 * time.Second
	idleConnLimit   = 60 * time.Second
)

var (
	rootDir   = flag.String("dir", ".", "repository root to serve")
	addr      = flag.String("addr", ":8090", "listen address")
	exportDir = flag.String("export", "", "if set, render the wiki to static HTML in this directory and exit")
)

var skipDirs = map[string]bool{
	".git": true, "node_modules": true, "vendor": true, "tmp": true,
	".claude": true, "dist": true, "build": true, ".next": true, ".cache": true,
}

var staticExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".svg": true, ".webp": true, ".ico": true,
}

type navItem struct {
	Title    string
	Path     string
	Active   bool
	IsDir    bool
	Children []navItem
}

type pageData struct {
	Title   string
	Content template.HTML
	Nav     []navItem
	Reload  bool
}

const tmplStr = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml;utf8,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3E%3Ccircle cx='8' cy='8' r='6' fill='none' stroke='%23888' stroke-width='2'/%3E%3C/svg%3E">
<title>{{.Title}}</title>
<style>
  :root {
    --bg: #fdfdfc; --fg: #1a1a1a; --muted: #5a5a5a;
    --border: #e3e3e0; --code-bg: #f4f3f0; --hover: rgba(0,0,0,0.05);
    --link: #0a4ea3;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #1a1a1a; --fg: #e6e6e6; --muted: #9a9a9a;
      --border: #2c2c2c; --code-bg: #232323; --hover: rgba(255,255,255,0.06);
      --link: #6db0ff;
    }
  }
  * { box-sizing: border-box; }
  body {
    margin: 0;
    color: var(--fg); background: var(--bg);
    font: 18px/1.7 XCharter, Charter, "Bitstream Charter", "Sitka Text", Cambria, Georgia, serif;
    padding-left: 280px;
  }
  nav.sidebar {
    position: fixed; top: 0; left: 0;
    width: 280px; height: 100vh;
    padding: 1.5rem 1.25rem; overflow-y: auto;
    border-right: 1px solid var(--border);
    font-size: 0.95rem; line-height: 1.5;
  }
  nav.sidebar a, nav.sidebar summary {
    display: block; padding: .2rem .5rem;
    color: inherit; text-decoration: none;
    border-radius: 3px;
  }
  nav.sidebar a:hover, nav.sidebar summary:hover { background: var(--hover); }
  nav.sidebar a.active { font-weight: 700; }
  nav.sidebar summary { cursor: pointer; font-weight: 600; }
  nav.sidebar details { margin: .25rem 0; }
  nav.sidebar details > *:not(summary) {
    margin-left: .6rem; padding-left: .6rem;
    border-left: 1px solid var(--border);
  }
  main {
    max-width: clamp(65ch, calc(100vw - 280px - 6rem), 90ch);
    margin: 0 auto; padding: 3rem 2rem 4rem;
  }
  article h1, article h2, article h3, article h4, article h5, article h6 { line-height: 1.25; margin: 2.5rem 0 1rem; }
  article h1 { font-size: 2rem; margin-top: 0; text-align: center; }
  article h2 { font-size: 1.5rem; }
  article h3 { font-size: 1.2rem; }
  article h4 { font-size: 1.05rem; }
  article h5 { font-size: 1rem; font-weight: 700; }
  article h6 { font-size: 1rem; font-weight: 700; color: var(--muted); }
  article p { margin: 0 0 1rem; }
  article a { color: var(--link); }
  article ul, article ol { margin: 0 0 1rem; padding-left: 1.5rem; }
  article li { margin: .25rem 0; }
  article li > ul, article li > ol { margin-bottom: 0; }
  article dl { margin: 0 0 1rem; }
  article dt { font-weight: 700; margin-top: .5rem; }
  article dd { margin: 0 0 .5rem 1.5rem; }
  article code, article pre {
    font-family: ui-monospace, SFMono-Regular, "SF Mono", "Cascadia Code", "Roboto Mono", Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
    font-feature-settings: "liga" 0, "calt" 0;
  }
  article code {
    font-size: .9em;
    background: var(--code-bg); padding: .15em .4em; border-radius: 0.25em;
    vertical-align: -0.05em;
    overflow-wrap: anywhere;
  }
  article pre {
    border-left: 3px solid var(--border);
    padding: .5rem 1rem; margin: 1rem 0;
    overflow-x: auto;
    font-size: .9em; line-height: 1.5;
  }
  article pre code {
    background: none; padding: 0; font-size: 1em;
    vertical-align: baseline; overflow-wrap: normal;
  }
  article blockquote {
    border-left: 3px solid var(--border);
    margin: 1rem 0; padding: .25rem 1rem;
    color: var(--muted);
  }
  article table { border-collapse: collapse; margin: 1rem 0; }
  article th, article td { border: 1px solid var(--border); padding: .5rem .75rem; }
  article th { background: var(--code-bg); }
  article img { max-width: 100%; height: auto; }
  article hr { border: 0; border-top: 1px solid var(--border); margin: 2rem 0; }
  article sup.footnote-ref { font-size: .75em; }
  article sup.footnote-ref a { text-decoration: none; }
  article .footnotes { margin-top: 3rem; font-size: .9em; color: var(--muted); }
  article .footnotes hr { margin-bottom: .5rem; }
  article .footnotes li { margin: .35rem 0; }
  article .footnotes li p { margin: 0; }
  @media (max-width: 768px) {
    body { padding-left: 0; }
    nav.sidebar {
      position: static; width: 100%; height: auto;
      max-height: 40vh; border-right: none;
      border-bottom: 1px solid var(--border);
    }
    main { padding: 2rem 1.25rem 3rem; }
    article h1 { font-size: 1.75rem; }
  }
</style>
</head>
<body>
<nav class="sidebar">{{template "nav" .Nav}}</nav>
<main><article>{{.Content}}</article></main>
{{if .Reload}}<script>(()=>{const es=new EventSource('/_cortex/events');es.onmessage=()=>location.reload();es.onerror=()=>{};})();</script>{{end}}
</body>
</html>
{{define "nav"}}{{range .}}{{if .IsDir}}<details open><summary>{{.Title}}</summary>{{template "nav" .Children}}</details>{{else}}<a href="{{.Path}}"{{if .Active}} class="active"{{end}}>{{.Title}}</a>{{end}}{{end}}{{end}}
`

var tmpl = template.Must(template.New("page").Parse(tmplStr))

// isSkipped reports whether any path segment is in skipDirs or starts with a dot.
func isSkipped(rel string) bool {
	rel = filepath.ToSlash(rel)
	for _, part := range strings.Split(rel, "/") {
		if part == "" || part == "." {
			continue
		}
		if skipDirs[part] || strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// collectMarkdown walks the root and returns relative paths (slash-separated) of every .md file.
func collectMarkdown() []string {
	var files []string
	_ = filepath.WalkDir(*rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(*rootDir, p)
		if relErr != nil {
			return nil
		}
		if d.IsDir() {
			if rel != "." && (skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(p), ".md") {
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	return files
}

// buildTree groups files by directory under prefix. activeURL is "/foo/bar" form.
func buildTree(prefix string, files []string, activeURL string) []navItem {
	subdirs := map[string][]string{}
	var here []string
	for _, f := range files {
		rest := strings.TrimPrefix(f, prefix)
		rest = strings.TrimPrefix(rest, "/")
		if i := strings.Index(rest, "/"); i >= 0 {
			subdirs[rest[:i]] = append(subdirs[rest[:i]], f)
		} else {
			here = append(here, f)
		}
	}

	// Loose files: README first (case-insensitive), then alphabetical (case-insensitive).
	sort.Slice(here, func(i, j int) bool {
		bi := strings.ToLower(filepath.Base(here[i]))
		bj := strings.ToLower(filepath.Base(here[j]))
		ri := bi == "readme.md"
		rj := bj == "readme.md"
		if ri != rj {
			return ri
		}
		return bi < bj
	})

	items := make([]navItem, 0, len(here)+len(subdirs))
	for _, f := range here {
		urlPath := urlForFile(f)
		items = append(items, navItem{
			Title:  titleForFile(f),
			Path:   urlPath,
			Active: urlPath == activeURL,
		})
	}

	// Subdirectories alphabetical.
	dirNames := make([]string, 0, len(subdirs))
	for name := range subdirs {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)
	for _, name := range dirNames {
		newPrefix := name
		if prefix != "" {
			newPrefix = prefix + "/" + name
		}
		items = append(items, navItem{
			Title:    name,
			IsDir:    true,
			Children: buildTree(newPrefix, subdirs[name], activeURL),
		})
	}
	return items
}

// resolveMarkdown returns the on-disk relative path of the .md file backing the URL, or "" if none.
// Lookup is case-insensitive on the .md extension and on the file stem, so /TODO matches TODO.MD.
func resolveMarkdown(urlPath string) string {
	urlPath = strings.TrimPrefix(urlPath, "/")
	urlPath = strings.TrimSuffix(urlPath, "/")
	if ext := filepath.Ext(urlPath); strings.EqualFold(ext, ".md") {
		urlPath = urlPath[:len(urlPath)-len(ext)]
	}

	var candidates []string
	if urlPath == "" {
		candidates = []string{"README.md", "readme.md"}
	} else {
		candidates = []string{
			urlPath + ".md",
			filepath.Join(urlPath, "README.md"),
			filepath.Join(urlPath, "readme.md"),
		}
	}
	for _, c := range candidates {
		if isSkipped(c) {
			continue
		}
		full := filepath.Join(*rootDir, c)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			return c
		}
	}

	// Fallback: case-insensitive directory scan for the stem (e.g. URL /TODO -> TODO.MD).
	if urlPath != "" && !isSkipped(urlPath) {
		dir := filepath.Dir(urlPath)
		if dir == "." {
			dir = ""
		}
		stem := filepath.Base(urlPath)
		searchDir := *rootDir
		if dir != "" {
			searchDir = filepath.Join(*rootDir, dir)
		}
		if entries, err := os.ReadDir(searchDir); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				name := e.Name()
				ext := filepath.Ext(name)
				if !strings.EqualFold(ext, ".md") {
					continue
				}
				if strings.EqualFold(name[:len(name)-len(ext)], stem) {
					rel := name
					if dir != "" {
						rel = filepath.Join(dir, name)
					}
					return rel
				}
			}
		}
	}
	return ""
}

// hrefAttrRe matches href and src attributes with double-quoted values.
// gomarkdown emits double quotes, which is what we target.
var hrefAttrRe = regexp.MustCompile(`(href|src)="([^"]*)"`)

// rewriteMdLinks rewrites internal .md hrefs to directory-style URLs so the
// same HTML works under both the live server (which accepts /foo and /foo.md)
// and a static host like GitHub Pages serving /foo/index.html as /foo/.
//
//	"other.md"          -> "other/"
//	"dir/page.md"       -> "dir/page/"
//	"README.md"         -> "./"
//	"dir/README.md"     -> "dir/"
//	"../foo.md#anchor"  -> "../foo/#anchor"
//
// External (http/https/protocol-relative/mailto) and non-.md hrefs are left
// untouched.
func rewriteMdLinks(in []byte) []byte {
	return hrefAttrRe.ReplaceAllFunc(in, func(match []byte) []byte {
		sub := hrefAttrRe.FindSubmatch(match)
		attr, url := string(sub[1]), string(sub[2])

		if url == "" ||
			strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") ||
			strings.HasPrefix(url, "//") || strings.HasPrefix(url, "mailto:") ||
			strings.HasPrefix(url, "#") {
			return match
		}

		urlPath, suffix := url, ""
		if i := strings.IndexAny(url, "#?"); i >= 0 {
			urlPath, suffix = url[:i], url[i:]
		}
		if !strings.EqualFold(filepath.Ext(urlPath), ".md") {
			return match
		}

		stem := urlPath[:len(urlPath)-len(".md")]
		var out string
		if strings.EqualFold(filepath.Base(stem), "README") {
			dir := path.Dir(stem)
			if dir == "." || dir == "" {
				out = "./"
			} else {
				out = dir + "/"
			}
		} else {
			out = stem + "/"
		}
		return []byte(fmt.Sprintf(`%s="%s%s"`, attr, out, suffix))
	})
}

func renderMarkdown(md []byte) template.HTML {
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.Footnotes)
	r := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})
	return template.HTML(rewriteMdLinks(markdown.Render(p.Parse(md), r)))
}

// titleForFile returns the page title for an on-disk markdown path.
// "foo/bar.md" -> "bar", "foo/README.md" -> "README".
func titleForFile(rel string) string {
	stem := strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
	return path.Base(stem)
}

func handler(w http.ResponseWriter, r *http.Request) {
	cleaned := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
	if strings.Contains(cleaned, "..") {
		http.NotFound(w, r)
		return
	}

	if rel := resolveMarkdown(cleaned); rel != "" {
		md, err := os.ReadFile(filepath.Join(*rootDir, rel))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, pageData{
			Title:   titleForFile(rel),
			Content: renderMarkdown(md),
			Nav:     buildTree("", collectMarkdown(), urlForFile(rel)),
			Reload:  true,
		}); err != nil {
			log.Printf("render %s: %v", r.URL.Path, err)
		}
		return
	}

	// Static asset fallback (images embedded in markdown, etc.).
	rel := strings.TrimPrefix(cleaned, "/")
	ext := strings.ToLower(filepath.Ext(rel))
	if rel != "" && staticExts[ext] && !isSkipped(rel) {
		http.ServeFile(w, r, filepath.Join(*rootDir, rel))
		return
	}

	http.NotFound(w, r)
}

// reloader fans out fsnotify events to connected SSE clients.
type reloader struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
}

func newReloader() *reloader { return &reloader{subs: map[chan struct{}]struct{}{}} }

func (r *reloader) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	r.mu.Lock()
	r.subs[ch] = struct{}{}
	r.mu.Unlock()
	return ch
}

func (r *reloader) unsubscribe(ch chan struct{}) {
	r.mu.Lock()
	delete(r.subs, ch)
	r.mu.Unlock()
}

func (r *reloader) broadcast() {
	r.mu.Lock()
	for ch := range r.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	r.mu.Unlock()
}

// watchTree recursively watches root for changes to .md files and static assets,
// debouncing rapid event bursts (typical of editors saving) into a single broadcast.
func watchTree(ctx context.Context, root string, r *reloader) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("watch disabled: %v", err)
		return
	}
	defer w.Close()

	addDir := func(p string) {
		if err := w.Add(p); err != nil {
			log.Printf("watch %s: %v", p, err)
		}
	}

	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if rel != "." && (skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
			return filepath.SkipDir
		}
		addDir(p)
		return nil
	})

	var (
		debounceMu sync.Mutex
		timer      *time.Timer
	)
	schedule := func() {
		debounceMu.Lock()
		defer debounceMu.Unlock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceWindow, r.broadcast)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if ev.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(ev.Name); err == nil && info.IsDir() {
					base := filepath.Base(ev.Name)
					if !skipDirs[base] && !strings.HasPrefix(base, ".") {
						addDir(ev.Name)
					}
				}
			}
			ext := strings.ToLower(filepath.Ext(ev.Name))
			if ext != ".md" && !staticExts[ext] {
				continue
			}
			schedule()
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Printf("watcher: %v", err)
		}
	}
}

func sseHandler(r *reloader) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		fmt.Fprint(w, ": connected\n\n")
		flusher.Flush()

		ch := r.subscribe()
		defer r.unsubscribe(ch)
		ping := time.NewTicker(ssePingInterval)
		defer ping.Stop()

		for {
			select {
			case <-req.Context().Done():
				return
			case <-ch:
				fmt.Fprint(w, "data: reload\n\n")
				flusher.Flush()
			case <-ping.C:
				fmt.Fprint(w, ": ping\n\n")
				flusher.Flush()
			}
		}
	}
}

// urlForFile maps an on-disk markdown path (slash-separated) to its canonical URL.
// "README.md" -> "/", "foo.md" -> "/foo", "foo/README.md" -> "/foo".
func urlForFile(rel string) string {
	rel = filepath.ToSlash(rel)
	stem := strings.TrimSuffix(rel, filepath.Ext(rel))
	if strings.EqualFold(filepath.Base(stem), "README") {
		dir := filepath.ToSlash(filepath.Dir(stem))
		if dir == "." {
			return "/"
		}
		return "/" + dir
	}
	return "/" + stem
}

// outPathForURL maps a URL to a static-HTML output path under outDir.
// "/" -> outDir/index.html, "/foo" -> outDir/foo/index.html.
func outPathForURL(outDir, url string) string {
	if url == "/" {
		return filepath.Join(outDir, "index.html")
	}
	return filepath.Join(outDir, filepath.FromSlash(strings.TrimPrefix(url, "/")), "index.html")
}

// writeStaticPage renders a page to a single HTML file under outPath.
func writeStaticPage(outPath string, page pageData) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()
	return tmpl.Execute(out, page)
}

// copyAsset copies src to dst, creating any missing parent directories.
func copyAsset(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

// exportSite renders every .md page to static HTML under outDir, then copies
// referenced static assets (images) into the same tree.
func exportSite(outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	files := collectMarkdown()
	seen := map[string]bool{}
	for _, f := range files {
		url := urlForFile(f)
		if seen[url] {
			log.Printf("skip %s: %q already produced by another file", f, url)
			continue
		}
		seen[url] = true

		rel := resolveMarkdown(url)
		if rel == "" {
			continue
		}
		md, err := os.ReadFile(filepath.Join(*rootDir, rel))
		if err != nil {
			return fmt.Errorf("read %s: %w", rel, err)
		}
		page := pageData{
			Title:   titleForFile(rel),
			Content: renderMarkdown(md),
			Nav:     buildTree("", files, url),
		}
		outPath := outPathForURL(outDir, url)
		if err := writeStaticPage(outPath, page); err != nil {
			return fmt.Errorf("render %s: %w", url, err)
		}
		log.Printf("wrote %s", outPath)
	}

	// Copy embedded static assets (images) into the same relative paths.
	return filepath.WalkDir(*rootDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(*rootDir, p)
		if d.IsDir() {
			if rel != "." && (skipDirs[d.Name()] || strings.HasPrefix(d.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if !staticExts[ext] || isSkipped(rel) {
			return nil
		}
		dst := filepath.Join(outDir, rel)
		if err := copyAsset(p, dst); err != nil {
			return fmt.Errorf("copy %s: %w", rel, err)
		}
		log.Printf("copied %s", dst)
		return nil
	})
}

func main() {
	flag.Parse()
	abs, err := filepath.Abs(*rootDir)
	if err != nil {
		log.Fatalf("resolve dir: %v", err)
	}
	*rootDir = abs
	if info, err := os.Stat(*rootDir); err != nil || !info.IsDir() {
		log.Fatalf("dir not found: %s", *rootDir)
	}

	if *exportDir != "" {
		if err := exportSite(*exportDir); err != nil {
			log.Fatalf("export: %v", err)
		}
		return
	}

	rl := newReloader()
	go watchTree(context.Background(), *rootDir, rl)

	mux := http.NewServeMux()
	mux.HandleFunc("/_cortex/events", sseHandler(rl))
	mux.HandleFunc("/", handler)

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderLimit,
		IdleTimeout:       idleConnLimit,
		// WriteTimeout intentionally unset: SSE connections are long-lived.
	}
	log.Printf("cortex: serving %s on http://localhost%s", *rootDir, *addr)
	log.Fatal(srv.ListenAndServe())
}
