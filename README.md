# S3Verify
S3 Verify is a project aimed at verifying the S3 compatibility of AWS V4 signature based object storage systems.

## INSTALLATION
### Prerequisites
- A working Golang Environment. If you do not have a working Golang environment, please follow [Install Golang](https://github.com/minio/minio/blob/master/INSTALLGO.md).
- [Minio Server] (https://github.com/minio/minio/blob/master/README.md) (optional--only necessary if you wish to test your own Minio Server for S3 Compatibility.)

### From Source
Currently s3verify is only available to be downloaded from source. 

```$ go get -u github.com/minio/s3verify```

## APIs
Currently S3 Verify is under heavy development and is subject to breaking changes. Currently we support four different API checks:
* PUT Bucket (Make Bucket)
* GET Service (List Buckets)
* DELETE Bucket (Remove Bucket)
* GET Object (Get Object)

We look forward to adding support for all standard AWS APIs for both Bucket and Object based operations.

## CLI USAGE
Currently s3verify supports two different working modes. If s3verify is run with no commands, but is correctly set up with environment variables or flags it will test all currently supported APIs. Otherwise s3verify can run one API test at a time, with all required environment variables or flags set.
```
$ s3verify [COMMAND...] [FLAGS]
```

### Flags
``s3verify`` implements the following flags
```
    --help      -h      Provides documentation for a given command.
    --access    -a      Allows user to input their AWS access key.
    --secretkey -s      Allows user to input their AWS secret access key.
    --url       -u      Allows user to input the host URL of the server they wish to test.
```

## EXAMPLES
Use s3verify to check the AWS S3 V4 compatibility of the Minio test server (https://play.minio.io:9000) with respect to the Make Bucket, Remove Bucket, List Buckets, and Get Object APIs.
```
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io
```
Use s3verify to check the AWS S3 V4 compatibility of the Minio test server (https://play.minio.io:9000) with respect to only the listbuckets API. 
```
$ s3verify -a YOUR_ACCESS_KEY -s YOUR_SECRET_KEY https://play.minio.io listbuckets
```

