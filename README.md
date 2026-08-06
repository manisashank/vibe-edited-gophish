![gophish logo](https://raw.github.com/gophish/gophish/master/static/images/gophish_purple.png)

Gophish
=======

![Build Status](https://github.com/gophish/gophish/workflows/CI/badge.svg) [![GoDoc](https://godoc.org/github.com/gophish/gophish?status.svg)](https://godoc.org/github.com/gophish/gophish)

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

### Install

Installation of Gophish is dead-simple - just download and extract the zip containing the [release for your system](https://github.com/gophish/gophish/releases/), and run the binary. Gophish has binary releases for Windows, Mac, and Linux platforms.

### Key Features
- **Phishing Campaigns**: Launch and track campaigns.
- **Attachment Tracking**: Track when recipients open attachments (Word, HTML) separately from email opens. [Read the Guide](docs/ATTACHMENT_TRACKING.md)
- **Email Tracking**: Know when emails are opened.
- **Click Tracking**: Track links clicked and credentials submitted.

### Building From Source
**If you are building from source, please note that Gophish requires Go v1.10 or above!**

To build Gophish from source, simply run ```git clone https://github.com/manisashank/vibe-edited-gophish.git``` and ```cd``` into the project source directory. Then, run ```go build```. After this, you should have a binary called ```gophish``` in the current directory.

### Setup
After running the Gophish binary, open an Internet browser to https://localhost:3333 and login with the default username and password listed in the log output.
e.g.
```
time="2020-07-29T01:24:08Z" level=info msg="Please login with the username admin and the password 4304d5255378177d"
```

### Documentation

Documentation can be found on [site](http://getgophish.com/documentation). Find something missing? Let us know by filing an issue!

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let us know! Please don't hesitate to [file an issue](https://github.com/gophish/gophish/issues/new) and we'll get right on it.

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
