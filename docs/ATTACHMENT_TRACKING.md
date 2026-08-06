# Attachment Tracking in Gophish

This document describes the new attachment tracking feature that allows separate tracking of email opens vs attachment opens.

## Overview

Gophish now supports tracking when recipients open attachments separately from when they open emails. This is useful for understanding recipient behavior in campaigns that include email attachments (e.g., Word documents, PDFs).

## New Template Variables

There are two variables, and which one you use depends entirely on whether the
attachment is an **HTML file** or an **Office document (.docx/.docm/.pptx/.xlsx/.xlsm)**.
Office documents are XML, not HTML - an HTML `<img>` tag has no meaning inside
one and using the wrong variable there will corrupt the file. Gophish will
reject the wrong combination with a clear error at template-save time, but
it's worth understanding why so the field is actually functional in Word.

### `{{.AttachmentTracker}}` / `{{.Tracker}}`

A tracking image **HTML `<img>` tag**, fully rendered:
```html
<img alt='' style='display: none' src='https://your-phishing-url/track/attachment?rid=xxxxx'/>
```

**When to use:**
- Inside the **email HTML body** (`{{.Tracker}}`)
- Inside **`.html` attachments** (`{{.AttachmentTracker}}`)
- Anywhere the surrounding document is genuinely HTML and can render an `<img>` tag

**Do not use inside `.docx`/`.docm`/`.pptx`/`.xlsx`/`.xlsm` attachments.** Those
files are XML containers (WordprocessingML/DrawingML/SpreadsheetML), and an
HTML `<img>` element is not valid content there - Word will report the file
as corrupted. Gophish now detects and rejects this combination when you save
the template, with a message pointing you at `{{.AttachmentTrackingURL}}` instead.

### `{{.AttachmentTrackingURL}}` / `{{.TrackingURL}}`

The raw tracking URL, with no HTML wrapper:
```
https://your-phishing-url/track/attachment?rid=xxxxx
```

**When to use:**
- Inside `.docx`/`.docm`/`.pptx`/`.xlsx`/`.xlsm` attachments - this is the **only**
  variable that's safe to use there, and it must be wired into a real Office
  field (see the walkthrough below), not typed as plain visible text.
- Anywhere else you need just the URL (custom image tags, scripting, etc.)

## Comparison with Email Tracking

| Variable | Purpose | URL Path | Event Type |
|----------|---------|----------|------------|
| `{{.Tracker}}` | Email body tracking | `/track?rid=xxx` | "Email Opened" |
| `{{.AttachmentTracker}}` | Attachment tracking | `/track/attachment?rid=xxx` | "Attachment Opened" |

## Statistics

The new `attachment_opened` statistic is now included in campaign statistics:

```json
{
  "stats": {
    "total": 100,
    "sent": 100,
    "opened": 45,
    "attachment_opened": 23,
    "clicked": 15,
    "submitted_data": 5,
    "email_reported": 2,
    "error": 0
  }
}
```

## Dashboard Display

Both the main dashboard and campaign results pages now display 6 pie charts:
1. Emails Sent
2. Emails Opened
3. **Attachments Opened** (new)
4. Clicked Link
5. Submitted Data
6. Email Reported

## Status Priority

When updating a recipient's status, the following priority is maintained:
1. Submitted Data (highest)
2. Clicked Link
3. Email Opened
4. Attachment Opened
5. Email Sent (lowest)

If a recipient opens an attachment but has already opened the email or clicked a link, their status won't be downgraded.

## Example: Word Document with Tracking

To track when a Word document attachment is opened, `{{.AttachmentTrackingURL}}`
needs to be wired into a real Word **field** - not typed as plain visible text.
Typed text just becomes a readable URL on the page; it never causes Word to
fetch anything.

1. In Word, place your cursor where the tracking image should go.
2. Press `Ctrl+F9` to insert a pair of empty field braces (`{ }`) - don't type
   braces yourself, Word needs to create real field delimiters.
3. Between the braces, type:
   ```
   INCLUDEPICTURE "{{.AttachmentTrackingURL}}" \* MERGEFORMAT
   ```
4. Press `F9` to update the field (or `Alt+F9` to toggle field codes off) and save.
5. Upload the `.docx` as an attachment on your email template. Gophish will
   substitute the real tracking URL into the field when the campaign sends.
6. When the recipient opens the document (with internet access and remote
   content allowed - see the caveat below), Word requests the image and the
   attachment-open event is recorded.

### Gotchas

- **Word silently splits typed text across multiple XML runs.** Autocorrect,
  spell-check, and grammar-check can all fragment a variable you've typed
  (`{{.AttachmentTrackingURL}}`) into multiple `<w:r>` runs behind the scenes,
  even though it looks like one continuous line on screen. Gophish repairs
  this automatically before templating, but if you still hit a template
  error, try disabling autocorrect-as-you-type first, or paste the variable
  with **Paste Special → Unformatted Text** instead of typing it directly.
- **Word may URL-encode the field's contents**, turning `{{.Foo}}` into
  something like `%7B%7B.Foo%7D%7D`. Gophish detects and reverses this
  automatically (case-insensitively) before templating.
- **`{{.Tracker}}` / `{{.AttachmentTracker}}` (the HTML `<img>` tag variables)
  are rejected outright** if used inside a `.docx`/`.docm`/`.pptx`/`.xlsx`/`.xlsm`
  attachment, at template-save time, with a message pointing you at
  `{{.AttachmentTrackingURL}}` instead - see [New Template Variables](#new-template-variables) above.
- **Word/Outlook may not auto-fetch the image at all**, independent of any of
  the above. Word blocks automatic loading of remote linked content by
  default for files that carry the "downloaded from the internet"/email
  mark of the web (Protected View, no auto-update of fields), so the
  recipient may need to click through a trust prompt or manually update the
  field (`F9`) before the pixel fires. This is a property of Word's own
  security model, not something Gophish can control - test the actual
  end-to-end behavior in your target environment rather than assuming it
  fires unconditionally.

## Architecture Note: Event-Based Counting

All campaign statistics (Sent, Opened, Clicked, Attachment Opened, Submitted, Reported) now use **event-based counting** derived directly from the events table. This ensures that:

- Counts are always additive (never decrease).
- Counts represent unique recipients for each specific action.
- Parallel actions (e.g. opening an attachment AND clicking a link) are both counted accurately without overwriting each other.

The frontend uses server-provided `campaign.stats` for all charts, ensuring data consistency across the dashboard and campaign results pages.

## Breaking Changes

Existing templates that use `{{.Tracker}}` inside an **HTML attachment** will
continue to work, but will record an "Email Opened" event rather than
"Attachment Opened" - use `{{.AttachmentTracker}}` there instead to get the
separate event.

Existing templates that use `{{.Tracker}}` or `{{.AttachmentTracker}}` inside an
**Office document attachment** (`.docx`/`.docm`/`.pptx`/`.xlsx`/`.xlsm`) will now
be **rejected with an error when you save the template**, rather than silently
producing a file that Word reports as corrupted. Switch those attachments to
`{{.AttachmentTrackingURL}}`, wired into a real Office field as described above.

To use the new separate tracking:
1. Update your email templates to use `{{.Tracker}}` in the email body only.
2. Update HTML attachments to use `{{.AttachmentTracker}}`.
3. Update Office document attachments to use `{{.AttachmentTrackingURL}}` via a real field.
