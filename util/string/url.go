// Copyright 2023 Nautes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package string

import (
	"fmt"
	"net"
	"net/url"
)

func ExtractPortFromURL(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if u.Hostname() == "" {
		return "", nil
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "", err
	}
	return port, nil
}

func ParseUrl(urlString string) (string, error) {
	// Check if the URL is empty
	if urlString == "" {
		return "", fmt.Errorf("the URL is empty")
	}

	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	// Check if the protocol is HTTP or HTTPS
	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return "", fmt.Errorf("the URL is not using the HTTP or HTTPS protocol")
	}

	// Parse the host part to get the IP address
	host, _, err := net.SplitHostPort(parsedUrl.Host)
	if err != nil {
		// If the host part doesn't contain a port number, SplitHostPort returns an error.
		// In this case, we can assume that the host part is an IP address.
		host = parsedUrl.Host
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("the host part of the URL is not a valid IP address")
	}

	return ip.String(), nil
}
