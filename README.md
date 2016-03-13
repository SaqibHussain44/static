# static

`static` is a super simple yaml-configured static file server that serves directories without authentication (over HTTP and HTTPS), or with HTTP Basic Authentication (over HTTPS only).

## Usage

```
$ ./static -h
Usage of ./static:
  -config string
      path to configuration file
  -gen-config
      generate example config file and print to stdout
```

## Example Configuration

```
$ ./static -gen-config
http_laddr: :80
https_laddr: :443
tls_cert_path: /etc/blah/example.cert
tls_key_path: /etc/blah/example.key
public_dirs:
- dir_path: /etc/www/pub1.com
  http_path: /pub1/
- dir_path: /etc/www/pub2
  http_path: /pub2/
authenticated_dirs:
- dir_path: /etc/www/secret
  http_path: /secret/
  usernames:
  - john
  - ha
users:
  ha: eioj
  huh: fjweoifj
  john: efjio
```
