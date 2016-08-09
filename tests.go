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

// Tests must be run in the following order:
// PutBucket,
// PutObject,
// ListBuckets,
// ListObjects,
// Multipart,
// HeadBucket,
// HeadObject,
// CopyObject,
// GetObject,
// RemoveObject,
// RemoveBucket,

// Tests are sorted into the following lists:
// preparedTests    -- tests that will use materials set up by the --prepare flag.
// unpreparedTests  -- tests that will be self-sufficient and create their own testing environment.

// preparedTests - holds all tests that must be run differently based on usage of the --prepared flag.
var preparedTests = []APItest{
	// Tests for PutBucket API.
	APItest{
		Test:     mainPutBucketPrepared,
		Extended: false, // PutBucket is not an extended API.
		Critical: false, // Because --prepared has been used this bucket is not necessary for future tests.
	},
	APItest{
		Test:     mainPutBucketInvalid,
		Extended: false, // PutBucket is not an extended API.
		Critical: false, // This test is not used for future tests.
	},

	// Tests for PutObject API.
	APItest{
		Test:     mainPutObjectPrepared,
		Extended: false, // PutObject is not an extended API.
		Critical: false, // Because --prepared has been used this object is not necessary for future tests.
	},
	APItest{
		Test:     mainPresignedPutObjectPrepared,
		Extended: false, // PutObject presigned is not an extended API.
		Critical: false, // This object is not needed for future tests.
	},

	// Tests for ListBuckets API.
	APItest{
		Test:     mainListBucketsPrepared,
		Extended: false, // ListBuckets is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for ListObjects API.
	APItest{
		Test:     mainListObjectsV1,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainListObjectsV2,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for Multipart API.
	APItest{
		Test:     mainInitiateMultipartUploadPrepared,
		Extended: false, // Initiate Multipart test must be run even without extended flags being set.
		Critical: true,  // Initiate Multipart test must pass before other tests can be run.
	},
	APItest{
		Test:     mainUploadPartPrepared,
		Extended: false, // Upload Part test must be run even without extended flag being set.
		Critical: true,  // Upload Part test must pass before other tests can be run.
	},
	APItest{
		Test:     mainListPartsPrepared,
		Extended: false, // List Part test must be run even without extended flag being set.
		Critical: false, // List Part test can fail without affecting other tests.
	},
	APItest{
		Test:     mainListMultipartUploadsPrepared,
		Extended: false, // List Multipart Uploads test must be run without extended flag being set.
		Critical: false, // List Multipart Uploads test can fail without affecting other tests.
	},
	APItest{
		Test:     mainCompleteMultipartUploadPrepared,
		Extended: false, // Complete Multipart test must be run even without extended flag being set.
		Critical: true,  // Complete Multipart test can fail without affecting other tests.
	},
	APItest{
		Test:     mainAbortMultipartUploadPrepared,
		Extended: false, // Abort Multipart test must be run even without extended flag being set.
		Critical: false, // Abort Multipart test can fail without affecting other tests.
	},

	// Tests for HeadBucket API.
	APItest{
		Test:     mainHeadBucketPrepared,
		Extended: false, // HeadBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for HeadObject API.
	APItest{
		Test:     mainHeadObjectPrepared,
		Extended: false, // HeadObject is not an extended API.
		Critical: true,  // This test affects future tests and must pass.
	},
	APItest{
		Test:     mainHeadObjectIfModifiedSincePrepared,
		Extended: true,  // HeadObject Preparedwith if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfUnModifiedSincePrepared,
		Extended: true,  // HeadObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfMatchPrepared,
		Extended: true,  // HeadObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfNoneMatchPrepared,
		Extended: true,  // HeadObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for CopyObject API.
	APItest{
		Test:     mainCopyObjectPrepared,
		Extended: false, // CopyObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfModifiedSincePrepared,
		Extended: true,  // CopyObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfUnModifiedSincePrepared,
		Extended: true,  // CopyObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfMatchPrepared,
		Extended: true,  // CopyObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfNoneMatchPrepared,
		Extended: true,  // CopyObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for GetObject API.
	APItest{
		Test:     mainGetObjectPrepared,
		Extended: false, // GetObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectPresignedPrepared,
		Extended: false, // GetObject Presigned is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	APItest{
		Test:     mainGetObjectIfModifiedSincePrepared,
		Extended: true,  // GetObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfUnModifiedSincePrepared,
		Extended: true,  // GetObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfMatchPrepared,
		Extended: true,  // GetObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfNoneMatchPrepared,
		Extended: true,  // GetObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectRangePrepared,
		Extended: true,  // GetObject with range header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Test for RemoveObject API.
	APItest{
		Test:     mainRemoveObjectExistsPrepared,
		Extended: false, // RemoveObject is not an extended API.
		Critical: true,  // This test does affect future tests.
	},

	// Tests for RemoveBucket API.
	APItest{
		Test:     mainRemoveBucketExistsPrepared,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: true,  // Removing this bucket is necessary for a good test.
	},
	APItest{
		Test:     mainRemoveBucketDNE,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
}

// unpreparedTests - holds all tests that must be run differently based on usage of the --prepared flag.
var unpreparedTests = []APItest{
	// Tests for PutBucket API.
	APItest{
		Test:     mainPutBucketUnPrepared,
		Extended: false, // PutBucket is not an extended API.
		Critical: true,  // This test does affect future tests.
	},
	APItest{
		Test:     mainPutBucketInvalid,
		Extended: false, // PutBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for PutObject API.
	APItest{
		Test:     mainPutObjectUnPrepared,
		Extended: false, // PutObject is not an extended API.
		Critical: true,  // These objects are necessary for future tests.
	},
	APItest{
		Test:     mainPresignedPutObjectUnPrepared,
		Extended: false, // PutObject presigned is not an extended API.
		Critical: false, // This test is not neccessay for future tests.
	},

	// Tests for ListBuckets API.
	APItest{
		Test:     mainListBucketsUnPrepared,
		Extended: false, // ListBuckets is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for ListObjects API.
	APItest{
		Test:     mainListObjectsV1,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainListObjectsV2,
		Extended: false, // ListObjects is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for Multipart API.
	APItest{
		Test:     mainInitiateMultipartUploadUnPrepared,
		Extended: false, // Initiate Multipart test must be run even without extended flags being set.
		Critical: true,  // Initiate Multipart test must pass before other tests can be run.
	},
	APItest{
		Test:     mainUploadPartUnPrepared,
		Extended: false, // Upload Part test must be run even without extended flag being set.
		Critical: true,  // Upload Part test must pass before other tests can be run.
	},
	APItest{
		Test:     mainListPartsUnPrepared,
		Extended: false, // List Part test must be run even without extended flag being set.
		Critical: false, // List Part test can fail without affecting other tests.
	},
	APItest{
		Test:     mainListMultipartUploadsUnPrepared,
		Extended: false, // List Multipart Uploads test must be run without extended flag being set.
		Critical: false, // List Multipart Uploads test can fail without affecting other tests.
	},
	APItest{
		Test:     mainCompleteMultipartUploadUnPrepared,
		Extended: false, // Complete Multipart test must be run even without extended flag being set.
		Critical: true,  // Complete Multipart test can fail without affecting other tests.
	},
	APItest{
		Test:     mainAbortMultipartUploadUnPrepared,
		Extended: false, // Abort Multipart test must be run even without extended flag being set.
		Critical: false, // Abort Multipart test can fail without affecting other tests.
	},

	// Tests for HeadBucket API.
	APItest{
		Test:     mainHeadBucketUnPrepared,
		Extended: false, // HeadBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for HeadObject API.
	APItest{
		Test:     mainHeadObjectUnPrepared,
		Extended: false, // HeadObject is not an extended API.
		Critical: true,  // This test affects future tests and must pass.
	},
	APItest{
		Test:     mainHeadObjectIfModifiedSinceUnPrepared,
		Extended: true,  // HeadObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfUnModifiedSinceUnPrepared,
		Extended: true,  // HeadObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfMatchUnPrepared,
		Extended: true,  // HeadObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainHeadObjectIfNoneMatchUnPrepared,
		Extended: true,  // HeadObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for CopyObject API.
	APItest{
		Test:     mainCopyObjectUnPrepared,
		Extended: false, // CopyObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfModifiedSinceUnPrepared,
		Extended: true,  // CopyObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfUnModifiedSinceUnPrepared,
		Extended: true,  // CopyObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfMatchUnPrepared,
		Extended: true,  // CopyObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainCopyObjectIfNoneMatchUnPrepared,
		Extended: true,  // CopyObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Tests for GetObject API.
	APItest{
		Test:     mainGetObjectUnPrepared,
		Extended: false, // GetObject is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectPresignedUnPrepared,
		Extended: false, // GetObject Presigned is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfModifiedSinceUnPrepared,
		Extended: true,  // GetObject with if-modified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfUnModifiedSinceUnPrepared,
		Extended: true,  // GetObject with if-unmodified-since header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfMatchUnPrepared,
		Extended: true,  // GetObject with if-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectIfNoneMatchUnPrepared,
		Extended: true,  // GetObject with if-none-match header is an extended API.
		Critical: false, // This test does not affect future tests.
	},
	APItest{
		Test:     mainGetObjectRangeUnPrepared,
		Extended: true,  // GetObject with range header is an extended API.
		Critical: false, // This test does not affect future tests.
	},

	// Test for RemoveObject API.
	APItest{
		Test:     mainRemoveObjectExistsUnPrepared,
		Extended: false, // Remove Object test must be run.
		Critical: true,  // Remove Object test must pass for future tests.
	},

	// Tests for RemoveBucket API.
	APItest{
		Test:     mainRemoveBucketExistsUnPrepared,
		Extended: false, // RemoveBucket is not an extended API.
		Critical: false, // This test does not affect future tests.
	},
}
