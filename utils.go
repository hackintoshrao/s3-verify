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
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represetn a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting into 63 bits.
)

// ExecRequest - Executes an HTTP request creating an HTTP response.
func ExecRequest(req *http.Request, client *http.Client) (*http.Response, error) {
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

// Check whether provided URL is an AWS endpoint.
func isAmazonEndpoint(endpointURL *url.URL) bool {
	if endpointURL == nil {
		return false
	}
	if endpointURL.Host == "s3.amazonaws.com" || endpointURL.Host == "s3.cn-north-1.amazonaws.com.cn" {
		return true
	}
	return false
}

// Check whether endpoint supports virtual style.
func isVirtualStyleHostSupported(endpointURL *url.URL) bool {
	// can return true for Amazon
	return isAmazonEndpoint(endpointURL)
}

// Generate a new URL from the user provided endpoint.
func makeTargetURL(endpoint, bucketName, objectName, region string) (*url.URL, error) {
	targetURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	targetURL.Path = "/"
	if bucketName != "" || objectName != "" {
		targetURL.Path = "/" + bucketName + "/" + objectName // Default to path style.
		if isVirtualStyleHostSupported(targetURL) {          // Virtual style supported, use virtual style.
			targetURL.Path = "/" + objectName
			targetURL.Host = bucketName + "." + getS3Endpoint(region)
		}
	}
	return targetURL, nil
}

// NewS3Client - Generate a new client to perform S3
func NewS3Client(endpoint, access, secret string) (*minio.Client, error) {
	secure := true // Use HTTPS connection.
	hostURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	s3Client, err := minio.New(hostURL.Host, access, secret, secure)
	if err != nil {
		return nil, err
	}
	return s3Client, nil
}

// Remove any s3verify created buckets and objects.
func cleanUpTest(s3Client minio.Client, bucketNames, objectNames []string) error {
	// If there are test object/s to remove, remove it/them first.
	if objectNames != nil {
		for _, objectName := range objectNames {
			// TODO: for now assume if objectNames are given they are all stored in the first passed bucketName.
			err := s3Client.RemoveObject(bucketNames[0], objectName)
			if minio.ToErrorResponse(err).Code != "NoSuchKey" {
				return err
			}

		}
	}
	// Remove any test buckets.
	for _, bucketName := range bucketNames {
		err := s3Client.RemoveBucket(bucketName)
		if err != nil {
			if minio.ToErrorResponse(err).Code != "NoSuchBucket" {
				fmt.Println(err)
				return err
			}
		}

	}
	return nil

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
