# Updates

1. rename manifest, there are two types, one with prefix test- is for push test, one with prefix ref- is for discovery test
2. deletion tests in management test should fail if resources can't be deleted

# New
## push

1. log specific message (old change)
2. create manifest with empty config and test pushing
3. create manifest with subject and test pushing 
4. test pushing 4 MB blob and manifest with 4 MB blob
5. create manifest with different layer blob media type: 
        "application/vnd.oci.image.layer.v1.tar",
		"application/vnd.oci.image.layer.v1.tar+gzip",
		"application/vnd.oci.image.layer.nondistributable.v1.tar",
		"application/vnd.oci.image.layer.nondistributable.v1.tar+gzip",
		"application/vnd.oci.image.layer.v1.tar+zstd",
    and test pushing
6. create manifest with custom layer blob media type and test pushing
7. create manifest with empty media type but with specified artifact type 
8. create manifest index and test pushing
9. create manifest index containing index and test pushing

## pull
1. pull test for head request with Range Header 

## discovery
1. manifest fetched by referrers api should contain annotations of original manifest

# Delete
1. delete noise from manifests, make sure each manifest is only for testing a single feature

