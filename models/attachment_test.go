package models

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/check.v1"
)

// buildDocxB64 builds a minimal in-memory zip archive containing the given
// named parts and returns it base64-encoded, as Attachment.Content expects.
// It's not a fully valid, openable Office document - just enough of a zip
// container for exercising ApplyTemplate's per-part XML handling in
// isolation, the same way Word's own OOXML parts are processed.
func buildDocxB64(c *check.C, parts map[string]string) string {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	for name, content := range parts {
		w, err := zw.Create(name)
		c.Assert(err, check.Equals, nil)
		_, err = w.Write([]byte(content))
		c.Assert(err, check.Equals, nil)
	}
	c.Assert(zw.Close(), check.Equals, nil)
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func (s *ModelsSuite) TestAttachment(c *check.C) {
	ptx := PhishingTemplateContext{
		BaseRecipient: BaseRecipient{
			FirstName: "Foo",
			LastName:  "Bar",
			Email:     "foo@bar.com",
			Position:  "Space Janitor",
		},
		BaseURL:     "http://testurl.com",
		URL:         "http://testurl.com/?rid=1234567",
		TrackingURL: "http://testurl.local/track?rid=1234567",
		Tracker:     "<img alt='' style='display: none' src='http://testurl.local/track?rid=1234567'/>",
		From:        "From Address",
		RId:         "1234567",
	}

	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		log.Fatalf("Failed to open attachment folder 'testdata': %v\n", err)
	}
	for _, ff := range files {
		if !ff.IsDir() && !strings.Contains(ff.Name(), "templated") {
			fname := ff.Name()
			fmt.Printf("Checking attachment file -> %s\n", fname)
			data := readFile("testdata/" + fname)
			if filepath.Ext(fname) == ".b64" {
				fname = fname[:len(fname)-4]
			}
			a := Attachment{
				Content: data,
				Name:    fname,
			}
			t, err := a.ApplyTemplate(ptx)
			c.Assert(err, check.Equals, nil)
			c.Assert(a.vanillaFile, check.Equals, strings.Contains(fname, "without-vars"))
			c.Assert(a.vanillaFile, check.Not(check.Equals), strings.Contains(fname, "with-vars"))

			// Verfify template was applied as expected
			tt, err := ioutil.ReadAll(t)
			if err != nil {
				log.Fatalf("Failed to parse templated file '%s': %v\n", fname, err)
			}
			templatedFile := base64.StdEncoding.EncodeToString(tt)
			expectedOutput := readFile("testdata/" + strings.TrimSuffix(ff.Name(), filepath.Ext(ff.Name())) + ".templated" + filepath.Ext(ff.Name())) // e.g text-file-with-vars.templated.txt
			c.Assert(templatedFile, check.Equals, expectedOutput)
		}
	}
}

// TestAttachmentRunSplitRepaired verifies that a {{ }} action which Word has
// fragmented across multiple XML runs - the cause of the
// "bad character U+003C '<'" template parse error - is repaired before the
// template is executed, so the variable still resolves correctly.
func (s *ModelsSuite) TestAttachmentRunSplitRepaired(c *check.C) {
	ptx := PhishingTemplateContext{
		AttachmentTrackingURL: "http://testurl.local/track/attachment?rid=1234567",
	}
	content := buildDocxB64(c, map[string]string{
		"word/document.xml": `<root><w:t>{{.AttachmentTracking</w:t></w:r><w:r><w:t>URL}}</w:t></root>`,
	})
	a := Attachment{Content: content, Name: "test.docx"}
	t, err := a.ApplyTemplate(ptx)
	c.Assert(err, check.Equals, nil)
	data := readAllBytes(c, t)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	c.Assert(err, check.Equals, nil)
	found := false
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			data, _ := ioutil.ReadAll(rc)
			c.Assert(strings.Contains(string(data), ptx.AttachmentTrackingURL), check.Equals, true)
			c.Assert(strings.Contains(string(data), "{{"), check.Equals, false)
			found = true
		}
	}
	c.Assert(found, check.Equals, true)
}

// TestAttachmentHTMLTrackerRejected verifies that using {{.Tracker}} or
// {{.AttachmentTracker}} - which render an HTML <img> tag - inside an Office
// XML part is rejected with a clear error instead of silently producing a
// document that Word will report as corrupted.
func (s *ModelsSuite) TestAttachmentHTMLTrackerRejected(c *check.C) {
	ptx := PhishingTemplateContext{
		AttachmentTracker: "<img alt='' style='display: none' src='http://testurl.local/track/attachment?rid=1234567'/>",
	}
	content := buildDocxB64(c, map[string]string{
		"word/document.xml": `<root><w:t>{{.AttachmentTracker}}</w:t></root>`,
	})
	a := Attachment{Content: content, Name: "test.docx"}
	_, err := a.ApplyTemplate(ptx)
	c.Assert(err, check.Not(check.Equals), nil)
	c.Assert(strings.Contains(err.Error(), "AttachmentTracker"), check.Equals, true)
}

// TestAttachmentSpecialCharactersEscaped verifies that a merge field value
// containing XML-significant characters (e.g. a recipient's name with an
// apostrophe or ampersand) is escaped rather than injected raw, so the
// resulting Office XML part stays well-formed instead of being corrupted.
func (s *ModelsSuite) TestAttachmentSpecialCharactersEscaped(c *check.C) {
	ptx := PhishingTemplateContext{
		BaseRecipient: BaseRecipient{
			FirstName: `Foo & <Bar> "Baz" 'Qux'`,
		},
	}
	content := buildDocxB64(c, map[string]string{
		"word/document.xml": `<root><w:t>{{.FirstName}}</w:t></root>`,
	})
	a := Attachment{Content: content, Name: "test.docx"}
	t, err := a.ApplyTemplate(ptx)
	c.Assert(err, check.Equals, nil)
	data := readAllBytes(c, t)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	c.Assert(err, check.Equals, nil)
	found := false
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			body, _ := ioutil.ReadAll(rc)
			c.Assert(validateWellFormedXML(body), check.Equals, nil)
			c.Assert(strings.Contains(string(body), "<Bar>"), check.Equals, false)
			found = true
		}
	}
	c.Assert(found, check.Equals, true)
}

// TestAttachmentURLEncodedActionCaseInsensitive verifies that Word's
// URL-encoded field-code form of a template action is un-escaped regardless
// of whether the '{' / '}' hex escapes are emitted in upper or lower case.
func (s *ModelsSuite) TestAttachmentURLEncodedActionCaseInsensitive(c *check.C) {
	ptx := PhishingTemplateContext{
		AttachmentTrackingURL: "http://testurl.local/track/attachment?rid=1234567",
	}
	content := buildDocxB64(c, map[string]string{
		"word/_rels/document.xml.rels": `<root>%7B%7B.AttachmentTrackingURL%7D%7D</root>`,
	})
	a := Attachment{Content: content, Name: "test.docx"}
	t, err := a.ApplyTemplate(ptx)
	c.Assert(err, check.Equals, nil)
	data := readAllBytes(c, t)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	c.Assert(err, check.Equals, nil)
	for _, f := range zr.File {
		rc, _ := f.Open()
		body, _ := ioutil.ReadAll(rc)
		c.Assert(strings.Contains(string(body), ptx.AttachmentTrackingURL), check.Equals, true)
	}
}

func readAllBytes(c *check.C, r io.Reader) []byte {
	data, err := ioutil.ReadAll(r)
	c.Assert(err, check.Equals, nil)
	return data
}

func readFile(fname string) string {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("Failed to open file '%s': %v\n", fname, err)
	}
	reader := bufio.NewReader(f)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read file '%s': %v\n", fname, err)
	}
	data := ""
	if filepath.Ext(fname) == ".b64" {
		data = string(content)
	} else {
		data = base64.StdEncoding.EncodeToString(content)
	}
	return data
}
