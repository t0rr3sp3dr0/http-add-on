package routing

import (
	"fmt"
	"net/url"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	httpv1alpha1 "github.com/kedacore/http-add-on/operator/apis/http/v1alpha1"
)

var _ = Describe("TableMemory", func() {
	var (
		httpso0 = httpv1alpha1.HTTPScaledObject{
			ObjectMeta: metav1.ObjectMeta{
				Name: "keda-sh",
			},
			Spec: httpv1alpha1.HTTPScaledObjectSpec{
				Host: "keda.sh",
			},
		}

		httpso1 = httpv1alpha1.HTTPScaledObject{
			ObjectMeta: metav1.ObjectMeta{
				Name: "one-one-one-one",
			},
			Spec: httpv1alpha1.HTTPScaledObjectSpec{
				Host: "1.1.1.1",
			},
		}

		httpsoList = httpv1alpha1.HTTPScaledObjectList{
			Items: []httpv1alpha1.HTTPScaledObject{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "/",
					},
					Spec: httpv1alpha1.HTTPScaledObjectSpec{
						Host:       "localhost",
						PathPrefix: "/",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "/f",
					},
					Spec: httpv1alpha1.HTTPScaledObjectSpec{
						Host:       "localhost",
						PathPrefix: "/f",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "fo",
					},
					Spec: httpv1alpha1.HTTPScaledObjectSpec{
						Host:       "localhost",
						PathPrefix: "fo",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo/",
					},
					Spec: httpv1alpha1.HTTPScaledObjectSpec{
						Host:       "localhost",
						PathPrefix: "foo/",
					},
				},
			},
		}
	)

	Context("New", func() {
		It("returns a tableMemory with initialized tree", func() {
			i := NewTableMemory()

			tm, ok := i.(tableMemory)
			Expect(ok).To(BeTrue())
			Expect(tm.tree).NotTo(BeNil())
		})
	})

	Context("Remember", func() {
		It("returns a tableMemory with new object inserted", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)

			key := tm.treeKeyForHTTPSO(&httpso0)
			httpso, ok := tm.tree.Get(key)
			Expect(ok).To(BeTrue())
			Expect(httpso).To(Equal(&httpso0))
		})

		It("returns a tableMemory with new object inserted and other objects retained", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)
			tm = tm.Remember(&httpso1).(tableMemory)

			key1 := tm.treeKeyForHTTPSO(&httpso1)
			ret1, ok := tm.tree.Get(key1)
			Expect(ok).To(BeTrue())
			Expect(ret1).To(Equal(&httpso1))

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			ret0, ok := tm.tree.Get(key0)
			Expect(ok).To(BeTrue())
			Expect(ret0).To(Equal(&httpso0))
		})

		It("returns a tableMemory with old object of same key replaced", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)

			httpso1 := httpso0
			httpso1.Spec.TargetPendingRequests = pointer.Int32(1)
			tm = tm.Remember(&httpso1).(tableMemory)

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			ret0, ok := tm.tree.Get(key0)
			Expect(ok).To(BeTrue())
			Expect(ret0).To(Equal(&httpso1))
		})

		It("returns a tableMemory with old object of same key replaced and other objects retained", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)
			tm = tm.Remember(&httpso1).(tableMemory)

			httpso2 := httpso1
			httpso2.Spec.TargetPendingRequests = pointer.Int32(1)
			tm = tm.Remember(&httpso2).(tableMemory)

			key1 := tm.treeKeyForHTTPSO(&httpso1)
			ret1, ok := tm.tree.Get(key1)
			Expect(ok).To(BeTrue())
			Expect(ret1).To(Equal(&httpso2))

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			ret0, ok := tm.tree.Get(key0)
			Expect(ok).To(BeTrue())
			Expect(ret0).To(Equal(&httpso0))
		})
	})

	Context("Forget", func() {
		It("returns a tableMemory with old object deleted", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			key := tm.treeKeyForHTTPSO(&httpso0)
			tm.tree, _, _ = tm.tree.Insert(key, &httpso0)

			tm = tm.Forget(&httpso0).(tableMemory)

			httpso, ok := tm.tree.Get(key)
			Expect(ok).To(BeFalse())
			Expect(httpso).To(BeNil())
		})

		It("returns a tableMemory with old object deleted and other objects retained", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			tm.tree, _, _ = tm.tree.Insert(key0, &httpso0)

			key1 := tm.treeKeyForHTTPSO(&httpso1)
			tm.tree, _, _ = tm.tree.Insert(key1, &httpso1)

			tm = tm.Forget(&httpso0).(tableMemory)

			ret1, ok := tm.tree.Get(key1)
			Expect(ok).To(BeTrue())
			Expect(ret1).To(Equal(&httpso1))

			ret0, ok := tm.tree.Get(key0)
			Expect(ok).To(BeFalse())
			Expect(ret0).To(BeNil())
		})

		It("returns unchanged tableMemory when object is absent", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			tm.tree, _, _ = tm.tree.Insert(key0, &httpso0)

			tree0 := *tm.tree
			tm = tm.Forget(&httpso1).(tableMemory)
			tree1 := *tm.tree
			Expect(tree1).To(Equal(tree0))
		})
	})

	Context("Recall", func() {
		It("returns object with matching key", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			tm.tree, _, _ = tm.tree.Insert(key0, &httpso0)

			httpso1 := httpso0
			httpso1.Spec.TargetPendingRequests = pointer.Int32(1)

			httpso := tm.Recall(&httpso1)
			Expect(httpso).To(Equal(&httpso0))
		})

		It("returns nil when object is absent", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			tm.tree, _, _ = tm.tree.Insert(key0, &httpso0)

			httpso := tm.Recall(&httpso1)
			Expect(httpso).To(BeNil())
		})
	})

	Context("Route", func() {
		It("returns nil when no matching host for URL", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			key := tm.treeKeyForHTTPSO(&httpso0)
			tm.tree, _, _ = tm.tree.Insert(key, &httpso0)

			url, err := url.Parse(fmt.Sprintf("https://%s.br", httpso0.Spec.Host))
			Expect(err).NotTo(HaveOccurred())
			Expect(url).NotTo(BeNil())
			httpso := tm.Route(url)
			Expect(httpso).To(BeNil())
		})

		It("returns expected object with matching host for URL", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			key0 := tm.treeKeyForHTTPSO(&httpso0)
			tm.tree, _, _ = tm.tree.Insert(key0, &httpso0)

			key1 := tm.treeKeyForHTTPSO(&httpso1)
			tm.tree, _, _ = tm.tree.Insert(key1, &httpso1)

			//goland:noinspection HttpUrlsUsage
			url0, err := url.Parse(fmt.Sprintf("http://%s", httpso0.Spec.Host))
			Expect(err).NotTo(HaveOccurred())
			Expect(url0).NotTo(BeNil())
			ret0 := tm.Route(url0)
			Expect(ret0).To(Equal(&httpso0))

			url1, err := url.Parse(fmt.Sprintf("https://%s:443/abc/def?123=456#789", httpso1.Spec.Host))
			Expect(err).NotTo(HaveOccurred())
			Expect(url1).NotTo(BeNil())
			ret1 := tm.Route(url1)
			Expect(ret1).To(Equal(&httpso1))
		})

		It("returns nil when no matching pathPrefix for URL", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			httpsoFoo := httpsoList.Items[3]
			key := tm.treeKeyForHTTPSO(&httpsoFoo)
			tm.tree, _, _ = tm.tree.Insert(key, &httpsoFoo)

			//goland:noinspection HttpUrlsUsage
			url, err := url.Parse(fmt.Sprintf("http://%s/bar%s", httpsoFoo.Spec.Host, httpsoFoo.Spec.PathPrefix))
			Expect(err).NotTo(HaveOccurred())
			Expect(url).NotTo(BeNil())
			httpso := tm.Route(url)
			Expect(httpso).To(BeNil())
		})

		It("returns expected object with matching pathPrefix for URL", func() {
			tm := tableMemory{
				tree: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			for _, httpso := range httpsoList.Items {
				httpso := httpso

				key := tm.treeKeyForHTTPSO(&httpso)
				tm.tree, _, _ = tm.tree.Insert(key, &httpso)
			}

			for _, httpso := range httpsoList.Items {
				url, err := url.Parse(fmt.Sprintf("https://%s/%s", httpso.Spec.Host, httpso.Spec.PathPrefix))
				Expect(err).NotTo(HaveOccurred())
				Expect(url).NotTo(BeNil())
				ret := tm.Route(url)
				Expect(ret).To(Equal(&httpso))
			}

			for _, httpso := range httpsoList.Items {
				url, err := url.Parse(fmt.Sprintf("https://%s/%s/bar", httpso.Spec.Host, httpso.Spec.PathPrefix))
				Expect(err).NotTo(HaveOccurred())
				Expect(url).NotTo(BeNil())
				ret := tm.Route(url)
				Expect(ret).To(Equal(&httpso))
			}
		})
	})

	Context("treeKeyForURL", func() {
		It("returns expected key for URL", func() {
			const (
				host = "kubernetes.io"
				path = "abc/def"
				norm = "//kubernetes.io/abc/def/"
			)

			var tm tableMemory

			url, err := url.Parse(fmt.Sprintf("https://%s:443/%s?123=456#789", host, path))
			Expect(err).NotTo(HaveOccurred())
			Expect(url).NotTo(BeNil())

			key := tm.treeKeyForURL(url)
			Expect(key).To(Equal([]byte(norm)))
		})

		It("returns nil for nil URL", func() {
			var tm tableMemory

			key := tm.treeKeyForURL(nil)
			Expect(key).To(BeNil())
		})
	})

	Context("treeKeyForHTTPSO", func() {
		It("returns expected key for HTTPSO", func() {
			const (
				host = "kubernetes.io"
				path = "abc/def"
				norm = "//kubernetes.io/abc/def/"
			)

			var tm tableMemory

			key := tm.treeKeyForHTTPSO(&httpv1alpha1.HTTPScaledObject{
				Spec: httpv1alpha1.HTTPScaledObjectSpec{
					Host:       host,
					PathPrefix: path,
				},
			})
			Expect(key).To(Equal([]byte(norm)))
		})

		It("returns nil for nil HTTPSO", func() {
			var tm tableMemory

			key := tm.treeKeyForHTTPSO(nil)
			Expect(key).To(BeNil())
		})
	})

	Context("treeKey", func() {
		const (
			host0 = "kubernetes.io"
			host1 = "kubernetes.io:443"
			path0 = "abc/def"
			path1 = "abc/def/"
			path2 = "abc/def//"
			path3 = "/abc/def"
			path4 = "/abc/def/"
			path5 = "/abc/def//"
			path6 = "//abc/def"
			path7 = "//abc/def/"
			path8 = "//abc/def//"
			norm0 = "///"
			norm1 = "//kubernetes.io/"
			norm2 = "///abc/def/"
			norm3 = "//kubernetes.io/abc/def/"
		)

		It("returns expected key for blank host and blank path", func() {
			var tm tableMemory

			key := tm.treeKey("", "")
			Expect(key).To(Equal([]byte(norm0)))
		})

		It("returns expected key for host without port", func() {
			var tm tableMemory

			key := tm.treeKey(host0, "")
			Expect(key).To(Equal([]byte(norm1)))
		})

		It("returns expected key for host with port", func() {
			var tm tableMemory

			key := tm.treeKey(host1, "")
			Expect(key).To(Equal([]byte(norm1)))
		})

		It("returns expected key for path with no leading slashes and no trailing slashes", func() {
			var tm tableMemory

			key := tm.treeKey("", path0)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with no leading slashes and single trailing slash", func() {
			var tm tableMemory

			key := tm.treeKey("", path1)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with no leading slashes and multiple trailing slashes", func() {
			var tm tableMemory

			key := tm.treeKey("", path2)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with single leading slashes and no trailing slashes", func() {
			var tm tableMemory

			key := tm.treeKey("", path3)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with single leading slash and single trailing slash", func() {
			var tm tableMemory

			key := tm.treeKey("", path4)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with single leading slash and multiple trailing slashes", func() {
			var tm tableMemory

			key := tm.treeKey("", path5)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with multiple leading slashes and no trailing slashes", func() {
			var tm tableMemory

			key := tm.treeKey("", path6)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with multiple leading slash and single trailing slash", func() {
			var tm tableMemory

			key := tm.treeKey("", path7)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for path with multiple leading slash and multiple trailing slashes", func() {
			var tm tableMemory

			key := tm.treeKey("", path8)
			Expect(key).To(Equal([]byte(norm2)))
		})

		It("returns expected key for non-blank host and non-blank path", func() {
			var tm tableMemory

			key := tm.treeKey(host1, path8)
			Expect(key).To(Equal([]byte(norm3)))
		})

		It("returns nil for nil HTTPSO", func() {
			var tm tableMemory

			key := tm.treeKeyForHTTPSO(nil)
			Expect(key).To(BeNil())
		})
	})

	Context("E2E", func() {
		It("succeeds", func() {
			tm := NewTableMemory()

			ret0 := tm.Recall(&httpso0)
			Expect(ret0).To(BeNil())

			tm = tm.Remember(&httpso0)

			ret1 := tm.Recall(&httpso0)
			Expect(ret1).To(Equal(&httpso0))

			tm = tm.Forget(&httpso0)

			ret2 := tm.Recall(&httpso0)
			Expect(ret2).To(BeNil())

			tm = tm.Remember(&httpso0)
			tm = tm.Remember(&httpso1)

			ret3 := tm.Recall(&httpso0)
			Expect(ret3).To(Equal(&httpso0))

			ret4 := tm.Recall(&httpso1)
			Expect(ret4).To(Equal(&httpso1))

			//goland:noinspection HttpUrlsUsage
			url0, err := url.Parse(fmt.Sprintf("http://%s:80?123=456#789", httpso0.Spec.Host))
			Expect(err).NotTo(HaveOccurred())
			Expect(url0).NotTo(BeNil())

			ret5 := tm.Route(url0)
			Expect(ret5).To(Equal(&httpso0))

			url1, err := url.Parse(fmt.Sprintf("https://user:pass@%s:443/abc/def", httpso1.Spec.Host))
			Expect(err).NotTo(HaveOccurred())
			Expect(url1).NotTo(BeNil())

			ret6 := tm.Route(url1)
			Expect(ret6).To(Equal(&httpso1))

			url2, err := url.Parse("http://0.0.0.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(url2).NotTo(BeNil())

			ret7 := tm.Route(url2)
			Expect(ret7).To(BeNil())

			tm = tm.Forget(&httpso0)

			ret8 := tm.Route(url0)
			Expect(ret8).To(BeNil())

			httpso := httpso1
			httpso.Spec.TargetPendingRequests = pointer.Int32(1)

			tm = tm.Remember(&httpso)

			ret9 := tm.Route(url1)
			Expect(ret9).To(Equal(&httpso))
		})
	})
})
