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
<title>{{.Title}}</title>
<style>
  :root { --bg:#fff; --fg:#1a1a1a; --muted:#666; --border:#e5e5e5; --hover:#f0f0f0; --active:#0070f3; --code-bg:#f5f5f5; }
  @media (prefers-color-scheme: dark) {
    :root { --bg:#1a1a1a; --fg:#e5e5e5; --muted:#999; --border:#333; --hover:#2a2a2a; --active:#3b82f6; --code-bg:#2a2a2a; }
  }
  *{box-sizing:border-box}
  body{font:16px/1.6 -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;margin:0;display:flex;background:var(--bg);color:var(--fg)}
  nav{width:280px;padding:1rem;height:100vh;overflow-y:auto;border-right:1px solid var(--border);flex-shrink:0}
  nav a{display:block;padding:.25rem .5rem;color:var(--fg);text-decoration:none;border-radius:4px;font-size:14px}
  nav a:hover{background:var(--hover)}
  nav a.active{background:var(--active);color:#fff}
  nav details{margin:.25rem 0}
  nav summary{cursor:pointer;padding:.25rem .5rem;font-size:14px;font-weight:600;border-radius:4px}
  nav summary:hover{background:var(--hover)}
  nav details > *:not(summary){margin-left:.75rem;padding-left:.5rem;border-left:1px solid var(--border)}
  main{padding:2rem 3rem;overflow-y:auto;height:100vh;flex:1}
  main > article{max-width:860px;margin:0 auto}
  main h1,main h2,main h3,main h4{margin-top:2rem;line-height:1.3}
  main h1:first-child{margin-top:0}
  main pre{background:var(--code-bg);padding:1rem;border-radius:6px;overflow-x:auto;font-size:14px}
  main code{background:var(--code-bg);padding:.15rem .35rem;border-radius:3px;font-size:.9em}
  main pre code{background:none;padding:0}
  main table{border-collapse:collapse;margin:1rem 0}
  main td,main th{border:1px solid var(--border);padding:.5rem .75rem}
  main blockquote{border-left:3px solid var(--border);margin:0;padding:.5rem 1rem;color:var(--muted)}
  main a{color:var(--active)}
  main img{max-width:100%}
  @media (max-width: 768px){
    body{flex-direction:column}
    nav{width:100%;height:auto;max-height:40vh;border-right:none;border-bottom:1px solid var(--border)}
    main{height:auto;padding:1.5rem}
  }
</style>
</head>
<body>
<nav>{{template "nav" .Nav}}</nav>
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
	p := parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs)
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
