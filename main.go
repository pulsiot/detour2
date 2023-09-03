package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
)

// Mapping struct to hold the proxy domain and target URL
type Mapping struct {
	Domain    string `yaml:"domain"`
	TargetURL string `yaml:"targetURL"`
}

func main() {
	// Read the detour.yaml file
    configFile := "/etc/detour/detour2.yaml" // Replace with your YAML config file path
	detourData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Failed to read detour2.yaml:", err)
	}

	// Parse the YAML data
	var detourConfig struct {
		Port      int       `yaml:"port"`
		CertFile  string    `yaml:"certFile"`
		KeyFile   string    `yaml:"keyFile"`
		Mappings  []Mapping `yaml:"mappings"`
	}
	err = yaml.Unmarshal(detourData, &detourConfig)
	if err != nil {
		log.Fatal("Failed to parse detour2.yaml:", err)
	}

	// Create a new TLS configuration
	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12, // Set the minimum TLS version
		PreferServerCipherSuites: true,             // Prefer server cipher suites
	}

	// Create a new HTTP client with TLS support
	client := &fasthttp.Client{
		TLSConfig:           tlsConfig,
		MaxResponseBodySize: 10 * 1024 * 1024, // 10MB max response size
		ReadBufferSize:      4096,
		WriteBufferSize:     4096,
	}

	// Create a new fasthttp server
	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			// Extract the requested host
			host := string(ctx.Host())

			// Find the target URL for the requested host in the mappings
			var targetURL string
			for _, mapping := range detourConfig.Mappings {
				if mapping.Domain == host {
					targetURL = mapping.TargetURL
					break
				}
			}

			if targetURL == "" {
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
			ctx.URI().SetQueryStringBytes(ctx.QueryArgs().QueryString())

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
	err = server.ListenAndServeTLS(fmt.Sprintf(":%d", detourConfig.Port), detourConfig.CertFile, detourConfig.KeyFile)
	if err != nil {
		log.Fatal(err)
	}
}
