# Pipeline Status Resource

A resource for tracking the running state of a Concourse pipeline.

**THIS IS UNDER CONSTRUCTION**


## Source Configuration

* `initial_version`: *Optional.* The version number to use when
bootstrapping, i.e. when there is not a version number present in the source.

* `driver`: *Optional. Currently only `s3` is supported.* The driver to use for tracking the
  version. Determines where the version is stored.

There are three supported drivers, with their own sets of properties for
configuring them.


### `s3` Driver

The `s3` driver works by modifying a file in an S3 compatible bucket.

* `bucket`: *Required.* The name of the bucket.

* `key`: *Required.* The key to use for the object in the bucket tracking
the version.

* `access_key_id`: *Required.* The AWS access key to use when accessing the
bucket.

* `secret_access_key`: *Required.* The AWS secret key to use when accessing
the bucket.

* `region_name`: *Optional. Default `us-east-1`.* The region the bucket is in.

* `endpoint`: *Optional.* Custom endpoint for using S3 compatible provider.

* `disable_ssl`: *Optional.* Disable SSL for the endpoint, useful for S3 compatible providers without SSL.

* `server_side_encryption`: *Optional.* The server-side encryption algorithm
used when storing the version object (e.g. `AES256`, `aws:kms`).

* `use_v2_signing`: *Optional.* Use v2 Signature signing default is false.
