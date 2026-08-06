![gophish logo](https://raw.github.com/gophish/gophish/master/static/images/gophish_purple.png)

Gophish
=======

![Build Status](https://github.com/manisashank/vibe-edited-gophish/workflows/CI/badge.svg)

Gophish: Open-Source Phishing Toolkit

[Gophish](https://getgophish.com) is an open-source phishing toolkit designed for businesses and penetration testers. It provides the ability to quickly and easily setup and execute phishing engagements and security awareness training.

### About This Fork

This is a modified ("vibe-edited") version of the original [Gophish](https://github.com/gophish/gophish) project.

In upstream Gophish, opening an emailed attachment isn't tracked as its own event - a recipient opening a Word document attachment shows up the same as (or gets conflated with) opening the email itself, so there's no way to tell whether someone actually opened the attached file at all, independent of the email or the link. This fork adds first-class, **independent attachment-open tracking**, so a campaign's results can distinguish every one of these outcomes per recipient:

- **Email Sent** - the email was successfully handed off to the SMTP server
- **Email Opened** - the recipient opened the email (the tracking pixel in the email body loaded)
- **Attachment Opened** - the recipient opened the attached file (Word document, HTML attachment, etc.) - tracked separately from the email/link events, via its own tracking pixel embedded in the attachment
- **Clicked Link** - the recipient clicked the phishing link in the email
- **Submitted Data** - the recipient submitted credentials/data on the landing page
- **Email Reported** - the recipient reported the email as suspicious

These are all recorded as distinct events and surfaced as separate stats/charts on the dashboard and campaign results pages, with a defined priority so a recipient's status is never downgraded (e.g. someone who clicked the link and *then* opened the attachment still shows as "Clicked Link", not "Attachment Opened").

See [docs/ATTACHMENT_TRACKING.md](docs/ATTACHMENT_TRACKING.md) for how to wire up attachment tracking in your own email templates.

### Key Features
- **Phishing Campaigns**: Launch and track campaigns.
- **Attachment Tracking**: Track when recipients open attachments (Word, HTML) separately from email opens. [Read the Guide](docs/ATTACHMENT_TRACKING.md)
- **Email Tracking**: Know when emails are opened.
- **Click Tracking**: Track links clicked and credentials submitted.

### Usage

#### Download a release (Windows, Linux, macOS)

The easiest way to run this fork is to download a prebuilt release for your platform from the
[**Releases page**](https://github.com/manisashank/vibe-edited-gophish/releases):

- **Windows**: download `gophish-<version>-windows-64bit.zip`, extract it, and run `gophish.exe`.
- **Linux**: download `gophish-<version>-linux-64bit.zip`, extract it, and run `./gophish`.
- **macOS**: download `gophish-<version>-osx-64bit.zip`, extract it, and run `./gophish`.

Run the binary from the directory it was extracted into (it expects `config.json`, `templates/`, `static/`, and `db/` alongside it).

#### Building from source

Building from source requires **Go v1.21 or above**, plus a C compiler (needed for the SQLite3 database driver via cgo - e.g. `build-essential`/`gcc` on Linux, Xcode Command Line Tools on macOS, or a MinGW-w64 toolchain on Windows).

```
git clone https://github.com/manisashank/vibe-edited-gophish.git
cd vibe-edited-gophish
go build
```

This produces a `gophish` (or `gophish.exe` on Windows) binary in the current directory.

### Setup
After running the Gophish binary, open an Internet browser to https://localhost:3333 and login with the default username and password listed in the log output.
e.g.
```
time="2020-07-29T01:24:08Z" level=info msg="Please login with the username admin and the password 4304d5255378177d"
```

### Documentation

General Gophish documentation can be found on the upstream project's [site](http://getgophish.com/documentation) - note that it describes upstream behavior and doesn't cover this fork's changes. For attachment tracking specifically (the main thing this fork adds), see [docs/ATTACHMENT_TRACKING.md](docs/ATTACHMENT_TRACKING.md).

### Issues

Find a bug? Want more features? Find something missing in the documentation? Please [file an issue](https://github.com/manisashank/vibe-edited-gophish/issues/new) on this repository.

### License
```
Gophish - Open-Source Phishing Framework

The MIT License (MIT)

Copyright (c) 2013 - 2020 Jordan Wright

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software ("Gophish Community Edition") and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
