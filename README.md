<h1 align="center">SimpleHTTPserver</h1>
<h4 align="center">Go alternative of python SimpleHTTPServer</h4>


<p align="center">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-_red.svg"></a>
<a href="https://github.com/projectdiscovery/simplehttpserver/issues"><img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat"></a>
<a href="https://goreportcard.com/badge/github.com/projectdiscovery/simplehttpserver"><img src="https://goreportcard.com/badge/github.com/projectdiscovery/simplehttpserver"></a>
<a href="https://hub.docker.com/r/projectdiscovery/simplehttpserver"><img src="https://img.shields.io/docker/pulls/projectdiscovery/simplehttpserver.svg"></a>
<a href="https://twitter.com/pdiscoveryio"><img src="https://img.shields.io/twitter/follow/pdiscoveryio.svg?logo=twitter"></a>
<a href="https://discord.gg/projectdiscovery"><img src="https://img.shields.io/discord/695645237418131507.svg?logo=discord"></a>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#usage">Usage</a> •
  <a href="#installing-simplehttpserver">Installation</a> •
  <a href="#running-simplehttpserver-in-the-current-folder">Run SimpleHTTPserver</a> •
  <a href="https://discord.gg/projectdiscovery">Join Discord</a>
</p>

---

SimpleHTTPserver is a go enhanced version of the well known python simplehttpserver with in addition a fully customizable TCP server, both supporting TLS.


# Features

- HTTP/S Web Server
- File Server with arbitrary directory support
- HTTP request/response dump
- Configurable ip address and listening port
- Configurable HTTP/TCP server with customizable response via YAML template


# Installing SimpleHTTPserver

SimpleHTTPserver requires **go1.17+** to install successfully. Run the following command to get the repo - 

```sh
go install -v github.com/projectdiscovery/simplehttpserver/cmd/simplehttpserver@latest
```

# Usage

```sh
simplehttpserver -h
```

This will display help for the tool. Here are all the switches it supports.

| Flag             | Description                                             | Example                                            |
|------------------|---------------------------------------------------------|----------------------------------------------------|
| `-listen`        | Configure listening ip:port (default 127.0.0.1:8000)    | `simplehttpserver -listen 127.0.0.1:8000`          |
| `-path`          | Fileserver folder (default current directory)           | `simplehttpserver -path /var/docs`                 |
| `-verbose`       | Verbose (dump request/response, default false)          | `simplehttpserver -verbose`                        |
| `-tcp`           | TCP server (default 127.0.0.1:8000)                     | `simplehttpserver -tcp 127.0.0.1:8000`             |
| `-tls`           | Enable TLS for TCP server                               | `simplehttpserver -tls`                            |
| `-rules`         | File containing yaml rules                              | `simplehttpserver -rules rule.yaml`                |
| `-upload`        | Enable file upload in case of http server               | `simplehttpserver -upload`                         |
| `-max-file-size` | Max Upload File Size (default 50 MB)                    | `simplehttpserver -max-file-size 100`              |
| `-sandbox`       | Enable sandbox mode                                     | `simplehttpserver -sandbox`                        |
| `-https`         | Enable HTTPS in case of http server                     | `simplehttpserver -https`                          |
| `-http1`         | Enable only HTTP1                                       | `simplehttpserver -http1`                          |
| `-cert`          | HTTPS/TLS certificate (self generated if not specified) | `simplehttpserver -cert cert.pem`                  |
| `-key`           | HTTPS/TLS certificate private key                       | `simplehttpserver -key cert.key`                   |
| `-domain`        | Domain name to use for the self-generated certificate   | `simplehttpserver -domain projectdiscovery.io`     |
| `-cors`          | Enable cross-origin resource sharing (CORS)             | `simplehttpserver -cors`                           |
| `-basic-auth`    | Basic auth (username:password)                          | `simplehttpserver -basic-auth user:password`       |
| `-realm`         | Basic auth message                                      | `simplehttpserver -realm "insert the credentials"` |
| `-version`       | Show version                                            | `simplehttpserver -version`                        |
| `-silent`        | Show only results                                       | `simplehttpserver -silent`                         |
| `-py`            | Emulate Python Style                                    | `simplehttpserver -py`                             |
| `-header`        | HTTP response header (can be used multiple times)       | `simplehttpserver -header 'X-Powered-By: Go'`      |

### Running simplehttpserver in the current folder  

This will run the tool exposing the current directory on port 8000 

```sh
simplehttpserver

2021/01/11 21:40:48 Serving . on http://0.0.0.0:8000/...
2021/01/11 21:41:15 [::1]:50181 "GET / HTTP/1.1" 200 383
2021/01/11 21:41:15 [::1]:50181 "GET /favicon.ico HTTP/1.1" 404 19
```

### Running simplehttpserver in the current folder with HTTPS

This will run the tool exposing the current directory on port 8000 over HTTPS with user provided certificate:

```sh
simplehttpserver -https -cert cert.pen -key cert.key

2021/01/11 21:40:48 Serving . on http://0.0.0.0:8000/...
2021/01/11 21:41:15 [::1]:50181 "GET / HTTP/1.1" 200 383
2021/01/11 21:41:15 [::1]:50181 "GET /favicon.ico HTTP/1.1" 404 19
```

Instead, to run with self-signed certificate and specific domain name:
```sh
simplehttpserver -https -domain localhost

2021/01/11 21:40:48 Serving . on http://0.0.0.0:8000/...
2021/01/11 21:41:15 [::1]:50181 "GET / HTTP/1.1" 200 383
2021/01/11 21:41:15 [::1]:50181 "GET /favicon.ico HTTP/1.1" 404 19
```

### Running simplehttpserver with basic auth and file upload

This will run the tool and will request the user to enter username and password before authorizing file uploads

```sh
simplehttpserver -basic-auth root:root -upload

2021/01/11 21:40:48 Serving . on http://0.0.0.0:8000/...
```

To upload files use the following curl request with basic auth header:
```sh
curl -v --user 'root:root' --upload-file file.txt http://localhost:8000/file.txt
```

### Running TCP server with custom responses

This will run the tool as TLS TCP server and enable custom responses based on YAML templates:

```sh
simplehttpserver -rules rules.yaml -tcp -tls -domain localhost
```

The rules are written as follows:
```yaml
rules:
  - match: regex-match
    match-contains: literal-match
    name: rule-name
    response: response data
```

For example to handle two different paths simulating an HTTP server or SMTP commands:
```yaml
rules:
  # HTTP Requests
  - match: GET /path1
    name: redirect
    response: |
              HTTP/1.0 200 OK
              Server: httpd/2.0
              x-frame-options: SAMEORIGIN
              x-xss-protection: 1; mode=block
              Date: Fri, 16 Apr 2021 14:30:32 GMT
              Content-Type: text/html
              Connection: close

              <HTML><HEAD><script>top.location.href='/Main_Login.asp';</script>
              </HEAD></HTML>
  - match: GET /path2
    name: "404"
    response: |
              HTTP/1.0 404 OK
              Server: httpd/2.0
            
              <HTML><HEAD></HEAD><BODY>Not found</BODY></HTML>
  # SMTP Commands
  - match: "EHLO example.com"
    name: smtp 
    response: |
              250-localhost Nice to meet you, [127.0.0.1]
              250-PIPELINING
              250-8BITMIME
              250-SMTPUTF8
              250-AUTH LOGIN PLAIN
              250 STARTTLS
  - match: "MAIL FROM: <noreply@example.com>"
    response: 250 Accepted
  - match: "RCPT TO: <test@example.com>"
    response: 250 Accepted

  - match-contains: !!binary |
      MAwCAQFgBwIBAwQAgAA=
    name: "ldap"
    # Request:  300c 0201 0160 0702 0103 0400 8000       0....`........
    # Response: 300c 0201 0161 070a 0100 0400 0400       0....a........
    response: !!binary |
      MAwCAQFhBwoBAAQABAA=
```

## Note

- This project is intended for development purposes only; it should not be used in production.

# Thanks

SimpleHTTPserver is made with 🖤 by the [projectdiscovery](https://projectdiscovery.io) team.
