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

var (
	// basicTests holds all tests that MUST pass for a server
	// to claim compatibility with AWS S3 Signature V4.
	basicTests = [][]APItest{
		putBucketTestsBasic,
		putObjectTestsBasic,
		multipartTestsBasic,
		headObjectTestsBasic,
		copyObjectTestsBasic,
		getObjectTestsBasic,
		listBucketsTestsBasic,
		removeObjectTestsBasic,
		removeBucketTestsBasic,
	}
	// basicMessages holds all messages that belong to tests
	// that MUST pass for a server to claim AWS S3 sign v4 compatability.
	basicMessages = [][]string{
		putBucketMessagesBasic,
		putObjectMessagesBasic,
		multipartMessagesBasic,
		headObjectMessagesBasic,
		copyObjectMessagesBasic,
		getObjectMessagesBasic,
		listBucketsMessagesBasic,
		removeObjectMessagesBasic,
		removeBucketMessagesBasic,
	}
	// extendedTests holds all the basic tests that MUST pass
	// as well as some tests for features that are not necessary
	// for compatibility.
	extendedTests = [][]APItest{
		putBucketTestsExtended,
		putObjectTestsExtended,
		multipartTestsExtended,
		headObjectTestsExtended,
		copyObjectTestsExtended,
		getObjectTestsExtended,
		listBucketsTestsExtended,
		removeObjectTestsExtended,
		removeBucketTestsExtended,
	}
	// extendedMessages holds all the messages that belong to
	// both the basic tests as well as those that test features that
	// are not necessary for compatibility.
	extendedMessages = [][]string{
		putBucketMessagesExtended,
		putObjectMessagesExtended,
		multipartMessagesExtended,
		headObjectMessagesExtended,
		copyObjectMessagesExtended,
		getObjectMessagesExtended,
		listBucketsMessagesExtended,
		removeObjectMessagesExtended,
		removeBucketMessagesExtended,
	}
)

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

// Tests and messages for PutBucket API,
// make sure future tests/messages are added here.
var (
	putBucketTestsBasic = []APItest{
		mainPutBucket,
	}
	putBucketMessagesBasic = []string{
		"PutBucket:",
	}
	// PutBucket API has no real extended tests
	putBucketTestsExtended = []APItest{
		mainPutBucket,
	}
	putBucketMessagesExtended = []string{
		"PutBucket:",
	}
)

// Tests and messages for PutObject API,
// make sure future tests/messages are added here.
var (
	putObjectTestsBasic = []APItest{
		mainPutObject,
	}
	putObjectMessagesBasic = []string{
		"PutObject:",
	}
	// For now PutObject API only has basic tests
	// in the future there will be extended tests added.
	putObjectTestsExtended = []APItest{
		mainPutObject,
	}
	putObjectMessagesExtended = []string{
		"PutObject:",
	}
)

// Tests and messages for Multipart API,
// make sure future tests/messages are added here.
var (
	multipartTestsBasic = []APItest{
		mainInitiateMultipartUpload,
		mainUploadPart,
		mainListParts,
		mainCompleteMultipartUpload,
	}
	multipartMessagesBasic = []string{
		"InitiateMultipartUpload:",
		"UploadPart:",
		"ListParts:",
		"CompleteMultipartUpload:",
	}
	multipartTestsExtended = []APItest{
		mainInitiateMultipartUpload,
		mainUploadPart,
		mainListParts,
		mainCompleteMultipartUpload,
	}
	multipartMessagesExtended = []string{
		"InitiateMultipartUpload:",
		"UploadPart:",
		"ListParts:",
		"CompleteMultipartUpload:",
	}
)

// Tests and messages for HeadObject API,
// make sure future tests/messages are added here.
var (
	headObjectTestsBasic = []APItest{
		mainHeadObject,
	}
	headObjectMessagesBasic = []string{
		"HeadObject:",
	}
	// For now HeadObject API only has a basic test
	// in the future there will be extended tests added.
	headObjectTestsExtended = []APItest{
		mainHeadObject,
		mainHeadObjectIfMatch,
		mainHeadObjectIfNoneMatch,
	}
	headObjectMessagesExtended = []string{
		"HeadObject:",
		"HeadObject (If-Match):",
		"HeadObject (If-None-Match):",
	}
)

// Tests and messages for CopyObject API,
// make sure future tests/messages are added here.
var (
	copyObjectTestsBasic = []APItest{
		mainCopyObject,
	}
	copyObjectMessagesBasic = []string{
		"CopyObject:",
	}
	copyObjectTestsExtended = []APItest{
		// When extended tests are run also run the basic.
		mainCopyObject,
		mainCopyObjectIfMatch,
		mainCopyObjectIfNoneMatch,
	}

	copyObjectMessagesExtended = []string{
		"CopyObject:",
		"CopyObject (If-Match):",
		"CopyObject (If-None-Match):",
	}
)

// Tests and messages for GetObject API,
// make sure future tests/messages are added here.
var (
	getObjectTestsBasic = []APItest{
		mainGetObject,
	}
	getObjectMessagesBasic = []string{
		"GetObject:",
	}
	getObjectTestsExtended = []APItest{
		mainGetObject,
		mainGetObjectRange,
		mainGetObjectIfMatch,
		mainGetObjectIfNoneMatch,
		mainGetObjectIfModifiedSince,
		mainGetObjectIfUnModifiedSince,
	}
	getObjectMessagesExtended = []string{
		"GetObject:",
		"GetObject (Range):",
		"GetObject (If-Match):",
		"GetObject (If-None-Match):",
		"GetObject (If-Modified-Since):",
		"GetObject (If-Unmodified-Since):",
	}
)

// Tests and messages for ListBuckets API,
// make sure future tests/messages are added here.
var (
	listBucketsTestsBasic = []APItest{
		mainListBucketsExist,
		// TODO: add test for listing non-existent buckets.
	}
	listBucketsMessagesBasic = []string{
		"ListBuckets (Buckets Exist):",
	}
	listBucketsTestsExtended = []APItest{
		mainListBucketsExist,
	}
	listBucketsMessagesExtended = []string{
		"ListBuckets (Buckets Exist):",
	}
)

// Tests and messages for RemoveObject API,
// make sure future tests/messages are added here.
var (
	removeObjectTestsBasic = []APItest{
		mainRemoveObjectExists,
		// TODO: add test for removeing non-existent object
	}
	removeObjectMessagesBasic = []string{
		"RemoveObject (Object Exists):",
	}
	removeObjectTestsExtended = []APItest{
		mainRemoveObjectExists,
		// TODO: add test for removeing non-existent object
	}
	removeObjectMessagesExtended = []string{
		"RemoveObject (Object Exists):",
	}
)

// Tests and messages for RemoveBucket API,
// make sure future tests/messages are added here.
var (
	removeBucketTestsBasic = []APItest{
		// Both of these are defined as basic tests.
		mainRemoveBucketExists,
		mainRemoveBucketDNE,
	}
	removeBucketMessagesBasic = []string{
		"RemoveBucket (Bucket Exists):",
		"RemoveBucket (Bucket DNE):",
	}
	removeBucketTestsExtended = []APItest{
		mainRemoveBucketExists,
		mainRemoveBucketDNE,
	}
	removeBucketMessagesExtended = []string{
		"RemoveBucket (Bucket Exists):",
		"RemoveBucket (Bucket DNE):",
	}
)
