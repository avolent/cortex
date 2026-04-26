package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withRoot runs fn with *rootDir temporarily set, then restores it.
// All filesystem-touching tests share this helper since *rootDir is global.
func withRoot(t *testing.T, dir string, fn func()) {
	t.Helper()
	old := *rootDir
	*rootDir = dir
	defer func() { *rootDir = old }()
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
