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
	"crypto/md5"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represetn a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting into 63 bits.
)

// xmlDecoder provide decoded value in xml.
func xmlDecoder(body io.Reader, v interface{}) error {
	d := xml.NewDecoder(body)
	return d.Decode(v)
}

// execRequest - Executes an HTTP request creating an HTTP response.
func execRequest(req *http.Request, client *http.Client) (*http.Response, error) {
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

// Check if the endpoint is for an AWS S3 server.
func isAmazonEndpoint(endpointURL *url.URL) bool {
	if endpointURL == nil {
		return false
	}
	if endpointURL.Host == "s3.amazonaws.com" || endpointURL.Host == "s3.cn-north-1.amazonaws.com.cn" {
		return true
	}
	return false
}

// Generate a new URL from the user provided endpoint.
func makeTargetURL(endpoint, bucketName, objectName, region string) (*url.URL, error) {
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if isAmazonEndpoint(targetURL) { // Change host to reflect the region.
		targetURL.Host = getS3Endpoint(region)
	}
	targetURL.Path = "/"
	if bucketName != "" {
		targetURL.Path = "/" + bucketName + "/" + objectName // Use path style requests only.
	}
	return targetURL, nil
}

// Verify the date field of an HTTP response is formatted with HTTP time format.
func verifyDate(respDateStr string) error {
	_, err := time.Parse(http.TimeFormat, respDateStr)
	if err != nil {
		err = fmt.Errorf("Invalid time format recieved, expected http.TimeFormat")
		return err
	}
	return nil
}

// Verify all standard headers in an HTTP response.
func verifyStandardHeaders(res *http.Response) error {
	// Check the date header.
	respDateStr := res.Header.Get("Date")
	if err := verifyDate(respDateStr); err != nil {
		return err
	}
	return nil
}

// Generate MD5 and SHA256 for an input readseeker.
func computeHash(reader io.ReadSeeker) (md5Sum, sha256Sum []byte, contentLength int64, err error) {
	// MD5 and SHA256 hasher.
	var hashMD5, hashSHA256 hash.Hash
	// MD5 and SHA256 hasher.
	hashMD5 = md5.New()
	hashSHA256 = sha256.New()
	hashWriter := io.MultiWriter(hashMD5, hashSHA256)

	// If no buffer is provided, no need to allocate just use io.Copy
	contentLength, err = io.Copy(hashWriter, reader)
	if err != nil {
		return nil, nil, 0, err
	}
	// Seek back to beginning location.
	if _, err := reader.Seek(0, 0); err != nil {
		return nil, nil, 0, err
	}
	// Finalize md5sum and sha256sum.
	md5Sum = hashMD5.Sum(nil)
	sha256Sum = hashSHA256.Sum(nil)

	return md5Sum, sha256Sum, contentLength, nil
}
