package k8s_test
//package k8s_test
//
//import (
//	"github.com/concourse/concourse/atc/k8s/k8sfakes"
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//)
//
//var _ = Describe("Pull", func() {
//	var (
//		fakeBlobRW *k8sfakes.FakeBlobstoreIO
//	)
//
//	BeforeEach(func() {
//		fakeBlobRW = new(k8sfakes.FakeBlobstoreIO)
//	})
//
//	Context("pulls from a blobstore", func() {
//		It("untars and populates into a volume", func() {
//			Expect(fakeBlobRW.InputBlobReaderCallCount()).To(Equal(1))
//		})
//	})
//})
//})