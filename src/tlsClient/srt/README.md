# Spoof RoundTripper

A spoofed Go's http.RoundTripper that leverages uTLS to fake TLS fingerprints (JA3, JA4, HTTP/2 Akamai, etc) of mainstream browsers for use in different HTTP client libraries to bypass Cloudflare or other firewalls.