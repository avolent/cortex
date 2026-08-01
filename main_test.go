package main

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withRoot runs fn with *rootDir temporarily set, then restores it.
// All filesystem-touching tests share this helper since *rootDir is global.
// The file index cache is invalidated on entry and exit so each test sees
// a fresh view of its fixture directory. Mirrors the EvalSymlinks step
// main() applies so safeResolve comparisons are consistent on platforms
// (e.g. macOS) where /var resolves to /private/var.
func withRoot(t *testing.T, dir string, fn func()) {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(dir)
	if err != nil {
		resolved = dir
	}
	old := *rootDir
	*rootDir = resolved
	index.invalidate()
	defer func() {
		*rootDir = old
		index.invalidate()
	}()
	fn()
}

// withIndexNames runs fn with *indexNames temporarily set, then restores it.
func withIndexNames(t *testing.T, value string, fn func()) {
	t.Helper()
	old := *indexNames
	*indexNames = value
	defer func() { *indexNames = old }()
	fn()
}

// writeTree creates files (with parent dirs) under root for fixture setup.
func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		full := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUrlForFile(t *testing.T) {
	cases := map[string]string{
		"README.md":       "/",
		"readme.md":       "/",
		"foo.md":          "/foo",
		"foo/bar.md":      "/foo/bar",
		"foo/README.md":   "/foo",
		"foo/readme.md":   "/foo",
		"a/b/c/README.md": "/a/b/c",
	}
	for in, want := range cases {
		if got := urlForFile(in); got != want {
			t.Errorf("urlForFile(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOutPathForURL(t *testing.T) {
	cases := []struct{ url, want string }{
		{"/", filepath.Join("out", "index.html")},
		{"/foo", filepath.Join("out", "foo", "index.html")},
		{"/foo/bar", filepath.Join("out", "foo", "bar", "index.html")},
	}
	for _, tc := range cases {
		if got := outPathForURL("out", tc.url); got != tc.want {
			t.Errorf("outPathForURL(%q) = %q, want %q", tc.url, got, tc.want)
		}
	}
}

func TestIsSkipped(t *testing.T) {
	cases := map[string]bool{
		"foo":           false,
		"foo/bar":       false,
		".":             false,
		".git":          true,
		"node_modules":  true,
		"foo/.hidden":   true,
		"foo/.git/x.md": true,
		".dotfile":      true,
	}
	for in, want := range cases {
		if got := isSkipped(in); got != want {
			t.Errorf("isSkipped(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestTitleForFile(t *testing.T) {
	cases := map[string]string{
		"README.md":      "README",
		"foo.md":         "foo",
		"foo/bar.md":     "bar",
		"foo/README.md":  "README",
		"a/b/c/notes.md": "notes",
	}
	for in, want := range cases {
		if got := titleForFile(in); got != want {
			t.Errorf("titleForFile(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRewriteMdLinks(t *testing.T) {
	cases := []struct{ in, want string }{
		// Internal .md links get rewritten.
		{`<a href="other.md">x</a>`, `<a href="other/">x</a>`},
		{`<a href="dir/page.md">x</a>`, `<a href="dir/page/">x</a>`},
		{`<a href="../foo.md">x</a>`, `<a href="../foo/">x</a>`},

		// README.md collapses to its parent dir.
		{`<a href="README.md">x</a>`, `<a href="./">x</a>`},
		{`<a href="dir/README.md">x</a>`, `<a href="dir/">x</a>`},
		{`<a href="../README.md">x</a>`, `<a href="../">x</a>`},

		// Fragments and queries preserved.
		{`<a href="other.md#section">x</a>`, `<a href="other/#section">x</a>`},
		{`<a href="other.md?q=1">x</a>`, `<a href="other/?q=1">x</a>`},
		{`<a href="dir/README.md#x">y</a>`, `<a href="dir/#x">y</a>`},

		// Case-insensitive on .md extension.
		{`<a href="other.MD">x</a>`, `<a href="other/">x</a>`},

		// External URLs left alone.
		{`<a href="https://example.com/x.md">x</a>`, `<a href="https://example.com/x.md">x</a>`},
		{`<a href="http://example.com/x.md">x</a>`, `<a href="http://example.com/x.md">x</a>`},
		{`<a href="//example.com/x.md">x</a>`, `<a href="//example.com/x.md">x</a>`},
		{`<a href="mailto:x@example.com">x</a>`, `<a href="mailto:x@example.com">x</a>`},

		// Anchor-only links left alone.
		{`<a href="#section">x</a>`, `<a href="#section">x</a>`},

		// Non-.md hrefs left alone.
		{`<a href="other.html">x</a>`, `<a href="other.html">x</a>`},
		{`<a href="other">x</a>`, `<a href="other">x</a>`},

		// Image src untouched (rarely .md, but covered for completeness).
		{`<img src="img.png">`, `<img src="img.png">`},
	}
	for _, tc := range cases {
		got := string(rewriteMdLinks([]byte(tc.in)))
		if got != tc.want {
			t.Errorf("rewriteMdLinks(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSymlinkEscapesRootRejected(t *testing.T) {
	if _, err := os.Stat("/etc/hosts"); err != nil {
		t.Skip("no /etc/hosts on this platform")
	}
	tmp := t.TempDir()
	writeTree(t, tmp, map[string]string{
		"README.md": "# Home\n",
	})
	// Symlink whose target is outside the served root.
	if err := os.Symlink("/etc/hosts", filepath.Join(tmp, "escape.md")); err != nil {
		t.Fatal(err)
	}

	withRoot(t, tmp, func() {
		// collectMarkdown must not include the escaping symlink.
		for _, f := range collectMarkdown() {
			if f == "escape.md" {
				t.Errorf("collectMarkdown included symlink escaping the root: %q", f)
			}
		}
		// The handler must 404 the URL even if the index were stale.
		req := httptest.NewRequest("GET", "/escape", nil)
		rec := httptest.NewRecorder()
		handler(rec, req)
		if rec.Code != 404 {
			t.Errorf("escape symlink: status = %d, want 404", rec.Code)
		}
	})
}

func TestSymlinkInsideRootAccepted(t *testing.T) {
	tmp := t.TempDir()
	writeTree(t, tmp, map[string]string{
		"real.md": "# Real\n",
	})
	if err := os.Symlink("real.md", filepath.Join(tmp, "alias.md")); err != nil {
		t.Fatal(err)
	}

	withRoot(t, tmp, func() {
		req := httptest.NewRequest("GET", "/alias", nil)
		rec := httptest.NewRecorder()
		handler(rec, req)
		if rec.Code != 200 {
			t.Errorf("in-root symlink: status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Real") {
			t.Errorf("expected rendered body of symlink target")
		}
	})
}

func TestReadMarkdownSizeLimit(t *testing.T) {
	tmp := t.TempDir()
	huge := strings.Repeat("x", maxFileSize+1)
	writeTree(t, tmp, map[string]string{"big.md": huge})

	withRoot(t, tmp, func() {
		if _, err := readMarkdown("big.md"); err == nil {
			t.Errorf("expected size-limit error for oversized file")
		}
	})
}

func TestStripFrontmatter(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"yaml", "---\ntitle: Hi\ndate: 2026-01-01\n---\n# Body\n", "# Body\n"},
		{"toml", "+++\ntitle = \"Hi\"\n+++\n# Body\n", "# Body\n"},
		{"yaml-crlf", "---\r\ntitle: Hi\r\n---\r\n# Body\r\n", "# Body\r\n"},
		{"yaml-no-trailing-newline", "---\ntitle: Hi\n---", ""},
		{"no-frontmatter", "# Body\n", "# Body\n"},
		{"hr-not-frontmatter", "Some text\n\n---\n\nMore\n", "Some text\n\n---\n\nMore\n"},
		{"unterminated-yaml", "---\ntitle: Hi\nno close fence\n", "---\ntitle: Hi\nno close fence\n"},
		{"empty", "", ""},
		{"plain-dashes", "---", "---"},
		{"mismatched-fence", "---\ntitle: Hi\n+++\n# Body\n", "---\ntitle: Hi\n+++\n# Body\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(stripFrontmatter([]byte(tc.in)))
			if got != tc.want {
				t.Errorf("stripFrontmatter(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestResolveMarkdown(t *testing.T) {
	tmp := t.TempDir()
	writeTree(t, tmp, map[string]string{
		"README.md":         "x",
		"page.md":           "x",
		"TODO.MD":           "x",
		"sub/README.md":     "x",
		"sub/article.md":    "x",
		".hidden/secret.md": "x",
	})

	cases := map[string]string{
		"/":               "README.md",
		"/page":           "page.md",
		"/page.md":        "page.md",
		"/sub":            "sub/README.md",
		"/sub/article":    "sub/article.md",
		"/missing":        "",
		"/.hidden/secret": "", // dotdir is skipped
	}
	withRoot(t, tmp, func() {
		for url, want := range cases {
			got := resolveMarkdown(url)
			want = filepath.FromSlash(want)
			if got != want {
				t.Errorf("resolveMarkdown(%q) = %q, want %q", url, got, want)
			}
		}
		// Case-insensitive stem fallback: assert resolution succeeds without
		// pinning the exact on-disk casing returned, since case-sensitive
		// filesystems (Linux ext4) and case-insensitive ones (macOS APFS)
		// reach the matching file via different code paths.
		if got := resolveMarkdown("/TODO"); got == "" {
			t.Errorf("resolveMarkdown(/TODO) = \"\", want TODO.MD or TODO.md")
		}
	})
}

func TestExtraIndexNames(t *testing.T) {
	withIndexNames(t, "", func() {
		if got := extraIndexNames(); got != nil {
			t.Errorf("extraIndexNames() = %v, want nil for empty flag", got)
		}
	})

	withIndexNames(t, " index , home.md ,, ", func() {
		got := extraIndexNames()
		want := []string{"index", "home"}
		if len(got) != len(want) {
			t.Fatalf("extraIndexNames() = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("extraIndexNames()[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})
}

func TestResolveMarkdownIndexOverride(t *testing.T) {
	tmp := t.TempDir()
	writeTree(t, tmp, map[string]string{
		"sub/index.md":   "# Sub Index\n",
		"sub2/README.md": "# Sub2\n",
		"sub3/README.md": "# Sub3 Readme\n",
		"sub3/index.md":  "# Sub3 Index\n",
	})

	withRoot(t, tmp, func() {
		withIndexNames(t, "index", func() {
			if got := resolveMarkdown("/sub"); got != filepath.FromSlash("sub/index.md") {
				t.Errorf("resolveMarkdown(/sub) = %q, want sub/index.md", got)
			}
			// A directory with only a README still resolves via README.
			if got := resolveMarkdown("/sub2"); got != filepath.FromSlash("sub2/README.md") {
				t.Errorf("resolveMarkdown(/sub2) = %q, want sub2/README.md", got)
			}
			// README takes priority over the -index fallback when both exist.
			if got := resolveMarkdown("/sub3"); got != filepath.FromSlash("sub3/README.md") {
				t.Errorf("resolveMarkdown(/sub3) = %q, want sub3/README.md", got)
			}
		})
		// Without the flag set, the index.md-only directory is unreachable at /sub.
		if got := resolveMarkdown("/sub"); got != "" {
			t.Errorf("resolveMarkdown(/sub) with no -index flag = %q, want \"\"", got)
		}
	})
}

func TestUrlForFileIndexOverride(t *testing.T) {
	withIndexNames(t, "index,home", func() {
		cases := map[string]string{
			"sub/index.md": "/sub",
			"sub/home.md":  "/sub",
			"sub/other.md": "/sub/other",
			"index.md":     "/",
		}
		for in, want := range cases {
			if got := urlForFile(in); got != want {
				t.Errorf("urlForFile(%q) = %q, want %q", in, got, want)
			}
		}
	})
	// Without the flag set, "index.md" is just a regular page.
	if got := urlForFile("sub/index.md"); got != "/sub/index" {
		t.Errorf("urlForFile(sub/index.md) with no -index flag = %q, want /sub/index", got)
	}
}

func TestHandler(t *testing.T) {
	tmp := t.TempDir()
	writeTree(t, tmp, map[string]string{
		"README.md":     "# Home\n",
		"page.md":       "# Page\n",
		"sub/README.md": "# Sub\n",
		"img.png":       "PNGBYTES",
		".git/config":   "secret",
	})

	cases := []struct {
		name, path string
		wantCode   int
		wantSub    string // substring check on body when non-empty
	}{
		{"root", "/", 200, "<title>README</title>"},
		{"page", "/page", 200, "<title>page</title>"},
		{"page-with-md-suffix", "/page.md", 200, "<title>page</title>"},
		{"subdir-readme", "/sub", 200, "<title>README</title>"},
		{"image", "/img.png", 200, "PNGBYTES"},
		{"missing", "/nope", 404, ""},
		// Path traversal: path.Clean collapses these; defence in depth.
		{"traversal-shallow", "/../etc/passwd", 404, ""},
		{"traversal-deep", "/a/../../../etc/passwd", 404, ""},
		// Skipped paths must not be served even if they exist.
		{"hidden-dotdir", "/.git/config", 404, ""},
	}
	withRoot(t, tmp, func() {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", tc.path, nil)
				rec := httptest.NewRecorder()
				handler(rec, req)
				if rec.Code != tc.wantCode {
					t.Errorf("status = %d, want %d; body = %s", rec.Code, tc.wantCode, rec.Body.String())
				}
				if tc.wantSub != "" && !strings.Contains(rec.Body.String(), tc.wantSub) {
					t.Errorf("body missing %q", tc.wantSub)
				}
			})
		}
	})
}

func TestPrintVersion(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	printVersion()
	w.Close()
	os.Stdout = old

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "cortex") {
		t.Errorf("printVersion() output = %q, want it to contain %q", out, "cortex")
	}
}

func TestReloaderSubscribeCap(t *testing.T) {
	rl := newReloader()
	chans := make([]chan struct{}, 0, maxSSEClients)
	for i := range maxSSEClients {
		ch := rl.subscribe()
		if ch == nil {
			t.Fatalf("subscribe %d: got nil before reaching cap %d", i, maxSSEClients)
		}
		chans = append(chans, ch)
	}
	if ch := rl.subscribe(); ch != nil {
		t.Errorf("subscribe at cap: got a channel, want nil (cap %d reached)", maxSSEClients)
	}

	rl.unsubscribe(chans[0])
	if ch := rl.subscribe(); ch == nil {
		t.Errorf("subscribe after freeing a slot: got nil, want a channel")
	}
}

func TestReloaderBroadcast(t *testing.T) {
	rl := newReloader()
	ch := rl.subscribe()
	if ch == nil {
		t.Fatal("subscribe returned nil")
	}

	rl.broadcast()

	select {
	case <-ch:
	default:
		t.Error("broadcast did not reach subscribed channel")
	}
}

func TestExportSiteSkipsEscapingAssetSymlink(t *testing.T) {
	if _, err := os.Stat("/etc/hosts"); err != nil {
		t.Skip("no /etc/hosts on this platform")
	}
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	out := filepath.Join(tmp, "out")

	writeTree(t, src, map[string]string{
		"README.md": "# Home\n",
	})
	// Symlinked static asset whose target is outside the served root.
	if err := os.Symlink("/etc/hosts", filepath.Join(src, "escape.png")); err != nil {
		t.Fatal(err)
	}

	var exportErr error
	withRoot(t, src, func() {
		exportErr = exportSite(out)
	})
	if exportErr != nil {
		t.Fatalf("exportSite: %v", exportErr)
	}

	if _, err := os.Stat(filepath.Join(out, "escape.png")); err == nil {
		t.Errorf("exportSite copied an asset symlink escaping the root")
	}
}

func TestExportSite(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	out := filepath.Join(tmp, "out")

	writeTree(t, src, map[string]string{
		"README.md":       "# Home\n\n[page](page.md)\n[sub home](sub/README.md)\n",
		"page.md":         "# Page\n",
		"sub/README.md":   "# Sub\n",
		"sub/article.md":  "# Article\n",
		"img/diagram.png": "fakepngbytes",
	})

	var exportErr error
	withRoot(t, src, func() {
		exportErr = exportSite(out)
	})
	if exportErr != nil {
		t.Fatalf("exportSite: %v", exportErr)
	}

	expectFiles := []string{
		"index.html",
		"page/index.html",
		"sub/index.html",
		"sub/article/index.html",
		"img/diagram.png",
	}
	for _, f := range expectFiles {
		full := filepath.Join(out, f)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("missing output %s: %v", f, err)
		}
	}

	body, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	bodyStr := string(body)

	if strings.Contains(bodyStr, "EventSource") {
		t.Error("exported HTML must not contain live-reload script")
	}
	if !strings.Contains(bodyStr, `href="page/"`) {
		t.Errorf("expected rewritten link href=\"page/\" in index.html; got:\n%s", bodyStr)
	}
	if !strings.Contains(bodyStr, `href="sub/"`) {
		t.Errorf("expected rewritten link href=\"sub/\" in index.html; got:\n%s", bodyStr)
	}

	// Sidebar nav must use canonical URLs (no /README suffix for README files).
	if strings.Contains(bodyStr, `href="/README"`) {
		t.Errorf("sidebar must not link to /README; expected canonical / for root README")
	}
	if strings.Contains(bodyStr, `href="/sub/README"`) {
		t.Errorf("sidebar must not link to /sub/README; expected canonical /sub for sub README")
	}
}
