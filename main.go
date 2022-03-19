package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
)

// WebReverseProxyConfiguration is a coniguration for the ReverseProxy
type WebReverseProxyConfiguration struct {
	RegistryProxyHost string
	ReleaseProxyHost  string
}

func main() {
	config := &WebReverseProxyConfiguration{
		RegistryProxyHost: "registry.local",
		ReleaseProxyHost:  "release.local",
	}
	proxy := NewWebReverseProxy(config)
	http.Handle("/", handlers.LoggingHandler(os.Stdout, proxy))

	// Start the server
	http.ListenAndServe(":8555", nil)
}

// This replaces all occurrences of http://releases.hashicorp.com with
// config.ReleaseProxyHost in the response body
func rewriteBody(config *WebReverseProxyConfiguration, resp *http.Response) (err error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = resp.Body.Close(); err != nil {
		return err
	}

	replacement := fmt.Sprintf("https://%s", config.ReleaseProxyHost)
	fmt.Printf("replacement: %+v\n", replacement)

	fmt.Printf("before: %+v\n", string(b))

	b = bytes.ReplaceAll(b, []byte("https://releases.hashicorp.com"), []byte(replacement)) // releases
	fmt.Printf("after: %+v\n", string(b))
	body := ioutil.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	return nil
}

func NewWebReverseProxy(config *WebReverseProxyConfiguration) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		if req.Host == config.RegistryProxyHost {
			req.URL.Scheme = "https"
			req.URL.Host = "registry.terraform.io"
			req.Header.Set("X-Terraform-Version", "1.1.7")
			req.Header.Set("User-Agent", "Terraform/1.1.7")
		} else if req.Host == config.ReleaseProxyHost {
			req.URL.Scheme = "https"
			req.URL.Host = "releases.hashicorp.com"
			req.Header.Set("User-Agent", "Terraform/1.1.7")
		}
	}

	responseDirector := func(res *http.Response) error {
		rewriteBody(config, res)
		if location := res.Header.Get("Location"); location != "" {
			url, err := url.ParseRequestURI(location)
			if err != nil {
				fmt.Println("Error!")
				return err
			}

			// Override redirect url Host with ProxyHost
			url.Host = config.RegistryProxyHost

			res.Header.Set("Location", url.String())
			res.Header.Set("X-Reverse-Proxy", "terraform-registry-reverse-proxy")
		}
		return nil
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: responseDirector,
	}
}
