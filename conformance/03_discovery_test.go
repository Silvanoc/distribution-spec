package conformance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/bloodorangeio/reggie"
	g "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	godigest "github.com/opencontainers/go-digest"
)

var test03ContentDiscovery = func() {
	g.Context(titleContentDiscovery, func() {

		var numTags = 4
		var tagList []string
		var supportPutNonExsistSubject = 1

		g.Context("Setup", func() {
			g.Specify("Populate registry with test blob", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", configs[2].Digest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", configs[2].ContentLength).
					SetBody(configs[2].Content)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test layer", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", layerBlobDigest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", layerBlobContentLength).
					SetBody(layerBlobData)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
			})

			g.Specify("Populate registry with test tags", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				for i := 0; i < numTags; i++ {
					for _, tag := range []string{"test" + strconv.Itoa(i), "TEST" + strconv.Itoa(i)} {
						tagList = append(tagList, tag)
						req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
							reggie.WithReference(tag)).
							SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
							SetBody(manifests[2].Content)
						resp, err := client.Do(req)
						Expect(err).To(BeNil())
						Expect(resp.StatusCode()).To(SatisfyAll(
							BeNumerically(">=", 200),
							BeNumerically("<", 300)))
					}
				}
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				tagList = getTagList(resp)
				_ = err
			})

			g.Specify("Populate registry with test tags (no push)", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIfNot(runContentDiscoverySetup)
				tagList = strings.Split(os.Getenv(envVarTagList), ",")
			})

			g.Specify("References setup", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)

				// Populate registry with empty JSON blob
				// validate expected empty JSON blob digest
				Expect(emptyJSONDescriptor.Digest).To(Equal(godigest.Digest("sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a")))
				req := client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", emptyJSONDescriptor.Digest.String()).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", fmt.Sprintf("%d", emptyJSONDescriptor.Size)).
					SetBody(emptyJSONBlob)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))

				// Populate registry with test blob 0
				req = client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", configs[0].Digest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", configs[0].ContentLength).
					SetBody(configs[0].Content)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))

				// Populate registry with test blob 4
				req = client.NewRequest(reggie.POST, "/v2/<name>/blobs/uploads/")
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				req = client.NewRequest(reggie.PUT, resp.GetRelativeLocation()).
					SetQueryParam("digest", configs[4].Digest).
					SetHeader("Content-Type", "application/octet-stream").
					SetHeader("Content-Length", configs[4].ContentLength).
					SetBody(configs[4].Content)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))

				// Populate registry with test manifest 4
				tag := testTagName
				req = client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(tag)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(manifests[4].Content)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))

				// Populate registry with test manifests with annotations
				req = client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(testManifestAnnotationDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(testManifestAnnotationContent)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)), getErrorsInfo(resp))
			})

			g.Specify("Populate registry with manifests containing Subject should return Header with OCI-Subject", func() {
				if !supportSubject {
					SkipIfDisabled(0)
				}
				// Populate registry with test references manifest
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestArtifactADigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestArtifactAContent)
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(manifests[4].Digest))

				// Populate registry with test references manifest with artifact type
				req = client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestArtifactBDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestArtifactBContent)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(manifests[4].Digest))

				// Populate registry with test references manifest refers to manifest with annotations
				req = client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestCopyAnnotationDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestCopyAnnotationContent)
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)), getErrorsInfo(resp))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(testManifestAnnotationDigest))

			})

			g.Specify("Populate registry with test references manifest to a non-existent subject", func() {
				req := client.NewRequest(reggie.PUT, "/v2/<name>/manifests/<reference>",
					reggie.WithReference(refsManifestCLayerArtifactDigest)).
					SetHeader("Content-Type", "application/vnd.oci.image.manifest.v1+json").
					SetBody(refsManifestCLayerArtifactContent)
				resp, err := client.Do(req)
				if err != nil || resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
					supportPutNonExsistSubject = 0
				}
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAll(
					BeNumerically(">=", 200),
					BeNumerically("<", 300)))
				Expect(resp.Header().Get("OCI-Subject")).To(Equal(manifests[3].Digest))
			})
		})

		g.Context("Test content discovery endpoints (listing tags)", func() {
			g.Specify("GET request to list tags should yield 200 response and be in sorted order", func() {
				SkipIfDisabled(contentDiscovery)
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK), getErrorsInfo(resp))
				tagList = getTagList(resp)
				numTags = len(tagList)
				// If the list is not empty, the tags MUST be in lexical order (i.e. case-insensitive alphanumeric order).
				sortedTagListLexical := append([]string{}, tagList...)
				sort.SliceStable(sortedTagListLexical, func(i, j int) bool {
					return strings.ToLower(sortedTagListLexical[i]) < strings.ToLower(sortedTagListLexical[j])
				})
				// Historically, registries have not been lexical, so allow `sort.Strings` to be valid too.
				sortedTagListAsciibetical := append([]string{}, tagList...)
				sort.Strings(sortedTagListAsciibetical)
				Expect(tagList).To(Or(Equal(sortedTagListLexical), Equal(sortedTagListAsciibetical)))
			})

			g.Specify("GET number of tags should be limitable by `n` query parameter", func() {
				SkipIfDisabled(contentDiscovery)
				numResults := numTags / 2
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK), getErrorsInfo(resp))
				tagList = getTagList(resp)
				Expect(len(tagList)).To(Equal(numResults))
			})

			g.Specify("GET start of tag is set by `last` query parameter", func() {
				SkipIfDisabled(contentDiscovery)
				numResults := numTags / 2
				req := client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK))
				tagList = getTagList(resp)
				last := tagList[numResults-1]
				req = client.NewRequest(reggie.GET, "/v2/<name>/tags/list").
					SetQueryParam("n", strconv.Itoa(numResults)).
					SetQueryParam("last", tagList[numResults-1])
				resp, err = client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK), getErrorsInfo(resp))
				tagList = getTagList(resp)
				Expect(len(tagList)).To(BeNumerically("<=", numResults))
				Expect(tagList).ToNot(ContainElement(last))
			})
		})

		g.Context("Test content discovery endpoints (listing references)", func() {

			g.Specify("GET request to existing blob should yield 200", func() {
				SkipIfDisabled(contentDiscovery)
				if !supportSubject {
					SkipIfDisabled(0)
				}
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[4].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK), getErrorsInfo(resp))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(Equal(2))
				Expect(index.Manifests[0].Digest).ToNot(Equal(index.Manifests[1].Digest))
			})

			g.Specify("GET request to existing blob with filter should yield 200", func() {
				SkipIfDisabled(contentDiscovery)
				if !supportSubject {
					SkipIfDisabled(0)
				}
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[4].Digest)).
					SetQueryParam("artifactType", "application/vnd.oci.descriptor.v1+json")
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK), getErrorsInfo(resp))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())

				// also check resp header "OCI-Filters-Applied: artifactType" denoting that an artifactType filter was applied
				if resp.Header().Get("OCI-Filters-Applied") != "" {
					Expect(len(index.Manifests)).To(Equal(1))
					Expect(resp.Header().Get("OCI-Filters-Applied")).To(Equal(artifactTypeFilter))
				} else {
					Expect(len(index.Manifests)).To(Equal(2))
					Warn("filtering by artifact-type is not implemented")
				}
			})

			g.Specify("GET request to missing manifest in subject should yield 200", func() {
				SkipIfDisabled(contentDiscovery)
				SkipIfDisabled(supportPutNonExsistSubject)
				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(manifests[3].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK), getErrorsInfo(resp))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(Equal(1))
				Expect(index.Manifests[0].Digest.String()).To(Equal(refsManifestCLayerArtifactDigest))
			})
		})

		g.Context("Manifest copy the annotations of the subject", func() {
			g.Specify("Manifests in the referrers list should contain the annotations of the subject", func() {
				SkipIfDisabled(contentDiscovery)
				if !supportAnnotation || !supportSubject {
					SkipIfDisabled(0)
				}

				req := client.NewRequest(reggie.GET, "/v2/<name>/referrers/<digest>",
					reggie.WithDigest(testManifestAnnotationDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(Equal(http.StatusOK), getErrorsInfo(resp))
				Expect(resp.Header().Get("Content-Type")).To(Equal("application/vnd.oci.image.index.v1+json"))

				var index index
				err = json.Unmarshal(resp.Body(), &index)
				Expect(err).To(BeNil())
				Expect(len(index.Manifests)).To(Equal(1))
				Expect(index.Manifests[0].Annotations[testAnnotationKey]).To(Equal(testAnnotationValues[testManifestAnnotationDigest]))
			})
		})

		g.Context("Teardown", func() {
			if deleteManifestBeforeBlobs {
				g.Specify("Delete created manifest & associated tags", func() {
					SkipIfDisabled(contentDiscovery)
					RunOnlyIf(runContentDiscoverySetup)
					references := []string{
						manifests[2].Digest,
						manifests[4].Digest,
						refsManifestArtifactADigest,
						refsManifestArtifactBDigest,
					}
					if supportPutNonExsistSubject == 1 {
						references = append(references, refsManifestCLayerArtifactDigest)
					}
					for _, ref := range references {
						req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(ref))
						resp, err := client.Do(req)
						Expect(err).To(BeNil())
						Expect(resp.StatusCode()).To(SatisfyAny(
							SatisfyAll(
								BeNumerically(">=", 200),
								BeNumerically("<", 300),
							),
							Equal(http.StatusNotFound),
							Equal(http.StatusMethodNotAllowed),
						))
					}
				})
			}

			g.Specify("Delete config blob created in tests", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)

				deleteReq := func(req *reggie.Request) {
					resp, err := client.Do(req)
					Expect(err).To(BeNil())
					Expect(resp.StatusCode()).To(SatisfyAny(
						SatisfyAll(
							BeNumerically(">=", 200),
							BeNumerically("<", 300),
						),
						Equal(http.StatusNotFound),
						Equal(http.StatusMethodNotAllowed),
					))
				}

				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configs[2].Digest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300),
					),
					Equal(http.StatusNotFound),
					Equal(http.StatusMethodNotAllowed),
				))

				// Delete config blob created in setup
				req = client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(configs[4].Digest))
				deleteReq(req)

				// Delete empty JSON blob created in setup
				req = client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(emptyJSONDescriptor.Digest.String()))
				deleteReq(req)
			})

			g.Specify("Delete layer blob created in setup", func() {
				SkipIfDisabled(contentDiscovery)
				RunOnlyIf(runContentDiscoverySetup)
				req := client.NewRequest(reggie.DELETE, "/v2/<name>/blobs/<digest>", reggie.WithDigest(layerBlobDigest))
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode()).To(SatisfyAny(
					SatisfyAll(
						BeNumerically(">=", 200),
						BeNumerically("<", 300),
					),
					Equal(http.StatusNotFound),
					Equal(http.StatusMethodNotAllowed),
				))
			})

			if !deleteManifestBeforeBlobs {
				g.Specify("Delete created manifest & associated tags", func() {
					SkipIfDisabled(contentDiscovery)
					RunOnlyIf(runContentDiscoverySetup)
					references := []string{
						manifests[2].Digest,
						manifests[4].Digest,
						refsManifestArtifactADigest,
						refsManifestArtifactBDigest,
						refsManifestCLayerArtifactDigest,
					}
					for _, ref := range references {
						req := client.NewRequest(reggie.DELETE, "/v2/<name>/manifests/<digest>", reggie.WithDigest(ref))
						resp, err := client.Do(req)
						Expect(err).To(BeNil())
						Expect(resp.StatusCode()).To(SatisfyAny(
							SatisfyAll(
								BeNumerically(">=", 200),
								BeNumerically("<", 300),
							),
							Equal(http.StatusMethodNotAllowed),
							Equal(http.StatusNotFound),
						))
					}
				})
			}
		})
	})
}
