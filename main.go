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
	"fmt"
	"os"
)

type TestConfig struct {
	access   string
	secret   string
	endpoint string
}

func main() {
	config := TestConfig{
		access:   os.Getenv("ACCESS_KEY"),
		secret:   os.Getenv("SECRET_KEY"),
		endpoint: os.Getenv("END_POINT"),
	}
	var verifyOps []S3Verify
	// Adding MakeBucket to list of verifiers.
	makeBucket := MakeBucket{BucketName: "random-xyz-abc"}
	verifyOps = append(verifyOps, makeBucket)
	// Start Verification.
	VerifyS3Compat(verifyOps, config)
}

func VerifyS3Compat(verifyOps []S3Verify, config TestConfig) error {

	for _, s3Verify := range verifyOps {
		err := execTest(s3Verify, config)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}
	return nil
}

func execTest(s3Verify S3Verify, config TestConfig) error {
	req, err := s3Verify.MakePlainRequest(config.endpoint)
	if err != nil {
		return err
	}
	// setting the request Headers.
	req = s3Verify.SetHeaders(req)
	// setting the request body.
	req = s3Verify.SetBody(req)
	// Getting the v4 Sign on the request.
	req = s3Verify.SignRequest(req, config.access, config.secret)

	resp, err := s3Verify.ExecRequest(req)
	if err != nil {
		return err
	}
	if err = s3Verify.VerifyResponse(resp); err != nil {
		return err
	}
	return nil
}
