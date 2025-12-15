# Attachment Tracking in Gophish

This document describes the new attachment tracking feature that allows separate tracking of email opens vs attachment opens.

## Overview

Gophish now supports tracking when recipients open attachments separately from when they open emails. This is useful for understanding recipient behavior in campaigns that include email attachments (e.g., Word documents, PDFs).

## New Template Variables

### `{{.AttachmentTracker}}`

A tracking image HTML tag designed for use inside attachments (e.g., Word documents).

**Usage:**
```html
<img alt='' style='display: none' src='{{.AttachmentTrackingURL}}'/>
```

Or use the shorthand which generates the full img tag:
```
{{.AttachmentTracker}}
```

**When to use:**
- Inside `.docx` Word documents
- Inside `.html` attachments
- Any attachment that can render HTML/images

### `{{.AttachmentTrackingURL}}`

The raw URL for attachment tracking, without the img tag wrapper.

**URL Format:**
```
https://your-phishing-url/track/attachment?rid=xxxxx
```

**When to use:**
- When you need just the URL (e.g., for custom image tags or other purposes)

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

To track when a Word document attachment is opened:

1. Create your `.docx` document
2. Add an image placeholder or use the Web Parts feature
3. Reference the `{{.AttachmentTrackingURL}}` for the image source
4. When the document is opened (and internet connected), the tracking pixel will load

**Note:** Word documents only load external images when:
- The recipient is connected to the internet
- The recipient hasn't blocked external content loading
- The document is opened in an application that renders external images

## Architecture Note: Event-Based Counting

All campaign statistics (Sent, Opened, Clicked, Attachment Opened, Submitted, Reported) now use **event-based counting** derived directly from the events table. This ensures that:

- Counts are always additive (never decrease).
- Counts represent unique recipients for each specific action.
- Parallel actions (e.g. opening an attachment AND clicking a link) are both counted accurately without overwriting each other.

The frontend uses server-provided `campaign.stats` for all charts, ensuring data consistency across the dashboard and campaign results pages.

## Breaking Changes

Existing templates that use `{{.Tracker}}` inside attachments will continue to work but will record events as "Email Opened" instead of the new "Attachment Opened" event.

To use the new separate tracking:
1. Update your email templates to use `{{.Tracker}}` in the email body only
2. Update your attachments to use `{{.AttachmentTracker}}` or `{{.AttachmentTrackingURL}}` instead
