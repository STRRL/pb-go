# pb-go

![Logo](./readme-logo.png)

Yet Another PasteBin implemented in Golang.

![GitHub stars](https://img.shields.io/github/stars/kmahyyg/pb-go?style=social)
![Go Report](https://goreportcard.com/badge/github.com/kmahyyg/pb-go)
[![Build Status](https://travis-ci.com/kmahyyg/pb-go.svg?branch=master)](https://travis-ci.com/kmahyyg/pb-go)
![GitHub](https://img.shields.io/github/license/kmahyyg/pb-go)
![GitHub last commit](https://img.shields.io/github/last-commit/kmahyyg/pb-go)
![GitHub All Releases](https://img.shields.io/github/downloads/kmahyyg/pb-go/total)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/269b77a2b64c41bbaa4aa109ecf4d55a)](https://www.codacy.com/manual/kmahyyg/pb-go)
![Codacy coverage](https://img.shields.io/codacy/coverage/269b77a2b64c41bbaa4aa109ecf4d55a?logo=codacy)

We use [Sentry.io](https://sentry.io) for bug tracking and log collection which was GDPR-complaint, 
their privacy policy can be found at: [here](https://sentry.io/legal/privacy/2.1.0/)

### Discussion

We need developer and help, for feature request and discussion, please go to our [Telegram Group](https://t.me/pb_go_discuss).

Bug report please attach log and finish the whole issue template. Thanks.

## Prerequisites

- MongoDB
- Reverse Proxy with HTTPS and Rate-Limit Support (Recommend: Traefik, Caddy)
- A Linux Server (If you need Windows version, compile by yourself.)

Note: Since we are offering public services, we don't want to implement any rate-limit
on application side. You must apply a reverse proxy or something else do that.
Your data is encrypted and finally stored on our server using Chacha20 algorithm.

## To-Do list (features)

- [ ] | Content detection, only allow pure texts.
- [X] [ ] | Expiring feature done in MongoDB. <del> (TTL done, DB Driver not implement) </del>
- [ ] | Private Share optionally, Share password using BLAKE2b stored. 
- [X] | <del> Rate-limit to avoid abusing. (SHOULD BE DONE IN REVERSE PROXY SIDE) </del>
- [ ] | ReCaptcha v2 support to prevent from a large scale abusing.
- [X] | Code Syntax Highlighting.
- [ ] | Shortlink using hashids.
- [X] | <del> Pure CLI. (Use `curl` instead)</del>
- [X] | Web page upload.

## Usage

TODO

## Compile

TODO

## License

 pb-go
 Copyright (C) 2020  kmahyyg
 
 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.
 
 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.
 
 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.

