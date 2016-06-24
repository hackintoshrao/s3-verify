/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represetn a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting into 63 bits.
)

// ExecRequest - Executes an HTTP request creating an HTTP response.
func ExecRequest(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// randString generates random names.
func randString(n int, src rand.Source, prefix string) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return prefix + string(b[0:30-len(prefix)])
}

// TODO: Check these functions.

// isValidAccessKey - Verify that the provided access key is valid.
func isValidAccessKey(accessKey string) bool {
	if accessKey == "" {
		return true
	}
	regex := regexp.MustCompile(`.{5,40}$`)
	return regex.MatchString(accessKey) && !strings.ContainsAny(accessKey, "$%^~`!|&*#@")
}

// isValidSecretKey - Verify that the provided secret key is valid.
func isValidSecretKey(secretKey string) bool {
	if secretKey == "" {
		return true
	}
	regex := regexp.MustCompile(`.{8,40}$`)
	return regex.MatchString(secretKey) && !strings.ContainsAny(secretKey, "$%^~`!|&*#@")
}

// isValidHostURL - Verify that the provided host is valid.
func isValidHostURL(urlStr string) bool {
	if strings.TrimSpace(urlStr) == "" {
		return false
	}
	targetURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	if targetURL.Scheme != "https" && targetURL.Scheme != "http" {
		return false
	}
	return true
}

func isAmazonEndpoint(endpointURL *url.URL) bool {
	if endpointURL == nil {
		return false
	}
	if endpointURL.Host == "s3.amazonaws.com" || endpointURL.Host == "s3.cn-north-1.amazonaws.com.cn" {
		return true
	}
	return false
}

func isGoogleEndpoint(endpointURL *url.URL) bool {
	if endpointURL == nil {
		return false
	}
	if endpointURL.Host == "storage.googleapis.com" {
		return true
	}
	return false
}

func isVirtualStyleHostSupported(endpointURL *url.URL) bool {
	// can return true for all other cases.
	return isAmazonEndpoint(endpointURL) || isGoogleEndpoint(endpointURL)
}
