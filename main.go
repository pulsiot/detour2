package main

import (
	"crypto/tls"
	"log"
	"net/url"

	"github.com/valyala/fasthttp"
)

func main() {
	// Create a new TLS configuration
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12, // Set the minimum TLS version
		PreferServerCipherSuites: true,             // Prefer server cipher suites
	}

	// Create a new HTTP client with TLS support
	client := &fasthttp.Client{
		TLSConfig: tlsConfig,
	}

	// Create a new fasthttp server
	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// Extract the requested host
			host := string(ctx.Host())

			// Define the mapping of domains to target URLs
			targetURLs := map[string]string{
				"domain.com":        "https://www.google.com",
				"sales.domain.com":  "https://debuggerboy.com",
			}

			// Get the target URL for the requested host
			targetURL, ok := targetURLs[host]
			if !ok {
				ctx.Error("Unknown host", fasthttp.StatusNotFound)
				return
			}

			// Parse the target URL
			target, err := url.Parse(targetURL)
			if err != nil {
				ctx.Error("Invalid target URL", fasthttp.StatusInternalServerError)
				return
			}

			// Modify the request URL to the target URL + requested path
			ctx.URI().SetScheme(target.Scheme)
			ctx.URI().SetHost(target.Host)
			ctx.URI().SetPath(string(target.Path) + string(ctx.Path()))

			// Set the host header to the target host
			ctx.Request.Header.SetHost(target.Host)

			// Send the request to the target server
			resp := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseResponse(resp)

			err = client.Do(&ctx.Request, resp)
			if err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
				return
			}

			// Copy the response from the target server to the original response
			resp.CopyTo(&ctx.Response)
		},
	}

	// Listen for incoming HTTPS/TLS connections
	err := server.ListenAndServeTLS(":443", "server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}
}

