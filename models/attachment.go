package models

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// wordEscapedActionRx matches a template action, tolerating XML markup that
// Word/Office may have inserted in the middle of it (e.g. spell-check or
// autocorrect splitting the typed text across multiple <w:r> runs).
var wordEscapedActionRx = regexp.MustCompile(`(?s)\{\{.*?\}\}`)

// xmlTagRx matches a single XML tag, used to strip tags that Word injected
// inside an otherwise-intact {{ }} template action.
var xmlTagRx = regexp.MustCompile(`<[^>]+>`)

// wordURLEncodedActionRx matches a {{.Field}} action that Word has
// URL-encoded (this happens when the variable is typed inside certain
// field codes, e.g. INCLUDEPICTURE). The hex escapes for '{' and '}' may be
// emitted in either case per RFC 3986, so the match is case-insensitive.
var wordURLEncodedActionRx = regexp.MustCompile(`(?i)%7b%7b\.([a-zA-Z]+)%7d%7d`)

// htmlTrackerInOfficeXMLRx detects use of the HTML <img>-tag tracker
// variables ({{.Tracker}} / {{.AttachmentTracker}}) inside an Office XML
// document part. These variables render literal HTML, which has no meaning
// in WordprocessingML/DrawingML/SpreadsheetML and will corrupt the file -
// {{.TrackingURL}} / {{.AttachmentTrackingURL}} (plain URLs) must be used
// instead, wired up via a real Office field.
var htmlTrackerInOfficeXMLRx = regexp.MustCompile(`\{\{\s*\.\s*(?:Attachment)?Tracker\s*\}\}`)

// defragmentTemplateActions repairs {{ }} template actions that Word has
// split across multiple XML runs by stripping any XML tags found strictly
// between the opening {{ and its matching }}. Everything outside of a
// template action is left untouched.
func defragmentTemplateActions(contents []byte) []byte {
	return wordEscapedActionRx.ReplaceAllFunc(contents, func(action []byte) []byte {
		return xmlTagRx.ReplaceAll(action, []byte(""))
	})
}

// xmlEscapeString escapes s so that it's safe to embed inside the text
// content or an attribute value of an XML document.
func xmlEscapeString(s string) string {
	buf := &bytes.Buffer{}
	xml.EscapeText(buf, []byte(s))
	return buf.String()
}

// xmlSafeContext returns a copy of ptx with every field XML-escaped. This is
// used when templating the internal XML parts of Office document formats
// (docx/docm/pptx/xlsx/xlsm), since those files are XML and, unlike an HTML
// email body, must never have unescaped angle brackets, ampersands, or
// quote characters injected into them - whether that value came from the
// phishing URL, a recipient's name, or anything else in the context.
func xmlSafeContext(ptx PhishingTemplateContext) PhishingTemplateContext {
	safe := ptx
	safe.From = xmlEscapeString(ptx.From)
	safe.URL = xmlEscapeString(ptx.URL)
	safe.Tracker = xmlEscapeString(ptx.Tracker)
	safe.TrackingURL = xmlEscapeString(ptx.TrackingURL)
	safe.AttachmentTracker = xmlEscapeString(ptx.AttachmentTracker)
	safe.AttachmentTrackingURL = xmlEscapeString(ptx.AttachmentTrackingURL)
	safe.RId = xmlEscapeString(ptx.RId)
	safe.BaseURL = xmlEscapeString(ptx.BaseURL)
	safe.BaseRecipient.Email = xmlEscapeString(ptx.BaseRecipient.Email)
	safe.BaseRecipient.FirstName = xmlEscapeString(ptx.BaseRecipient.FirstName)
	safe.BaseRecipient.LastName = xmlEscapeString(ptx.BaseRecipient.LastName)
	safe.BaseRecipient.Position = xmlEscapeString(ptx.BaseRecipient.Position)
	return safe
}

// validateWellFormedXML confirms that data is well-formed XML. It's used as
// a last line of defense to catch a templated Office XML part that would
// otherwise silently corrupt the document, so that Gophish fails loudly at
// template-save/send time instead of the recipient discovering it when
// Word/Excel/PowerPoint reports the file as corrupted.
func validateWellFormedXML(data []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		_, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// Attachment contains the fields and methods for
// an email attachment
type Attachment struct {
	Id          int64  `json:"-"`
	TemplateId  int64  `json:"-"`
	Content     string `json:"content"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	vanillaFile bool   // Vanilla file has no template variables
}

// Validate ensures that the provided attachment uses the supported template variables correctly.
func (a Attachment) Validate() error {
	vc := ValidationContext{
		FromAddress: "foo@bar.com",
		BaseURL:     "http://example.com",
	}
	td := Result{
		BaseRecipient: BaseRecipient{
			Email:     "foo@bar.com",
			FirstName: "Foo",
			LastName:  "Bar",
			Position:  "Test",
		},
		RId: "123456",
	}
	ptx, err := NewPhishingTemplateContext(vc, td.BaseRecipient, td.RId)
	if err != nil {
		return err
	}
	_, err = a.ApplyTemplate(ptx)
	return err
}

// ApplyTemplate parses different attachment files and applies the supplied phishing template.
func (a *Attachment) ApplyTemplate(ptx PhishingTemplateContext) (io.Reader, error) {

	decodedAttachment := base64.NewDecoder(base64.StdEncoding, strings.NewReader(a.Content))

	// If we've already determined there are no template variables in this attachment return it immediately
	if a.vanillaFile == true {
		return decodedAttachment, nil
	}

	// Decided to use the file extension rather than the content type, as there seems to be quite
	//  a bit of variability with types. e.g sometimes a Word docx file would have:
	//   "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	fileExtension := filepath.Ext(a.Name)

	switch fileExtension {

	case ".docx", ".docm", ".pptx", ".xlsx", ".xlsm":
		// Most modern office formats are xml based and can be unarchived.
		// .docm and .xlsm files are comprised of xml, and a binary blob for the macro code

		// Zip archives require random access for reading, so it's hard to stream bytes. Solution seems to be to use a buffer.
		// See https://stackoverflow.com/questions/16946978/how-to-unzip-io-readcloser
		b := new(bytes.Buffer)
		b.ReadFrom(decodedAttachment)
		zipReader, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(b.Len())) // Create a new zip reader from the file

		if err != nil {
			return nil, err
		}

		newZipArchive := new(bytes.Buffer)
		zipWriter := zip.NewWriter(newZipArchive) // For writing the new archive

		// i. Read each file from the Word document archive
		// ii. Apply the template to it
		// iii. Add the templated content to a new zip Word archive
		a.vanillaFile = true
		for _, zipFile := range zipReader.File {
			ff, err := zipFile.Open()
			if err != nil {
				return nil, err
			}
			defer ff.Close()
			contents, err := ioutil.ReadAll(ff)
			if err != nil {
				return nil, err
			}
			subFileExtension := filepath.Ext(zipFile.Name)
			var tFile string
			if subFileExtension == ".xml" || subFileExtension == ".rels" { // Ignore other files, e.g binary ones and images
				// Word sometimes splits a typed {{.Foo}} across multiple XML
				// runs (autocorrect/spell-check/grammar-check all create new
				// <w:r> runs), leaving raw XML tags in between the braces.
				// Strip those out so the template action is contiguous again.
				contents := defragmentTemplateActions(contents)

				// Word can also URL-encode our template variables entirely,
				// turning {{.Foo}} into %7b%7b.foo%7d%7d. This seems to
				// happen when inserting a remote image via a field code.
				// See https://stackoverflow.com/questions/68287630/disable-url-encoding-for-includepicture-in-microsoft-word
				contents = wordURLEncodedActionRx.ReplaceAllFunc(contents, func(m []byte) []byte {
					d, err := url.QueryUnescape(string(m))
					if err != nil {
						return m
					}
					return []byte(d)
				})

				// The URL-decode step above can itself reveal an action that
				// was split by an embedded tag, so repair it once more.
				contents = defragmentTemplateActions(contents)

				// {{.Tracker}} / {{.AttachmentTracker}} render a literal
				// HTML <img> tag, which is meaningless (and corrupting)
				// inside an Office XML part. Reject it early with a clear
				// message rather than let it silently break the document.
				if htmlTrackerInOfficeXMLRx.Match(contents) {
					zipWriter.Close()
					return nil, fmt.Errorf("%s: {{.Tracker}} and {{.AttachmentTracker}} render an HTML <img> tag and can't be used inside an Office document (%s) - use {{.TrackingURL}} or {{.AttachmentTrackingURL}} instead, inserted via a real Word/Office field", zipFile.Name, fileExtension)
				}

				// For each file apply the template, escaping every
				// substituted value for safe inclusion in XML - this is not
				// HTML, so nothing (a URL, a recipient's name, anything
				// else in the context) should be injected unescaped here.
				tFile, err = ExecuteTemplate(string(contents), xmlSafeContext(ptx))
				if err != nil {
					zipWriter.Close() // Don't use defer when writing files https://www.joeshaw.org/dont-defer-close-on-writable-files/
					return nil, err
				}
				// Belt-and-suspenders: confirm the templated part is still
				// well-formed XML before we ship it, so a bug here fails
				// loudly at save/send time instead of producing a file that
				// only breaks when the recipient opens it.
				if err = validateWellFormedXML([]byte(tFile)); err != nil {
					zipWriter.Close()
					return nil, fmt.Errorf("%s: template produced invalid XML: %w", zipFile.Name, err)
				}
				// Check if the subfile changed. We only need this to be set once to know in the future to check the 'parent' file
				if tFile != string(contents) {
					a.vanillaFile = false
				}
			} else {
				tFile = string(contents) // Could move this to the declaration of tFile, but might be confusing to read
			}
			// Write new Word archive
			newZipFile, err := zipWriter.Create(zipFile.Name)
			if err != nil {
				zipWriter.Close() // Don't use defer when writing files https://www.joeshaw.org/dont-defer-close-on-writable-files/
				return nil, err
			}
			_, err = newZipFile.Write([]byte(tFile))
			if err != nil {
				zipWriter.Close()
				return nil, err
			}
		}
		zipWriter.Close()
		return bytes.NewReader(newZipArchive.Bytes()), err

	case ".txt", ".html", ".ics":
		b, err := ioutil.ReadAll(decodedAttachment)
		if err != nil {
			return nil, err
		}
		processedAttachment, err := ExecuteTemplate(string(b), ptx)
		if err != nil {
			return nil, err
		}
		if processedAttachment == string(b) {
			a.vanillaFile = true
		}
		return strings.NewReader(processedAttachment), nil
	default:
		return decodedAttachment, nil // Default is to simply return the file
	}

}
