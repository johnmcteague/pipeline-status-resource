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
