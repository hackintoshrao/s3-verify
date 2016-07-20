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

// Tests must be run in the following order
// PutBucket,
// PutObject,
// Multipart,
// HeadObject,
// CopyObject,
// GetObject,
// ListBuckets,
// RemoveObject,
// RemoveBucket,
var apiTests = []APItest{
	// Test for PutBucket API.
	APItest{
		Test:     mainPutBucket,
		Extended: false, // PUT bucket tests must be run even without extended flags being set.
		Critical: true,  // PUT bucket tests must pass before other tests can be run.
	},

	// Test for PutObject API.
	APItest{
		Test:     mainPutObject,
		Extended: false, // PUT object tests must be run even witout extended flags being set.
		Critical: true,  // PUT object tests must pass before other tests can be run.
	},

	// Tests for Multipart API.
	APItest{
		Test:     mainInitiateMultipartUpload,
		Extended: false, // Initiate Multipart test must be run even without extended flags being set.
		Critical: true,  // Initiate Multipart test must pass before other tests can be run.
	},
	APItest{
		Test:     mainUploadPart,
		Extended: false, // Upload Part test must be run even without extended flag being set.
		Critical: true,  // Upload Part test must pass before other tests can be run.
	},
	APItest{
		Test:     mainListParts,
		Extended: false, // List Part test must be run even without extended flag being set.
		Critical: false, // List Part test can fail without affecting other tests.
	},
	APItest{
		Test:     mainCompleteMultipartUpload,
		Extended: false, // Complete Multipart test must be run even without extended flag being set.
		Critical: true,  // Complete Multipart test can fail without affecting other tests.
	},

	// Tests for HeadObject API.
	APItest{
		Test:     mainHeadObject,
		Extended: false, // Head Object test must be run even without extended flags being set.
		Critical: true,  // Head Object test must pass before other tests can be run.
	},
	APItest{
		Test:     mainHeadObjectIfMatch,
		Extended: true,  // Head Object (If-Match) test does not need to be run unless explicitly asked for.
		Critical: false, // Head Object (If-Match) does not need to pass before other tests can be run.
	},
	APItest{
		Test:     mainHeadObjectIfNoneMatch,
		Extended: true,  // Head Object (If-None-Match) test does not need to be run unless explicitly asked for.
		Critical: false, // Head Object (If-None-Match) test does not need pass before other tests can be run.
	},
	APItest{
		Test:     mainHeadObjectIfUnModifiedSince,
		Extended: true,  // Head Object (If-Unmodified-Since) test does not need to be run unless explicitly asked for.
		Critical: false, // Head Object (If-Unmodified-Since) test does not need to pass before other tests can be run.
	},
	APItest{
		Test:     mainHeadObjectIfModifiedSince,
		Extended: true,  // Head Object (If-Modified-Since) test does not need to be run unless explicitly asked for.
		Critical: false, // This test does not need to pass before others are run.
	},

	// Tests for CopyObject.
	APItest{
		Test:     mainCopyObject,
		Extended: false, // Copy Object test must be run even without extended flags being set.
		Critical: false, // Copy Object test can fail and not effect other tests.
	},
	APItest{
		Test:     mainCopyObjectIfMatch,
		Extended: true,  // Copy Object (If-Match) test does not need to be run unless explicitly asked for.
		Critical: false, // Copy Object (If-Match) test can fail and not affect other tests.
	},
	APItest{
		Test:     mainCopyObjectIfNoneMatch,
		Extended: true,  // Copy Object (If-None-Match) test does not need to be run unless explicitly asked for.
		Critical: false, // Copy Object (If-Match) test can fail and not affect other tests.
	},
	APItest{
		Test:     mainCopyObjectIfModifiedSince,
		Extended: true,  // Copy Object (If-Modified-Since) test does not need to be run.
		Critical: false, // Copy Object (If-Modified-Since) can fail and not affect other tests.
	},

	// Tests for GetObject API.
	APItest{
		Test:     mainGetObject,
		Extended: false, // Get Object test must be run.
		Critical: false, // Get Object can fail and not affect other tests.
	},
	APItest{
		Test:     mainGetObjectIfMatch,
		Extended: true,  // Get Object (If-Match) does not need to be run.
		Critical: false, // Get Object (If-Match) can fail and not affect other tests.
	},
	APItest{
		Test:     mainGetObjectIfNoneMatch,
		Extended: true,  // Get Object (If-None-Match) does not need to be run.
		Critical: false, // Get Object (If-None-Match) can fail and not affect other tests.
	},
	APItest{
		Test:     mainGetObjectIfModifiedSince,
		Extended: true,  // Get Object (If-Modified-Since) does not need to be run.
		Critical: false, // Get Object (If-Modified-Since) can fail and not affect other tests.
	},
	APItest{
		Test:     mainGetObjectIfUnModifiedSince,
		Extended: true,  // Get Object (If-Unmodified-Since) does not need to be run.
		Critical: false, // Get Object (If-Unmodified-Since) can fail and not affect other tests.
	},
	APItest{
		Test:     mainGetObjectRange,
		Extended: true,  // Get Object (Range) does not need to be run.
		Critical: false, // Get Object (Range) can fail and not affect other tests.
	},

	// Tests for ListBuckets API.
	APItest{
		Test:     mainListBucketsExist,
		Extended: false, // List Buckets test must be run.
		Critical: false, // List Buckets test can fail and not affect other tests.
	},

	// Test for RemoveObject API.
	APItest{
		Test:     mainRemoveObjectExists,
		Extended: false, // Remove Object test must be run.
		Critical: true,  // Remove Object test must pass for future tests.
	},

	// Tests for RemoveBucket API.
	APItest{
		Test:     mainRemoveBucketDNE,
		Extended: false, // Remove Bucket test DNE must be run.
		Critical: false, // Remove Bucket DNE test can fail and not affect other tests.
	},
	APItest{
		Test:     mainRemoveBucketExists,
		Extended: false, // Remove Bucket test Exists must be run.
		Critical: false, // Remove Bucket test Exists can fail and not affect other tests.
	},
}
