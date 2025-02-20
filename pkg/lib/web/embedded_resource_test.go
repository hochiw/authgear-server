package web_test

import (
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/web"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGlobalEmbeddedResourceManager(t *testing.T) {
	Convey("GlobalEmbeddedResourceManager", t, func() {
		Convey("should create GlobalEmbeddedResourceManager", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/" + web.DefaultResourceDir,
				ResourcePrefix: web.DefaultResourcePrefix,
				Name:           web.DefaultManifestName,
			})
			So(err, ShouldBeNil)
			defer m.Close()

			So(m.Manifest.ResourceDir, ShouldEqual, "testdata/"+web.DefaultResourceDir)
			So(m.Manifest.ResourcePrefix, ShouldEqual, web.DefaultResourcePrefix)
			So(m.Manifest.Name, ShouldEqual, web.DefaultManifestName)
		})

		Convey("should throw error if manifest file does not exist", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/" + web.DefaultResourceDir,
				ResourcePrefix: web.DefaultResourcePrefix,
				Name:           "test.json",
			})
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
			So(m, ShouldBeNil)
		})

		Convey("should throw error if resource directory does not exist", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/123/",
				ResourcePrefix: web.DefaultResourcePrefix,
				Name:           "test.json",
			})
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
			So(m, ShouldBeNil)
		})

		Convey("should throw error if resource prefix does not exist", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/" + web.DefaultResourceDir,
				ResourcePrefix: "123",
				Name:           "test.json",
			})
			So(err.Error(), ShouldContainSubstring, "no such file or directory")
			So(m, ShouldBeNil)
		})

		Convey("should load manifest content after manager created", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/" + web.DefaultResourceDir,
				ResourcePrefix: web.DefaultResourcePrefix,
				Name:           web.DefaultManifestName,
			})
			So(err, ShouldBeNil)
			defer m.Close()

			manifestContext := m.GetManifestContext()

			So(manifestContext, ShouldResemble, &web.ManifestContext{
				Content: map[string]string{"test.js": "test.12345678.js"},
			})
			So(err, ShouldBeNil)
		})

		Convey("should reload manifest with any changes", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/" + web.DefaultResourceDir,
				ResourcePrefix: web.DefaultResourcePrefix,
				Name:           web.DefaultManifestName,
			})
			So(err, ShouldBeNil)
			defer m.Close()

			filePath := "testdata/" + web.DefaultResourceDir + web.DefaultResourcePrefix + web.DefaultManifestName

			f, err := os.Open(filePath)
			So(err, ShouldBeNil)
			originalContent, err := io.ReadAll(f)
			So(err, ShouldBeNil)
			defer f.Close()

			recoverFileContent := func() {
				// nolint:gosec
				_ = os.WriteFile(
					filePath,
					originalContent,
					0666,
				)
			}

			defer recoverFileContent()

			// nolint:gosec
			err = os.WriteFile(
				filePath,
				[]byte(`{"anotherTest.js": "anotherTest.12345678.js"}`),
				0666,
			)
			So(err, ShouldBeNil)
			runtime.Gosched()
			time.Sleep(500 * time.Millisecond)

			manifestContext := m.GetManifestContext()
			So(manifestContext, ShouldResemble, &web.ManifestContext{
				Content: map[string]string{"anotherTest.js": "anotherTest.12345678.js"},
			})
		})

		Convey("should return asset file name and prefix by key", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/" + web.DefaultResourceDir,
				ResourcePrefix: web.DefaultResourcePrefix,
				Name:           web.DefaultManifestName,
			})
			So(err, ShouldBeNil)
			defer m.Close()

			// if key exists
			prefix, assetFileName, err := m.AssetPath("test.js")
			So(err, ShouldBeNil)
			So(prefix, ShouldEqual, web.DefaultResourcePrefix)
			So(assetFileName, ShouldEqual, "test.12345678.js")

			// if key does not exist
			prefix, assetFileName, err = m.AssetPath("test123.js")
			So(err, ShouldBeError, "specified resource is not configured")
			So(prefix, ShouldBeEmpty)
			So(assetFileName, ShouldBeEmpty)
		})

		Convey("should return asset file name by resource path", func() {
			m, err := web.NewGlobalEmbeddedResourceManager(&web.Manifest{
				ResourceDir:    "testdata/" + web.DefaultResourceDir,
				ResourcePrefix: web.DefaultResourcePrefix,
				Name:           web.DefaultManifestName,
			})
			So(err, ShouldBeNil)
			defer m.Close()

			// if key exists with correct hash
			assetFileName, ok := m.Resolve(web.DefaultResourcePrefix + "test.12345678.js")
			So(ok, ShouldBeTrue)
			So(assetFileName, ShouldEqual, "test.12345678.js")

			// if key exists with incorrect hash
			assetFileName, ok = m.Resolve(web.DefaultResourcePrefix + "test.abcdefgh.js")
			So(ok, ShouldBeFalse)
			So(assetFileName, ShouldBeEmpty)

			// if key does not exist
			assetFileName, ok = m.Resolve(web.DefaultResourcePrefix + "test123.12345678.js")
			So(ok, ShouldBeFalse)
			So(assetFileName, ShouldBeEmpty)
		})
	})
}
