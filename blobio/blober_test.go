package blobio

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	_ "gocloud.dev/blob/fileblob"
)

var _ = Describe("Blober", func() {
	var (
		blobstoreDir string

		err          error
		bucketURL    string

		logger lager.Logger
		bucketConfig BucketConfig

	)

	Context("Pull blobs using local file system backed bucket", func() {
		var (
			fileToPull *os.File
			tmpInputDir string
			sourcePath string
		)
		BeforeEach(func() {
			logger = lager.NewLogger("porter-pull-test")
			sink := lager.NewWriterSink(os.Stderr, lager.DEBUG)
			logger.RegisterSink(sink)

			blobstoreDir, err = ioutil.TempDir("/tmp", "porter-test-pull-blobstore")
			if err != nil {
				log.Fatal(err)
			}

			sourcePath = "my-key"

			fileToPull, err = ioutil.TempFile("/tmp", "some-pull-file")
			if err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile(fileToPull.Name(), []byte("hello world"), 0777)
			if err != nil {
				log.Fatal(err)
			}

			CreateTarball(blobstoreDir+"/" + sourcePath, []string{fileToPull.Name()})

			bucketURL = "file://" + filepath.ToSlash(blobstoreDir)

			bucketConfig = BucketConfig{
				URL: bucketURL,
			}

			tmpInputDir, err = ioutil.TempDir("/tmp", "porter-test-destination")
			if err != nil {
				log.Fatal(err)
			}

		})
		AfterEach(func() {
			os.Remove(fileToPull.Name())
			os.RemoveAll(blobstoreDir)
			os.RemoveAll(tmpInputDir)

		})
		It("pull an archive with the correct extracted content", func() {

			err := Pull(logger, context.Background(), bucketConfig, sourcePath, tmpInputDir)
			Expect(err).ToNot(HaveOccurred())

			fileContents, err := ioutil.ReadFile(tmpInputDir + fileToPull.Name())
			Expect(string(fileContents)).To(Equal("hello world"))
			// invoke in
			// ensure the blobs match what we expect
		})

	})


	Context("Push blobs using local file system backed bucket", func() {
		var (
			tmpOutputDir string
			tmpInnerDir string
			destinationKey string
		)
		BeforeEach(func() {
			logger = lager.NewLogger("porter-push-test")
			sink := lager.NewWriterSink(os.Stderr, lager.DEBUG)
			logger.RegisterSink(sink)


			tmpOutputDir, err = ioutil.TempDir("/tmp", "porter-test-push-dir")
			if err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile( tmpOutputDir + "/foo-file", []byte("hello world"), 0644)
			if err != nil {
				log.Fatal(err)
			}

			tmpInnerDir, err = ioutil.TempDir(tmpOutputDir, "bar-dir")
			if err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile( tmpInnerDir + "/baz-file", []byte("hello world 2"), 0644)
			if err != nil {
				log.Fatal(err)
			}

			blobstoreDir, err = ioutil.TempDir("/tmp", "porter-test-push-blobstore")
			if err != nil {
				log.Fatal(err)
			}

			// when using the file blobstore, the destination is called fileblob
			// https://github.com/google/go-cloud/blob/a88a3e5785d930df88d97df4b4b2965ca99f05b9/blob/fileblob/fileblob.go#L487
			destinationKey = "fileblob"

			bucketURL = "file://" + filepath.ToSlash(blobstoreDir)

			bucketConfig = BucketConfig{
				URL: bucketURL,
			}


		})
		AfterEach(func() {
			os.RemoveAll(blobstoreDir)
			os.RemoveAll(tmpOutputDir)

		})
		It("push outputs dir to a compressed archive", func() {
			err := Push(logger, context.Background(), bucketConfig, tmpOutputDir, destinationKey)
			Expect(err).ToNot(HaveOccurred())
			// Test compressed file is present in blobstore and its contents are hello world

			var files []string
			// get filename in blobstore
			err = filepath.Walk(blobstoreDir, func(path string, info os.FileInfo, err error) error {
				files = append(files, path)
				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			contents, err := readArchive(files[1])
			Expect(err).ToNot(HaveOccurred())
			Expect(contents[tmpOutputDir + "/foo-file"]).To(Equal("hello world"))
			Expect(contents[tmpInnerDir + "/baz-file"]).To(Equal("hello world 2"))
		})

	})
})
