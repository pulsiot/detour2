# detour2 development

## TLS Support

We can create TLS Certificates using:

```
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -sha256 -days 3650 -nodes -subj "/C=IN/ST=Kerala/L=Kochi/O=Detour/OU=DetourProxy/CN=localhost"
```

Or

```
openssl req -x509 -new -newkey rsa:4096 -sha256 -days 3650 -nodes -out server.crt -keyout server.key

```
