package conformance

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/bloodorangeio/reggie"
	"github.com/google/uuid"
	g "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/formatter"
	godigest "github.com/opencontainers/go-digest"
)

type (
	TagList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	TestBlob struct {
		Content       []byte
		ContentLength string
		Digest        string
	}
)

const (
	pull = 1 << iota
	push
	contentDiscovery
	contentManagement
	numWorkflows

	BLOB_UNKNOWN = iota
	BLOB_UPLOAD_INVALID
	BLOB_UPLOAD_UNKNOWN
	DIGEST_INVALID
	MANIFEST_BLOB_UNKNOWN
	MANIFEST_INVALID
	MANIFEST_UNKNOWN
	MANIFEST_UNVERIFIED
	NAME_INVALID
	NAME_UNKNOWN
	SIZE_INVALID
	TAG_INVALID
	UNAUTHORIZED
	DENIED
	UNSUPPORTED

	envVarRootURL                   = "OCI_ROOT_URL"
	envVarNamespace                 = "OCI_NAMESPACE"
	envVarUsername                  = "OCI_USERNAME"
	envVarPassword                  = "OCI_PASSWORD"
	envVarDebug                     = "OCI_DEBUG"
	envVarPull                      = "OCI_TEST_PULL"
	envVarPush                      = "OCI_TEST_PUSH"
	envVarContentDiscovery          = "OCI_TEST_CONTENT_DISCOVERY"
	envVarContentManagement         = "OCI_TEST_CONTENT_MANAGEMENT"
	envVarPushEmptyLayer            = "OCI_SKIP_EMPTY_LAYER_PUSH_TEST"
	envVarBlobDigest                = "OCI_BLOB_DIGEST"
	envVarManifestDigest            = "OCI_MANIFEST_DIGEST"
	envVarTagName                   = "OCI_TAG_NAME"
	envVarTagList                   = "OCI_TAG_LIST"
	envVarHideSkippedWorkflows      = "OCI_HIDE_SKIPPED_WORKFLOWS"
	envVarAuthScope                 = "OCI_AUTH_SCOPE"
	envVarDeleteManifestBeforeBlobs = "OCI_DELETE_MANIFEST_BEFORE_BLOBS"
	envVarCrossmountNamespace       = "OCI_CROSSMOUNT_NAMESPACE"
	envVarAutomaticCrossmount       = "OCI_AUTOMATIC_CROSSMOUNT"
	envVarReportDir                 = "OCI_REPORT_DIR"

	emptyLayerTestTag = "emptylayer"
	testTagName       = "tagtest0"

	titlePull              = "Pull"
	titlePush              = "Push"
	titleContentDiscovery  = "Content Discovery"
	titleContentManagement = "Content Management"

	//	layerBase64String is a base64 encoding of a simple tarball, obtained like this:
	//		$ echo 'you bothered to find out what was in here. Congratulations!' > test.txt
	//		$ tar czvf test.tar.gz test.txt
	//		$ cat test.tar.gz | base64
	layerBase64String = "H4sIAAAAAAAAA+3OQQrCMBCF4a49xXgBSUnaHMCTRBptQRNpp6i3t0UEV7oqIv7fYgbmzeJpHHSjVy0" +
		"WZCa1c/MufWVe94N3RWlrZ72x3k/30nhbFWKWLPU0Dhp6keJ8im//PuU/6pZH2WVtYx8b0Sz7LjWSR5VLG6YRBumSzOlGtjkd+qD" +
		"jMWiX07Befbs7AAAAAAAAAAAAAAAAAPyzO34MnqoAKAAA"

	// filter types
	artifactTypeFilter = "artifactType"
)

var (
	testMap = map[string]int{
		envVarPull:              pull,
		envVarPush:              push,
		envVarContentDiscovery:  contentDiscovery,
		envVarContentManagement: contentManagement,
	}

	testBlobA                          []byte
	testBlobALength                    string
	testBlobADigest                    string
	testRefBlobA                       []byte
	testRefBlobALength                 string
	testRefBlobADigest                 string
	testRefArtifactTypeA               string
	testRefArtifactTypeB               string
	testRefArtifactTypeIndex           string
	testRefBlobB                       []byte
	testRefBlobBLength                 string
	testRefBlobBDigest                 string
	testBlobB                          []byte
	testBlobBDigest                    string
	testBlobBChunk1                    []byte
	testBlobBChunk1Length              string
	testBlobBChunk2                    []byte
	testBlobBChunk2Length              string
	testBlobBChunk1Range               string
	testBlobBChunk2Range               string
	testBlob4MBDigest                  string
	testBlob4MBLength                  string
	testBlob4MB                        []byte
	testAnnotationKey                  string
	testAnnotationValues               map[string]string
	testMediaTypes                     []string
	testTarManifestContent             [][]byte
	testTarDigest                      []string
	testManifestAnnotationContent      []byte
	testManifestAnnotationDigest       string
	client                             *reggie.Client
	crossmountNamespace                string
	dummyDigest                        string
	errorCodes                         []string
	invalidManifestContent             []byte
	layerBlobData                      []byte
	layerBlobDigest                    string
	layerBlobContentLength             string
	emptyLayerManifestContent          []byte
	emptyLayerManifestDigest           string
	emptyConfigManifestContent         []byte
	emptyConfigManifestDigest          string
	emptyMediaTypeManifestContent      []byte
	emptyMediaTypeManifestDigest       string
	nonexistentManifest                string
	emptyJSONBlob                      []byte
	emptyJSONDescriptor                descriptor
	refsManifestAConfigArtifactContent []byte
	refsManifestAConfigArtifactDigest  string
	refsManifestALayerArtifactContent  []byte
	refsManifestALayerArtifactDigest   string
	refsManifestBConfigArtifactContent []byte
	refsManifestBConfigArtifactDigest  string
	refsManifestBLayerArtifactContent  []byte
	refsManifestBLayerArtifactDigest   string
	refsManifestCLayerArtifactContent  []byte
	refsManifestCLayerArtifactDigest   string
	refsManifestDLayerArtifactDigest   string
	refsIndexArtifactContent           []byte
	refsIndexArtifactDigest            string
	refsNestedIndexArtifactContent     []byte
	refsNestedIndexArtifactDigest      string
	reportJUnitFilename                string
	reportHTMLFilename                 string
	httpWriter                         *httpDebugWriter
	testsToRun                         int
	testManifestArtifactTypeContent    []byte
	testManifestArtifactTypeDigest     string
	testManifestConfigTypeContent      []byte
	testManifestConfigTypeDigest       string
	testManifestLayerTypeContent       []byte
	testManifestLayerTypeDigest        string
	testManifestSubjectContent         []byte
	testManifestSubjectDigest          string
	testManifest4MBContent             []byte
	testManifest4MBDigest              string
	suiteDescription                   string
	runPullSetup                       bool
	runPushSetup                       bool
	runContentDiscoverySetup           bool
	runContentManagementSetup          bool
	deleteManifestBeforeBlobs          bool
	runAutomaticCrossmountTest         bool
	automaticCrossmountEnabled         bool
	configs                            []TestBlob
	manifests                          []TestBlob
	seed                               int64
	Version                            = "unknown"
)

func init() {
	var err error

	seed = g.GinkgoRandomSeed()
	hostname := os.Getenv(envVarRootURL)
	namespace := os.Getenv(envVarNamespace)
	username := os.Getenv(envVarUsername)
	password := os.Getenv(envVarPassword)
	authScope := os.Getenv(envVarAuthScope)
	crossmountNamespace = os.Getenv(envVarCrossmountNamespace)
	if len(crossmountNamespace) == 0 {
		crossmountNamespace = fmt.Sprintf("conformance-%s", uuid.New())
	}

	debug, _ := strconv.ParseBool(os.Getenv(envVarDebug))

	for envVar, enableTest := range testMap {
		if varIsTrue, _ := strconv.ParseBool(os.Getenv(envVar)); varIsTrue {
			testsToRun |= enableTest
		}
	}

	httpWriter = newHTTPDebugWriter(debug)
	logger := newHTTPDebugLogger(httpWriter)
	client, err = reggie.NewClient(hostname,
		reggie.WithDefaultName(namespace),
		reggie.WithUsernamePassword(username, password),
		reggie.WithDebug(true),
		reggie.WithUserAgent("distribution-spec-conformance-tests"),
		reggie.WithAuthScope(authScope),
		reggie.WithInsecureSkipTLSVerify(true))
	if err != nil {
		panic(err)
	}

	client.SetLogger(logger)
	client.SetCookieJar(nil)

	// create a unique config for each workflow category
	for i := 0; i < numWorkflows; i++ {

		// in order to get a unique blob digest, we create a new author
		// field for the config on each run.
		randomAuthor := randomString(16)
		config := image{
			Architecture: "amd64",
			OS:           "linux",
			RootFS: rootFS{
				Type:    "layers",
				DiffIDs: []godigest.Digest{},
			},
			Author: randomAuthor,
		}
		configBlobContent, err := json.MarshalIndent(&config, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		configBlobContentLength := strconv.Itoa(len(configBlobContent))
		configBlobDigestRaw := godigest.FromBytes(configBlobContent)
		configBlobDigest := configBlobDigestRaw.String()
		if v := os.Getenv(envVarBlobDigest); v != "" {
			configBlobDigest = v
		}

		configs = append(configs, TestBlob{
			Content:       configBlobContent,
			ContentLength: configBlobContentLength,
			Digest:        configBlobDigest,
		})
	}

	// create layer blob
	layerBlobData, err = base64.StdEncoding.DecodeString(layerBase64String)
	if err != nil {
		log.Fatal(err)
	}

	layerBlobDigestRaw := godigest.FromBytes(layerBlobData)
	layerBlobDigest = layerBlobDigestRaw.String()
	layerBlobContentLength = fmt.Sprintf("%d", len(layerBlobData))

	layers := []descriptor{{
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Size:      int64(len(layerBlobData)),
		Digest:    layerBlobDigestRaw,
	}}

	// create a unique manifest for each workflow category
	for i := 0; i < numWorkflows; i++ {
		manifest := manifest{
			SchemaVersion: 2,
			MediaType:     "application/vnd.oci.image.manifest.v1+json",
			Config: descriptor{
				MediaType:           "application/vnd.oci.image.config.v1+json",
				Digest:              godigest.Digest(configs[i].Digest),
				Size:                int64(len(configs[i].Content)),
				Data:                configs[i].Content,    // must be the config content.
				NewUnspecifiedField: []byte("hello world"), // content doesn't matter.
			},
			Layers: layers,
		}

		manifestContent, err := json.MarshalIndent(&manifest, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		manifestContentLength := strconv.Itoa(len(manifestContent))
		manifestDigest := godigest.FromBytes(manifestContent).String()
		if v := os.Getenv(envVarManifestDigest); v != "" {
			manifestDigest = v
		}

		manifests = append(manifests, TestBlob{
			Content:       manifestContent,
			ContentLength: manifestContentLength,
			Digest:        manifestDigest,
		})
	}

	// PUSH Test
	// create manifest for push test
	emptyJSONBlob = []byte("{}")
	emptyJSONDescriptor = descriptor{
		MediaType: "application/vnd.oci.empty.v1+json",
		Size:      int64(len(emptyJSONBlob)),
		Digest:    godigest.FromBytes(emptyJSONBlob),
	}

	nonexistentManifest = ".INVALID_MANIFEST_NAME"
	invalidManifestContent = []byte("blablabla")

	dig, blob := randomBlob(42, seed+1)
	testBlobA = blob
	testBlobALength = strconv.Itoa(len(testBlobA))
	testBlobADigest = dig.String()

	setupChunkedBlob(42)

	testRefBlobA = []byte("NHL Peanut Butter on my NHL bagel")
	testRefBlobALength = strconv.Itoa(len(testRefBlobA))
	testRefBlobADigest = godigest.FromBytes(testRefBlobA).String()

	testRefArtifactTypeA = "application/vnd.nhl.peanut.butter.bagel"

	testRefBlobB = []byte("NBA Strawberry Jam on my NBA croissant")
	testRefBlobBLength = strconv.Itoa(len(testRefBlobB))
	testRefBlobBDigest = godigest.FromBytes(testRefBlobB).String()

	testRefArtifactTypeB = "application/vnd.nba.strawberry.jam.croissant"

	testAnnotationKey = "org.opencontainers.conformance.test"
	testAnnotationValues = map[string]string{}

	// manifest with empty mediatype
	emptyMediaTypeManifest := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.empty.v1+json",
		ArtifactType:  "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    godigest.Digest(configs[0].Digest),
			Size:      int64(len(configs[0].Content)),
		},
		Layers: layers,
	}
	emptyMediaTypeManifestContent, err = json.MarshalIndent(&emptyMediaTypeManifest, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	emptyMediaTypeManifestDigest = string(godigest.FromBytes(emptyMediaTypeManifestContent))

	// manifest with empty layer
	emptyLayerManifest := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType:           "application/vnd.oci.image.config.v1+json",
			Digest:              godigest.Digest(configs[1].Digest),
			Size:                int64(len(configs[1].Content)),
			Data:                configs[1].Content,    // must be the config content.
			NewUnspecifiedField: []byte("hello world"), // content doesn't matter.
		},
		Layers: []descriptor{
			emptyJSONDescriptor,
		},
	}

	emptyLayerManifestContent, err = json.MarshalIndent(&emptyLayerManifest, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	emptyLayerManifestDigest = string(godigest.FromBytes(emptyLayerManifestContent))

	// manifest with empty config
	emptyConfigManifest := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config:        emptyJSONDescriptor,
		Layers:        layers,
	}

	emptyConfigManifestContent, err = json.MarshalIndent(&emptyConfigManifest, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	emptyConfigManifestDigest = string(godigest.FromBytes(emptyConfigManifestContent))

	// Manifest with artifact using custom config.MediaType = artifactType
	testManifestConfigType := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: testRefArtifactTypeA,
			Size:      int64(len(testRefBlobA)),
			Digest:    godigest.FromBytes(testRefBlobA),
		},
		Layers: layers,
	}

	testManifestConfigTypeContent, err = json.MarshalIndent(&testManifestConfigType, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	testManifestConfigTypeDigest = godigest.FromBytes(testManifestConfigTypeContent).String()

	// Manifest with custom layer media type
	testManifestLayerType := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    godigest.Digest(configs[0].Digest),
			Size:      int64(len(configs[0].Content)),
			Data:      configs[0].Content,
		},
		Layers: []descriptor{{
			MediaType: testRefArtifactTypeA,
			Size:      int64(len(testRefBlobA)),
			Digest:    godigest.FromBytes(testRefBlobA),
		}},
	}

	testManifestLayerTypeContent, err = json.MarshalIndent(&testManifestLayerType, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	testManifestLayerTypeDigest = godigest.FromBytes(testManifestLayerTypeContent).String()

	// Manifests for different tar MediaType
	testMediaTypes = []string{
		"application/vnd.oci.image.layer.v1.tar",
		"application/vnd.oci.image.layer.v1.tar+gzip",
		"application/vnd.oci.image.layer.nondistributable.v1.tar",
		"application/vnd.oci.image.layer.nondistributable.v1.tar+gzip",
		"application/vnd.oci.image.layer.v1.tar+zstd",
	}
	for _, t := range testMediaTypes {
		newManifest := manifest{
			SchemaVersion: 2,
			MediaType:     "application/vnd.oci.image.manifest.v1+json",
			Config: descriptor{
				MediaType: "application/vnd.oci.image.config.v1+json",
				Digest:    godigest.Digest(configs[0].Digest),
				Size:      int64(len(configs[0].Content)),
				Data:      configs[0].Content,
			},
			Layers: []descriptor{{
				MediaType: t,
				Size:      int64(len(layerBlobData)),
				Digest:    layerBlobDigestRaw,
			}},
		}
		testManifestArtifactTypeContent, err = json.MarshalIndent(&newManifest, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
		testTarManifestContent = append(testTarManifestContent, testManifestArtifactTypeContent)
		testManifestArtifactTypeDigest = godigest.FromBytes(testManifestArtifactTypeContent).String()
		testTarDigest = append(testTarDigest, testManifestArtifactTypeDigest)
	}

	// artifact with Subject ref using ArtifactType, config.MediaType = emptyJSON
	testManifestArtifactType := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  testRefArtifactTypeA,
		Config: descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(testRefBlobA)),
			Digest:    godigest.FromBytes(testRefBlobA),
		},
		Layers: layers,
	}

	testManifestArtifactTypeContent, err = json.MarshalIndent(&testManifestArtifactType, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	testManifestArtifactTypeDigest = godigest.FromBytes(testManifestArtifactTypeContent).String()

	//Manifest with Subject
	testManifestSubject := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    godigest.FromBytes(configs[0].Content),
			Size:      int64(len(configs[0].Content)),
			Data:      configs[0].Content, // must be the config content.
		},
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[0].Content)),
			Digest:    godigest.FromBytes(manifests[0].Content),
		},
		Layers: layers,
	}

	testManifestSubjectContent, err = json.MarshalIndent(&testManifestSubject, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	testManifestSubjectDigest = godigest.FromBytes(testManifestSubjectContent).String()

	// Manifest with Annotations
	testManifestAnnotation := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    godigest.Digest(configs[0].Digest),
			Size:      int64(len(configs[0].Content)),
		},
		Layers: layers,
		Annotations: map[string]string{
			testAnnotationKey: "test annotation",
		},
	}
	testManifestAnnotationContent, err = json.MarshalIndent(&testManifestAnnotation, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	testManifestAnnotationDigest = godigest.FromBytes(testManifestAnnotationContent).String()
	testAnnotationValues[testManifestAnnotationDigest] = testManifestAnnotation.Annotations[testAnnotationKey]

	// Manifest with 4 Megabytes
	digest4MB, blob4MB := randomBlob(4*1024*1024, seed+1)
	testBlob4MB = blob4MB
	testBlob4MBDigest = digest4MB.String()
	testBlob4MBLength = strconv.Itoa(len(testBlob4MB))
	testManifest4MB := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    godigest.Digest(configs[0].Digest),
			Size:      int64(len(configs[0].Content)),
			Data:      configs[0].Content,
		},
		Layers: []descriptor{{
			MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			Size:      int64(len(blob4MB)), //4MB
			Digest:    digest4MB,
		}},
	}
	testManifest4MBContent, err = json.MarshalIndent(&testManifest4MB, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	testManifest4MBDigest = godigest.FromBytes(testManifest4MBContent).String()

	// Index manifest
	manifestContentLength, err := strconv.Atoi(manifests[0].ContentLength)
	if err != nil {
		log.Fatal(err)
	}
	refsIndexArtifact := index{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		ArtifactType:  testRefArtifactTypeIndex,
		Manifests: []descriptor{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Size:      int64(manifestContentLength),
				Digest:    godigest.Digest(manifests[0].Digest),
			},
		},
		Annotations: map[string]string{
			testAnnotationKey: "test index",
		},
	}
	refsIndexArtifactContent, err = json.MarshalIndent(&refsIndexArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	refsIndexArtifactDigest = godigest.FromBytes(refsIndexArtifactContent).String()

	// refsIndexArtifactContentLength := strconv.Itoa(len(refsIndexArtifactContent))
	refsIndexArtifactContentLength := len(refsIndexArtifactContent)

	// Nested index manifest
	refsNestedIndexArtifact := index{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.index.v1+json",
		Manifests: []descriptor{
			{
				MediaType: "application/vnd.oci.image.index.v1+json",
				Size:      int64(refsIndexArtifactContentLength),
				Digest:    godigest.Digest(refsIndexArtifactDigest),
			},
		},
		Annotations: map[string]string{
			testAnnotationKey: "test index",
		},
	}
	refsNestedIndexArtifactContent, err = json.MarshalIndent(&refsNestedIndexArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	refsNestedIndexArtifactDigest = godigest.FromBytes(refsNestedIndexArtifactContent).String()

	// CONTENT DISCOVERY Test
	// create manifests for referrers test (artifacts with Subject field set)
	// artifact with Subject ref using config.MediaType = artifactType
	refsManifestAConfigArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: testRefArtifactTypeA,
			Size:      int64(len(testRefBlobA)),
			Digest:    godigest.FromBytes(testRefBlobA),
		},
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			emptyJSONDescriptor,
		},
		Annotations: map[string]string{
			testAnnotationKey: "test config a",
		},
	}

	refsManifestAConfigArtifactContent, err = json.MarshalIndent(&refsManifestAConfigArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestAConfigArtifactDigest = godigest.FromBytes(refsManifestAConfigArtifactContent).String()
	testAnnotationValues[refsManifestAConfigArtifactDigest] = refsManifestAConfigArtifact.Annotations[testAnnotationKey]

	refsManifestBConfigArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		Config: descriptor{
			MediaType: testRefArtifactTypeB,
			Size:      int64(len(testRefBlobB)),
			Digest:    godigest.FromBytes(testRefBlobB),
		},
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			emptyJSONDescriptor,
		},
		Annotations: map[string]string{
			testAnnotationKey: "test config b",
		},
	}

	refsManifestBConfigArtifactContent, err = json.MarshalIndent(&refsManifestBConfigArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestBConfigArtifactDigest = godigest.FromBytes(refsManifestBConfigArtifactContent).String()
	testAnnotationValues[refsManifestBConfigArtifactDigest] = refsManifestBConfigArtifact.Annotations[testAnnotationKey]

	// artifact with Subject ref using ArtifactType, config.MediaType = emptyJSON
	refsManifestALayerArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  testRefArtifactTypeA,
		Config:        emptyJSONDescriptor,
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			{
				MediaType: testRefArtifactTypeA,
				Size:      int64(len(testRefBlobA)),
				Digest:    godigest.FromBytes(testRefBlobA),
			},
		},
		Annotations: map[string]string{
			testAnnotationKey: "test layer a",
		},
	}

	refsManifestALayerArtifactContent, err = json.MarshalIndent(&refsManifestALayerArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestALayerArtifactDigest = godigest.FromBytes(refsManifestALayerArtifactContent).String()
	testAnnotationValues[refsManifestALayerArtifactDigest] = refsManifestALayerArtifact.Annotations[testAnnotationKey]

	refsManifestBLayerArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  testRefArtifactTypeB,
		Config:        emptyJSONDescriptor,
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[4].Content)),
			Digest:    godigest.FromBytes(manifests[4].Content),
		},
		Layers: []descriptor{
			{
				MediaType: testRefArtifactTypeB,
				Size:      int64(len(testRefBlobB)),
				Digest:    godigest.FromBytes(testRefBlobB),
			},
		},
		Annotations: map[string]string{
			testAnnotationKey: "test layer b",
		},
	}

	refsManifestBLayerArtifactContent, err = json.MarshalIndent(&refsManifestBLayerArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	refsManifestBLayerArtifactDigest = godigest.FromBytes(refsManifestBLayerArtifactContent).String()
	testAnnotationValues[refsManifestBLayerArtifactDigest] = refsManifestBLayerArtifact.Annotations[testAnnotationKey]

	// ManifestCLayerArtifact is the same as B but based on a subject that has not been pushed
	refsManifestCLayerArtifact := manifest{
		SchemaVersion: 2,
		MediaType:     "application/vnd.oci.image.manifest.v1+json",
		ArtifactType:  testRefArtifactTypeB,
		Config:        emptyJSONDescriptor,
		Subject: &descriptor{
			MediaType: "application/vnd.oci.image.manifest.v1+json",
			Size:      int64(len(manifests[3].Content)),
			Digest:    godigest.FromBytes(manifests[3].Content),
		},
		Layers: []descriptor{
			{
				MediaType: testRefArtifactTypeB,
				Size:      int64(len(testRefBlobB)),
				Digest:    godigest.FromBytes(testRefBlobB),
			},
		},
	}

	refsManifestCLayerArtifactContent, err = json.MarshalIndent(&refsManifestCLayerArtifact, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	dummyDigest = godigest.FromString("hello world").String()

	errorCodes = []string{
		BLOB_UNKNOWN:          "BLOB_UNKNOWN",
		BLOB_UPLOAD_INVALID:   "BLOB_UPLOAD_INVALID",
		BLOB_UPLOAD_UNKNOWN:   "BLOB_UPLOAD_UNKNOWN",
		DIGEST_INVALID:        "DIGEST_INVALID",
		MANIFEST_BLOB_UNKNOWN: "MANIFEST_BLOB_UNKNOWN",
		MANIFEST_INVALID:      "MANIFEST_INVALID",
		MANIFEST_UNKNOWN:      "MANIFEST_UNKNOWN",
		MANIFEST_UNVERIFIED:   "MANIFEST_UNVERIFIED",
		NAME_INVALID:          "NAME_INVALID",
		NAME_UNKNOWN:          "NAME_UNKNOWN",
		SIZE_INVALID:          "SIZE_INVALID",
		TAG_INVALID:           "TAG_INVALID",
		UNAUTHORIZED:          "UNAUTHORIZED",
		DENIED:                "DENIED",
		UNSUPPORTED:           "UNSUPPORTED",
	}

	runPullSetup = true
	runPushSetup = true
	runContentDiscoverySetup = true
	runContentManagementSetup = true
	deleteManifestBeforeBlobs = true

	if os.Getenv(envVarTagName) != "" &&
		os.Getenv(envVarManifestDigest) != "" &&
		os.Getenv(envVarBlobDigest) != "" {
		runPullSetup = false
	}

	if os.Getenv(envVarTagList) != "" {
		runContentDiscoverySetup = false
	}

	if v, ok := os.LookupEnv(envVarDeleteManifestBeforeBlobs); ok {
		deleteManifestBeforeBlobs, _ = strconv.ParseBool(v)
	}
	automaticCrossmountVal := ""
	automaticCrossmountVal, runAutomaticCrossmountTest = os.LookupEnv(envVarAutomaticCrossmount)
	automaticCrossmountEnabled, _ = strconv.ParseBool(automaticCrossmountVal)

	if dir := os.Getenv(envVarReportDir); dir != "none" {
		reportJUnitFilename = filepath.Join(dir, "junit.xml")
		reportHTMLFilename = filepath.Join(dir, "report.html")
	}
	suiteDescription = "OCI Distribution Conformance Tests"
}

func SkipIfDisabled(test int) {
	if userDisabled(test) {
		report := generateSkipReport()
		g.Skip(report)
	}
}

func RunOnlyIf(v bool) {
	if !v {
		g.Skip("you have skipped this test.")
	}
}

func RunOnlyIfNot(v bool) {
	if v {
		g.Skip("you have skipped this test.")
	}
}

func Warn(message string) {
	// print message
	fmt.Fprint(os.Stderr, formatter.Fi(2, "\n{{magenta}}WARNING: %s\n{{/}}", message))
	// print file:line
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(os.Stderr, formatter.Fi(2, "\n%s:%d\n", file, line))
}

func generateSkipReport() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "you have skipped this test; if this is an error, check your environment variable settings:\n")
	for k := range testMap {
		fmt.Fprintf(buf, "\t%s=%s\n", k, os.Getenv(k))
	}
	return buf.String()
}

func userDisabled(test int) bool {
	return !(test&testsToRun > 0)
}

func getTagList(resp *reggie.Response) []string {
	jsonData := resp.Body()
	tagList := &TagList{}
	err := json.Unmarshal(jsonData, tagList)
	if err != nil {
		return []string{}
	}

	return tagList.Tags
}

// Adapted from https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func randomString(n int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}

// randomBlob outputs a reproducible random blob (based on the seed) for testing
func randomBlob(size int, seed int64) (godigest.Digest, []byte) {
	r := mathrand.New(mathrand.NewSource(seed))
	b := make([]byte, size)
	if n, err := r.Read(b); err != nil {
		panic(err)
	} else if n != size {
		panic("unable to read enough bytes")
	}
	return godigest.FromBytes(b), b
}

func setupChunkedBlob(size int) {
	dig, blob := randomBlob(size, seed+2)
	testBlobB = blob
	testBlobBDigest = dig.String()
	testBlobBChunk1 = testBlobB[:size/2+1]
	testBlobBChunk1Length = strconv.Itoa(len(testBlobBChunk1))
	testBlobBChunk1Range = fmt.Sprintf("0-%d", len(testBlobBChunk1)-1)
	testBlobBChunk2 = testBlobB[size/2+1:]
	testBlobBChunk2Length = strconv.Itoa(len(testBlobBChunk2))
	testBlobBChunk2Range = fmt.Sprintf("%d-%d", len(testBlobBChunk1), len(testBlobB)-1)
}
