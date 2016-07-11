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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/minio/s3verify/signv4"
)

var (
	// First object to test PUT on.
	putObj1 = &ObjectInfo{
		Key: "s3verify-put-object-test1",
		// LastModified: to be set dynamically,
		// Size: to be set dynamically,
		// ETag: to be set dynamically,
		ContentType: "application/octet-stream",
		Body:        []byte("Nemo enim ipsam voluptatem, quia voluptas sit, aspernature aut odit aut fugit,"),
	}
	// Second object to test PUT on.
	putObj2 = &ObjectInfo{
		Key: "s3verify-put-object-test2",
		// LastModified: to be set dynamically,
		// Size: to be set dynamically,
		// ETag: to be set dynamically,
		ContentType: "application/octet-stream",
		Body:        []byte("sed quia consequuntur magni dolores eos, qui ratione voluptatem sequi nescuint,"),
	}
	// Third object to test PUT on.
	putObj3 = &ObjectInfo{
		Key: "s3verify-put-object-test3",
		// LastModified: to be set dynamically,
		// Size: to be set dynamically,
		// ETag: to be set dynamically,
		ContentType: "application/octet-stream",
		Body:        []byte("neque porro quisquam est, qui dolorem ipsum, quia dolor sit amet, consectetur,"),
	}
)

// Store all regularly PUT objects.
var objects = []*ObjectInfo{
	putObj1,
	putObj2,
	putObj3,
}

// Store all objects that were copied.
var copyObjects = []*ObjectInfo{}

// newPutObjectReq - Create a new HTTP request for PUT object.
func newPutObjectReq(config ServerConfig, bucketName, objectName string, objectData []byte) (*http.Request, error) {
	// An HTTP request for a PUT object.
	var putObjectReq = &http.Request{
		Header: map[string][]string{
		// Set Content SHA dynamically because it is based on data being uploaded.
		// Set Content MD5 dynamically because it is based on data being uploaded.
		// Set Content-Length dynamically because it is based on data being uploaded.
		},
		// Body will be set dynamically.
		// Body:
		Method: "PUT",
	}
	targetURL, err := makeTargetURL(config.Endpoint, bucketName, objectName, config.Region)
	if err != nil {
		return nil, err
	}
	// Fill request headers and URL.
	putObjectReq.URL = targetURL

	// Compute md5Sum and sha256Sum from the input data.
	reader := bytes.NewReader(objectData)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return nil, err
	}
	putObjectReq.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	putObjectReq.Header.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	putObjectReq.ContentLength = contentLength
	// Set the body to the data held in objectData.
	putObjectReq.Body = ioutil.NopCloser(reader)
	putObjectReq = signv4.SignV4(*putObjectReq, config.Access, config.Secret, config.Region)
	return putObjectReq, nil
}

// putObjectVerify - Verify the response matches what is expected.
func putObjectVerify(res *http.Response, expectedStatus string) error {
	if err := verifyHeaderPutObject(res); err != nil {
		return err
	}
	if err := verifyStatusPutObject(res, expectedStatus); err != nil {
		return err
	}
	if err := verifyBodyPutObject(res); err != nil {
		return err
	}
	return nil
}

// verifyStatusPutObject - Verify that the res status code matches what is expected.
func verifyStatusPutObject(res *http.Response, expectedStatus string) error {
	if res.Status != expectedStatus {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatus, res.Status)
		return err
	}
	return nil
}

// verifyBodyPutObject - Verify that the body returned matches what is uploaded.
func verifyBodyPutObject(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	// A PUT request should give back an empty body.
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Recieved: expected empty body but recieved: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderPutObject - Verify that the header returned matches waht is expected.
func verifyHeaderPutObject(res *http.Response) error {
	if err := verifyStandardHeaders(res); err != nil {
		return err
	}
	return nil
}

// Test a PUT object request with no special headers set. This adds one object to each of the test buckets.
func mainPutObject(config ServerConfig, message string) error {
	// TODO: create tests designed to fail.
	for _, object := range objects {
		bucket := testBuckets[0]
		// Spin scanBar
		scanBar(message)
		// PUT each object in each available bucket.
		// Generate a new PUT object HTTP req.
		req, err := newPutObjectReq(config, bucket.Name, object.Key, object.Body)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Execute the request.
		res, err := execRequest(req, config.Client)
		if err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
		// Verify the response.
		if err := putObjectVerify(res, "200 OK"); err != nil {
			return err
		}
		// Spin scanBar
		scanBar(message)
	}
	return nil
}
