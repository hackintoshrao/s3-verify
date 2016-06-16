/*
 * Minio Go Library for Amazon S3 Compatible Cloud Storage (C) 2015, 2016 Minio, Inc.
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
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	"github.com/minio/s3-verify/signv4"
)

func GetMakeBucketHeader() map[string]string {
	return map[string][]string{
		"X-Amz-Content-Sha256": {hex.EncodeToString(signv4.Sum256([]byte{}))},
	}
}

func GetMakeBucketReqType() string {
	return "PUT"
}

func GetMakeBucketBody(location string) io.ReadSeeker {
	switch location {
	case "us-east-1":
		return nil
	default:
		// TODO: Implement body generation In case of other locations.
		return nil
	}
}

// Construct URLPath for MakeBucket.
func GetMakeBucketURL(endPoint string) (string, error) {
	makeBucketPath := func(bucketName string) string {
		return "/" + bucketName + "/"
	}
	targetURL, err := url.Parse(endPoint)
	if err != nil {
		return nil, err
	}
	targetURL.Path = makeBucketPath(bucketName)
	return targetURL.String(), nil
}

// Construct Http request with URL Path for MakeBucket.
func NewMakeBucketReq(endPointStr string) (*http.Request, error) {
	// returns path for creating the bucket.
	// Parse parses rawurl into a URL structure.
	targetURLStr, err := GetMakeBucketURL(endPoint)
	if err != nil {
		return nil, err
	}
	req, err := InitHttpRequest(GetMakeBucketReqType(), targetURLStr, GetMakeBucketBody())
	if err != nil {
		return nil, err
	}
	return req, nil
}

func StartMakeBucketTest() error {
	reqMakeBucket, err := NewMakeBucketReq(endPointStr)
	if err != nil {
		return err
	}
	reqMakeBucket = SignRequestV4(reqMakeBucket, accessKey, SecretKey, Region)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if err = verifyResponse(resp); err != nil {
		return err
	}
	return nil
}

// Verify the HTTP response.
func VerifyResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		fmt.Println("=======Make Bucket Failed=====")
		return nil
	}
	fmt.Println("=======Make Bucket Passed=====")
	return nil
}
