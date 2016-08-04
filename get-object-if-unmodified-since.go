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
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// newGetObjectIfUnModifiedSinceReq - Create a new HTTP GET request with the If-Unmodified-Since header set to perform.
func newGetObjectIfUnModifiedSinceReq(config ServerConfig, bucketName, objectName string, lastModified time.Time) (Request, error) {
	// An HTTP GET request with the If-Unmodified-Since header set.
	var getObjectIfUnModifiedSinceReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	getObjectIfUnModifiedSinceReq.bucketName = bucketName
	getObjectIfUnModifiedSinceReq.objectName = objectName

	reader := bytes.NewReader([]byte{}) // Compute hash using empty body because GET requests do not send a body.
	_, sha256Sum, _, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	// Set the headers.
	getObjectIfUnModifiedSinceReq.customHeader.Set("If-Unmodified-Since", lastModified.Format(http.TimeFormat))
	getObjectIfUnModifiedSinceReq.customHeader.Set("User-Agent", appUserAgent)
	getObjectIfUnModifiedSinceReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))

	return getObjectIfUnModifiedSinceReq, nil
}

// verifyGetObjectIfUnModifiedSince - Verify the response matches what is expected.
func verifyGetObjectIfUnModifiedSince(res *http.Response, expectedBody []byte, expectedStatusCode int, shouldFail bool) error {
	if err := verifyBodyGetObjectIfUnModifiedSince(res.Body, expectedBody, shouldFail); err != nil {
		return err
	}
	if err := verifyStatusGetObjectIfUnModifiedSince(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyHeaderGetObjectIfUnModifiedSince(res.Header); err != nil {
		return err
	}
	return nil
}

// verifyGetObjectIfUnModifiedSinceBody - Verify that the response body matches what is expected.
func verifyBodyGetObjectIfUnModifiedSince(resBody io.Reader, expectedBody []byte, shouldFail bool) error {
	if shouldFail {
		// Decode the supposed error response.
		errBody := ErrorResponse{}
		err := xmlDecoder(resBody, &errBody)
		if err != nil {
			return err
		}
		if errBody.Code != "PreconditionFailed" {
			err := fmt.Errorf("Unexpected Error Response: wanted PreconditionFailed, got %v", errBody.Code)
			return err
		}
	} else {
		// The body should be returned in full.
		body, err := ioutil.ReadAll(resBody)
		if err != nil {
			return err
		}
		if !bytes.Equal(body, expectedBody) {
			err := fmt.Errorf("Unexpected Body Received: wanted %v, got %v", string(expectedBody), string(body))
			return err
		}
	}
	// Otherwise test failed / passed as expected.
	return nil
}

// verifyStatusGetObjectIfUnModifiedSince - Verify that the response status matches what is expected.
func verifyStatusGetObjectIfUnModifiedSince(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyHeaderGetObjectIfUnModifiedSince - Verify that the header returned matches what is expected.
func verifyHeaderGetObjectIfUnModifiedSince(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// Test the GET object API with the If-Unmodified-Since header set.
func mainGetObjectIfUnModifiedSince(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] GetObject (If-Unmodified-Since):", curTest, globalTotalNumTest)
	// Spin scanBar
	scanBar(message)
	// Set up past date.
	pastDate, err := time.Parse(http.TimeFormat, "Thu, 01 Jan 1970 00:00:00 GMT")
	if err != nil {
		printMessage(message, err)
		return false
	}
	errCh := make(chan error, globalTotalNumTest)
	bucket := validBuckets[0]
	for _, object := range objects {
		// Spin scanBar
		scanBar(message)
		go func(objectKey string, objectLastModified time.Time, objectBody []byte) {
			// Form a request with a pastDate to make sure the object is not returned.
			req, err := newGetObjectIfUnModifiedSinceReq(config, bucket.Name, objectKey, pastDate)
			if err != nil {
				errCh <- err
				return
			}
			// Execute the request.
			res, err := config.execRequest("GET", req)
			if err != nil {
				errCh <- err
				return
			}
			defer closeResponse(res)
			// Verify that the response returns an error.
			if err := verifyGetObjectIfUnModifiedSince(res, []byte(""), http.StatusPreconditionFailed, true); err != nil {
				errCh <- err
				return
			}
			// Form a request with a date in the past.
			goodReq, err := newGetObjectIfUnModifiedSinceReq(config, bucket.Name, objectKey, objectLastModified)
			if err != nil {
				errCh <- err
				return
			}
			// Execute current request.
			goodRes, err := config.execRequest("GET", goodReq)
			if err != nil {
				errCh <- err
				return
			}
			defer closeResponse(goodRes)
			// Verify that the lastModified date in a request returns the object.
			if err := verifyGetObjectIfUnModifiedSince(goodRes, objectBody, http.StatusOK, false); err != nil {
				errCh <- err
				return
			}
			errCh <- nil
		}(object.Key, object.LastModified, object.Body)
		// Spin scanBar
		scanBar(message)
	}
	count := len(objects)
	for count > 0 {
		count--
		// Spin scanBar
		scanBar(message)
		err, ok := <-errCh
		if !ok {
			return false
		}
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
