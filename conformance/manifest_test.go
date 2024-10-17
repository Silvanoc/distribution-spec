/* OCI image registry compatibility tests
 *
 * Copyright (c) Siemens AG, 2024
 *
 * Authors:
 *  Tobias Schaffner <tobias.schaffner@siemens.com>
 *  Silvano Cirujano Cuesta <silvano.cirujano-cuesta@siemens.com>
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package conformance

import (
	"context"
	"encoding/json"
	"os"

	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"github.com/regclient/regclient"
	od "github.com/regclient/regclient/types/descriptor"
	om "github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	v1 "github.com/regclient/regclient/types/oci/v1"
	"github.com/regclient/regclient/types/ref"
)

func manifestPutOCI(rc *regclient.RegClient, ref ref.Ref, m v1.Manifest) (om.Manifest, error) {
	ctx := context.Background()

	manifest, err := om.New(om.WithOrig(m))
	if err != nil {
		return manifest, err
	}

	return manifest, rc.ManifestPut(ctx, ref, manifest)
}

func blobPut(rc *regclient.RegClient, ref ref.Ref, path string) error {
	ctx := context.Background()

	raw, err := os.Open(path)
	if err != nil {
		return err
	}

	_, err = rc.BlobPut(ctx, ref, od.Descriptor{}, raw)
	return err
}

func getTestManifest() v1.Manifest {
	m := v1.Manifest{
		Config: od.Descriptor{
			MediaType: mediatype.OCI1ImageConfig,
			Digest:    ignoreError(digest.Parse("sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a")),
			Size:      2,
		},
		Layers: []od.Descriptor{
			{
				MediaType: mediatype.OCI1LayerGzip,
				Digest:    ignoreError(digest.Parse("sha256:e63246ad2bce533bcfc8cdfcbc936eba500552aa49ff4527204b4c36d99c3e98")),
				Size:      69,
			},
		},
	}

	m.SchemaVersion = 2

	return m
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// mediaType [...] This property SHOULD be used [...]
// Therefore not specifying this property MUST be supported.
var testNoManifestMediaType = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest without any mediaType", func() {
				// Create manifest
				m := getTestManifest()

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.manifest.v1+json [...]
var testDefaultMediaType = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with default mediaType (application/vnd.oci.image.manifest.v1+json)", func() {
				// Create manifest
				m := getTestManifest()
				m.MediaType = mediatype.OCI1Manifest

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// config/mediaType [...] Implementations MUST support at least the following media types: application/vnd.oci.image.config.v1+json [...]
var testDefaultConfigType = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with default config/mediaType (application/vnd.oci.image.config.v1+json)", func() {
				// Create manifest
				m := getTestManifest()
				m.MediaType = mediatype.OCI1Manifest
				m.Config.MediaType = mediatype.OCI1ImageConfig

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// artifactType [...] This MUST be set when config.mediaType is set to the empty value [...]
var testEmptyConfigFileAndArtifactType = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with custom artifactType and empty config/mediaType", func() {
				// Create manifest
				m := getTestManifest()
				m.MediaType = mediatype.OCI1Manifest
				m.ArtifactType = "application/my-artifact"
				m.Config.MediaType = mediatype.OCI1Empty

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// config/mediaType [...] MUST NOT error on encountering a value that is unknown to the implementation [...]
var testArtifactTypeOverConfigType = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with custom config/mediaType (representing the artifact type)", func() {
				// Create manifest
				m := getTestManifest()
				m.MediaType = mediatype.OCI1Manifest
				m.Config.MediaType = "application/my-artifact-legacy"

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// layers/mediaType [...] MUST NOT error on encountering a mediaType that is unknown to the implementation [...]
var testBlobMediaType = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with custom layer/mediaType (representing the blob type)", func() {
				// Create manifest
				m := getTestManifest()
				m.MediaType = mediatype.OCI1Manifest
				m.Layers[0].MediaType = "application/my-blob-format"

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())
			})
		})
	})
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// mediaType [...] when used, this field MUST contain [...] application/vnd.oci.image.manifest.v1+json [...]
var testWrongManifestMediaTypeFails = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with invalid mediaType", func() {
				// Create manifest
				m := getTestManifest()
				m.MediaType = "application/wrong.type+json"

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(MatchError("manifest contains an unexpected media type: expected application/vnd.oci.image.manifest.v1+json, received application/wrong.type+json"))
			})
		})
	})
}

// OCI Image Specification - Manifest -> https://github.com/opencontainers/image-spec/blob/v1.1.0/manifest.md
// Specification says:
// subject [...] This OPTIONAL property specifies a descriptor of another manifest [...]
var testManifestWithSubjectEntry = func() {
	g.Context(titleManifest, func() {
		g.Context("Setup", func() {
			g.Specify("Push file and config", func() {
				Expect(blobPut(regctlClient, imReference, "test-data/demo-file.txt")).To(BeNil())
				Expect(blobPut(regctlClient, imReference, "test-data/demo-config.txt")).To(BeNil())
			})
		})

		g.Context("Push", func() {
			g.Specify("Push manifest with invalid mediaType", func() {
				// Create first manifest
				m := getTestManifest()
				m.MediaType = mediatype.OCI1Manifest

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				first_manifest, err := manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())

				// Create second manifest
				m = getTestManifest()
				m.MediaType = mediatype.OCI1Manifest
				var subject od.Descriptor
				subject.Digest = first_manifest.GetDescriptor().Digest
				subject.Size = first_manifest.GetDescriptor().Size
				subject.MediaType = mediatype.OCI1Manifest
				m.Subject = &subject

				// Log and push to registry
				g.GinkgoWriter.Print(string(ignoreError(json.MarshalIndent(m, "", "    "))))
				_, err = manifestPutOCI(regctlClient, imReference, m)
				Expect(err).To(BeNil())
			})
		})
	})
}
