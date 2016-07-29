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
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/minio/mc/pkg/console"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represetn a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting into 63 bits.
)

// List of success status.
var successStatus = []int{
	http.StatusOK,
	http.StatusNoContent,
	http.StatusPartialContent,
}

// printMessage - Print test pass/fail messages with errors.
func printMessage(message string, err error) {
	// Erase the old progress line.
	console.Eraseline()
	if err != nil {
		message += strings.Repeat(" ", messageWidth-len([]rune(message))) + "[FAIL]\n" + err.Error()
		console.Println(message)
	} else {
		message += strings.Repeat(" ", messageWidth-len([]rune(message))) + "[OK]"
		console.Println(message)
	}
}

// verifyHostReachable - Execute a simple get request against the provided endpoint to make sure its reachable.
func verifyHostReachable(endpoint, region string) error {
	targetURL, err := makeTargetURL(endpoint, "", "", region, nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		// Only give server 3 seconds to complete the request.
		Timeout: 3000 * time.Millisecond,
	}
	req := &http.Request{
		Method: "GET",
		URL:    targetURL,
	}
	if _, err := client.Do(req); err != nil {
		return err
	}
	return nil
}

// xmlDecoder provide decoded value in xml.
func xmlDecoder(body io.Reader, v interface{}) error {
	d := xml.NewDecoder(body)
	return d.Decode(v)
}

// execRequest - Executes an HTTP request creating an HTTP response and implements retry logic for predefined retryable errors.
func execRequest(req *http.Request, client *http.Client, bucketName, objectName string) (resp *http.Response, err error) {
	var isRetryable bool         // Indicates if request can be retried.
	var bodyReader io.ReadSeeker // io.Seeking for seeking.
	if req.Body != nil {
		// FIXME: remove this and reduce ioutil.NopCloser usage elsewhere.
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		isRetryable = true
		bodyReader = bytes.NewReader(buf)
	}
	// Do not need the index.
	for _ = range newRetryTimer(MaxRetry, time.Second, time.Second*30, MaxJitter, globalRandom) {
		if isRetryable {
			// Seek back to beginning for each attempt.
			if _, err := bodyReader.Seek(0, 0); err != nil {
				// If seek failed no need to retry.
				return resp, err
			}
		}
		if bodyReader != nil {
			req.Body = ioutil.NopCloser(bodyReader)
		}
		resp, err = client.Do(req)
		if err != nil {
			// For supported network errors verify.
			if isNetErrorRetryable(err) {
				continue // Retry.
			}
			// For other errors there is no need to retry.
			return resp, err
		}
		// For any known successful http status, return quickly.
		for _, httpStatus := range successStatus {
			if httpStatus == resp.StatusCode {
				return resp, nil
			}
		}
		// Read the body to be saved later.
		errBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		// Save the body.
		errBodySeeker := bytes.NewReader(errBodyBytes)
		resp.Body = ioutil.NopCloser(errBodySeeker)

		// For errors verify if its retryable otherwise fail quickly.
		errResponse := ToErrorResponse(httpRespToErrorResponse(resp, bucketName, objectName))

		//Verify if error response code is retryable.
		if isS3CodeRetryable(errResponse.Code) {
			continue // Retry.
		}
		// Verify if http status code is retryable.
		if isHTTPStatusRetryable(resp.StatusCode) {
			continue // Retry.
		}

		// Save the body back again.
		errBodySeeker.Seek(0, 0) // Seek back to starting point.
		resp.Body = ioutil.NopCloser(errBodySeeker)

		// For all other cases break out of the retry loop.
		break
	}
	return resp, err
}

// closeResponse close non nil response with any response Body.
// convenient wrapper to drain any remaining data on response body.
//
// Subsequently this allows golang http RoundTripper
// to re-use the same connection for future requests. (Connection pooling).
func closeResponse(res *http.Response) {
	// Callers should close resp.Body when done reading from it.
	// If resp.Body is not closed, the Client's underlying RoundTripper
	// (typically Transport) may not be able to re-use a persistent TCP
	// connection to the server for a subsequent "keep-alive" request.
	if res != nil && res.Body != nil {
		// Drain any remaining Body and then close the connection.
		// Without this closing connection would disallow re-using
		// the same connection for future uses.
		// - http://stackoverflow.com/a/17961593/4465767
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}
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
func makeTargetURL(endpoint, bucketName, objectName, region string, queryValues url.Values) (*url.URL, error) {
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
	if len(queryValues) > 0 { // If there are query values include them.
		targetURL.RawQuery = queryValues.Encode()
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
