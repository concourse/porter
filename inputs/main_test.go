package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	_ "gocloud.dev/blob/fileblob"
)

var _ = Describe("Inputs", func() {
	var (
		blobstoreDir string
		fileToPull *os.File
		err          error
		bucketURL    string
		sourcePath string

	)

	Context("using local file system backed bucket", func() {
		BeforeEach(func() {
			blobstoreDir, err = ioutil.TempDir("/tmp", "porter-test-blobstore")
			if err != nil {
				log.Fatal(err)
			}

			sourcePath = "my-key"

			fileToPull, err = ioutil.TempFile("/tmp", "some-file")
			if err != nil {
				log.Fatal(err)
			}

			err = ioutil.WriteFile(fileToPull.Name(), []byte("hello world"), 0777)
			if err != nil {
				log.Fatal(err)
			}

			CreateTarball(blobstoreDir+"/" + sourcePath, []string{fileToPull.Name()})

			bucketURL = "file://" + filepath.ToSlash(blobstoreDir)


		})
		AfterEach(func() {
			os.RemoveAll(blobstoreDir)
			os.Remove(fileToPull.Name())

		})
		It("pull an archive with the correct extracted content", func() {

			destinationDir, err := ioutil.TempDir("/tmp", "porter-test-destination")
			if err != nil {
				log.Fatal(err)
			}
			defer os.RemoveAll(destinationDir)

			pullCommand := PullCommand{
				BucketURL:       bucketURL,
				SourcePath:      sourcePath,
				DestinationPath: destinationDir,
			}

			logger = lager.NewLogger("porter-pull-test")
			sink := lager.NewWriterSink(os.Stderr, lager.DEBUG)
			logger.RegisterSink(sink)

			err = pullCommand.Execute([]string{})
			Expect(err).ToNot(HaveOccurred())

			fileContents, err := ioutil.ReadFile(destinationDir + fileToPull.Name())
			Expect(string(fileContents)).To(Equal("hello world"))
			// invoke in
			// ensure the blobs match what we expect
		})

	})

})
