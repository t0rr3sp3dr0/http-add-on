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
	"github.com/kedacore/http-add-on/pkg/k8s"
)

var _ = Describe("TableMemory", func() {
	var (
		httpso0 = httpv1alpha1.HTTPScaledObject{
			ObjectMeta: metav1.ObjectMeta{
				Name: "keda-sh",
			},
			Spec: httpv1alpha1.HTTPScaledObjectSpec{
				Hosts: []string{
					"keda.sh",
				},
			},
		}

		httpso0NamespacedName = *k8s.NamespacedNameFromObject(&httpso0)

		httpso1 = httpv1alpha1.HTTPScaledObject{
			ObjectMeta: metav1.ObjectMeta{
				Name: "one-one-one-one",
			},
			Spec: httpv1alpha1.HTTPScaledObjectSpec{
				Hosts: []string{
					"1.1.1.1",
				},
			},
		}

		httpso1NamespacedName = *k8s.NamespacedNameFromObject(&httpso1)

		// TODO(pedrotorres): uncomment this when we support path prefix
		// httpsoList = httpv1alpha1.HTTPScaledObjectList{
		// 	Items: []httpv1alpha1.HTTPScaledObject{
		// 		{
		// 			ObjectMeta: metav1.ObjectMeta{
		// 				Name: "/",
		// 			},
		// 			Spec: httpv1alpha1.HTTPScaledObjectSpec{
		// 				Host: "localhost",
		// 				PathPrefix: "/",
		// 			},
		// 		},
		// 		{
		// 			ObjectMeta: metav1.ObjectMeta{
		// 				Name: "/f",
		// 			},
		// 			Spec: httpv1alpha1.HTTPScaledObjectSpec{
		// 				Host: "localhost",
		// 				PathPrefix: "/f",
		// 			},
		// 		},
		// 		{
		// 			ObjectMeta: metav1.ObjectMeta{
		// 				Name: "fo",
		// 			},
		// 			Spec: httpv1alpha1.HTTPScaledObjectSpec{
		// 				Host: "localhost",
		// 				PathPrefix: "fo",
		// 			},
		// 		},
		// 		{
		// 			ObjectMeta: metav1.ObjectMeta{
		// 				Name: "foo/",
		// 			},
		// 			Spec: httpv1alpha1.HTTPScaledObjectSpec{
		// 				Host: "localhost",
		// 				PathPrefix: "foo/",
		// 			},
		// 		},
		// 	},
		// }

		assertIndex = func(tm tableMemory, input *httpv1alpha1.HTTPScaledObject, expected *httpv1alpha1.HTTPScaledObject) {
			okMatcher := BeTrue()
			if expected == nil {
				okMatcher = BeFalse()
			}

			httpsoMatcher := Equal(expected)
			if expected == nil {
				httpsoMatcher = BeNil()
			}

			namespacedName := k8s.NamespacedNameFromObject(input)
			indexKey := newTableMemoryIndexKey(namespacedName)
			httpso, ok := tm.index.Get(indexKey)
			Expect(ok).To(okMatcher)
			Expect(httpso).To(httpsoMatcher)
		}

		assertStore = func(tm tableMemory, input *httpv1alpha1.HTTPScaledObject, expected *httpv1alpha1.HTTPScaledObject) {
			okMatcher := BeTrue()
			if expected == nil {
				okMatcher = BeFalse()
			}

			httpsoMatcher := Equal(expected)
			if expected == nil {
				httpsoMatcher = BeNil()
			}

			storeKeys := NewKeysFromHTTPSO(input)
			for _, storeKey := range storeKeys {
				httpso, ok := tm.store.Get(storeKey)
				Expect(ok).To(okMatcher)
				Expect(httpso).To(httpsoMatcher)
			}
		}

		assertTrees = func(tm tableMemory, input *httpv1alpha1.HTTPScaledObject, expected *httpv1alpha1.HTTPScaledObject) {
			assertIndex(tm, input, expected)
			assertStore(tm, input, expected)
		}

		insertIndex = func(tm tableMemory, httpso *httpv1alpha1.HTTPScaledObject) tableMemory {
			namespacedName := k8s.NamespacedNameFromObject(httpso)
			indexKey := newTableMemoryIndexKey(namespacedName)
			tm.index, _, _ = tm.index.Insert(indexKey, httpso)

			return tm
		}

		insertStore = func(tm tableMemory, httpso *httpv1alpha1.HTTPScaledObject) tableMemory {
			storeKeys := NewKeysFromHTTPSO(httpso)
			for _, storeKey := range storeKeys {
				tm.store, _, _ = tm.store.Insert(storeKey, httpso)
			}

			return tm
		}

		insertTrees = func(tm tableMemory, httpso *httpv1alpha1.HTTPScaledObject) tableMemory {
			tm = insertIndex(tm, httpso)
			tm = insertStore(tm, httpso)
			return tm
		}
	)

	Context("New", func() {
		It("returns a tableMemory with initialized tree", func() {
			i := NewTableMemory()

			tm, ok := i.(tableMemory)
			Expect(ok).To(BeTrue())
			Expect(tm.index).NotTo(BeNil())
			Expect(tm.store).NotTo(BeNil())
		})
	})

	Context("Remember", func() {
		It("returns a tableMemory with new object inserted", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)

			assertTrees(tm, &httpso0, &httpso0)
		})

		It("returns a tableMemory with new object inserted and other objects retained", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)
			tm = tm.Remember(&httpso1).(tableMemory)

			assertTrees(tm, &httpso1, &httpso1)
			assertTrees(tm, &httpso0, &httpso0)
		})

		It("returns a tableMemory with old object of same key replaced", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)

			httpso1 := *httpso0.DeepCopy()
			httpso1.Spec.TargetPendingRequests = pointer.Int32(1)
			tm = tm.Remember(&httpso1).(tableMemory)

			assertTrees(tm, &httpso0, &httpso1)
		})

		It("returns a tableMemory with old object of same key replaced and other objects retained", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = tm.Remember(&httpso0).(tableMemory)
			tm = tm.Remember(&httpso1).(tableMemory)

			httpso2 := *httpso1.DeepCopy()
			httpso2.Spec.TargetPendingRequests = pointer.Int32(1)
			tm = tm.Remember(&httpso2).(tableMemory)

			assertTrees(tm, &httpso1, &httpso2)
			assertTrees(tm, &httpso0, &httpso0)
		})

		It("returns a tableMemory with deep-copied object", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}

			httpso := *httpso0.DeepCopy()
			tm = tm.Remember(&httpso).(tableMemory)

			httpso.Spec.Hosts[0] += ".br"
			assertTrees(tm, &httpso0, &httpso0)
		})
	})

	Context("Forget", func() {
		It("returns a tableMemory with old object deleted", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)

			tm = tm.Forget(&httpso0NamespacedName).(tableMemory)

			assertTrees(tm, &httpso0, nil)
		})

		It("returns a tableMemory with old object deleted and other objects retained", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)
			tm = insertTrees(tm, &httpso1)

			tm = tm.Forget(&httpso0NamespacedName).(tableMemory)

			assertTrees(tm, &httpso1, &httpso1)
			assertTrees(tm, &httpso0, nil)
		})

		It("returns unchanged tableMemory when object is absent", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)

			index0 := *tm.index
			store0 := *tm.store
			tm = tm.Forget(&httpso1NamespacedName).(tableMemory)
			index1 := *tm.index
			store1 := *tm.store
			Expect(index1).To(Equal(index0))
			Expect(store1).To(Equal(store0))
		})
	})

	Context("Recall", func() {
		It("returns object with matching key", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)

			httpso := tm.Recall(&httpso0NamespacedName)
			Expect(httpso).To(Equal(&httpso0))
		})

		It("returns nil when object is absent", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)

			httpso := tm.Recall(&httpso1NamespacedName)
			Expect(httpso).To(BeNil())
		})

		It("returns deep-copied object", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)

			httpso := tm.Recall(&httpso0NamespacedName)
			Expect(httpso).To(Equal(&httpso0))

			httpso.Spec.Hosts[0] += ".br"

			assertTrees(tm, &httpso0, &httpso0)
		})
	})

	Context("Route", func() {
		It("returns nil when no matching host for URL", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)

			url, err := url.Parse(fmt.Sprintf("https://%s.br", httpso0.Spec.Hosts[0]))
			Expect(err).NotTo(HaveOccurred())
			Expect(url).NotTo(BeNil())
			urlKey := NewKeyFromURL(url)
			Expect(urlKey).NotTo(BeNil())
			httpso := tm.Route(urlKey)
			Expect(httpso).To(BeNil())
		})

		It("returns expected object with matching host for URL", func() {
			tm := tableMemory{
				index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
				store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
			}
			tm = insertTrees(tm, &httpso0)
			tm = insertTrees(tm, &httpso1)

			//goland:noinspection HttpUrlsUsage
			url0, err := url.Parse(fmt.Sprintf("http://%s", httpso0.Spec.Hosts[0]))
			Expect(err).NotTo(HaveOccurred())
			Expect(url0).NotTo(BeNil())
			url0Key := NewKeyFromURL(url0)
			Expect(url0Key).NotTo(BeNil())
			ret0 := tm.Route(url0Key)
			Expect(ret0).To(Equal(&httpso0))

			url1, err := url.Parse(fmt.Sprintf("https://%s:443/abc/def?123=456#789", httpso1.Spec.Hosts[0]))
			Expect(err).NotTo(HaveOccurred())
			Expect(url1).NotTo(BeNil())
			url1Key := NewKeyFromURL(url1)
			Expect(url1Key).NotTo(BeNil())
			ret1 := tm.Route(url1Key)
			Expect(ret1).To(Equal(&httpso1))
		})

		// TODO(pedrotorres): uncomment this when we support path prefix
		//
		// It("returns nil when no matching pathPrefix for URL", func() {
		// 	var (
		// 		httpsoFoo = httpsoList.Items[3]
		// 	)
		//
		// 	tm := tableMemory{
		// 		index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
		// 		store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
		// 	}
		// 	tm = insertTrees(tm, &httpsoFoo)
		//
		// 	//goland:noinspection HttpUrlsUsage
		// 	url, err := url.Parse(fmt.Sprintf("http://%s/bar%s", httpsoFoo.Spec.Host, httpsoFoo.Spec.PathPrefix))
		// 	Expect(err).NotTo(HaveOccurred())
		// 	Expect(url).NotTo(BeNil())
		// 	urlKey := NewKeyFromURL(url)
		// 	Expect(urlKey).NotTo(BeNil())
		// 	httpso := tm.Route(urlKey)
		// 	Expect(httpso).To(BeNil())
		// })
		//
		// It("returns expected object with matching pathPrefix for URL", func() {
		// 	tm := tableMemory{
		// 		index: iradix.New[*httpv1alpha1.HTTPScaledObject](),
		// 		store: iradix.New[*httpv1alpha1.HTTPScaledObject](),
		// 	}
		// 	for _, httpso := range httpsoList.Items {
		// 		httpso := httpso
		//
		// 		tm = insertTrees(tm, &httpso)
		// 	}
		//
		// 	for _, httpso := range httpsoList.Items {
		// 		url, err := url.Parse(fmt.Sprintf("https://%s/%s", httpso.Spec.Host, httpso.Spec.PathPrefix))
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(url).NotTo(BeNil())
		// 		urlKey := NewKeyFromURL(url)
		// 		Expect(urlKey).NotTo(BeNil())
		// 		ret := tm.Route(urlKey)
		// 		Expect(ret).To(Equal(&httpso))
		// 	}
		//
		// 	for _, httpso := range httpsoList.Items {
		// 		url, err := url.Parse(fmt.Sprintf("https://%s/%s/bar", httpso.Spec.Host, httpso.Spec.PathPrefix))
		// 		Expect(err).NotTo(HaveOccurred())
		// 		Expect(url).NotTo(BeNil())
		// 		urlKey := NewKeyFromURL(url)
		// 		Expect(urlKey).NotTo(BeNil())
		// 		ret := tm.Route(urlKey)
		// 		Expect(ret).To(Equal(&httpso))
		// 	}
		// })
	})

	Context("E2E", func() {
		It("succeeds", func() {
			tm := NewTableMemory()

			ret0 := tm.Recall(&httpso0NamespacedName)
			Expect(ret0).To(BeNil())

			tm = tm.Remember(&httpso0)

			ret1 := tm.Recall(&httpso0NamespacedName)
			Expect(ret1).To(Equal(&httpso0))

			tm = tm.Forget(&httpso0NamespacedName)

			ret2 := tm.Recall(&httpso0NamespacedName)
			Expect(ret2).To(BeNil())

			tm = tm.Remember(&httpso0)
			tm = tm.Remember(&httpso1)

			ret3 := tm.Recall(&httpso0NamespacedName)
			Expect(ret3).To(Equal(&httpso0))

			ret4 := tm.Recall(&httpso1NamespacedName)
			Expect(ret4).To(Equal(&httpso1))

			//goland:noinspection HttpUrlsUsage
			url0, err := url.Parse(fmt.Sprintf("http://%s:80?123=456#789", httpso0.Spec.Hosts[0]))
			Expect(err).NotTo(HaveOccurred())
			Expect(url0).NotTo(BeNil())

			url0Key := NewKeyFromURL(url0)
			Expect(url0Key).NotTo(BeNil())

			ret5 := tm.Route(url0Key)
			Expect(ret5).To(Equal(&httpso0))

			url1, err := url.Parse(fmt.Sprintf("https://user:pass@%s:443/abc/def", httpso1.Spec.Hosts[0]))
			Expect(err).NotTo(HaveOccurred())
			Expect(url1).NotTo(BeNil())

			url1Key := NewKeyFromURL(url1)
			Expect(url1Key).NotTo(BeNil())

			ret6 := tm.Route(url1Key)
			Expect(ret6).To(Equal(&httpso1))

			url2, err := url.Parse("http://0.0.0.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(url2).NotTo(BeNil())

			url2Key := NewKeyFromURL(url2)
			Expect(url2Key).NotTo(BeNil())

			ret7 := tm.Route(url2Key)
			Expect(ret7).To(BeNil())

			tm = tm.Forget(&httpso0NamespacedName)

			ret8 := tm.Route(url0Key)
			Expect(ret8).To(BeNil())

			httpso := *httpso1.DeepCopy()
			httpso.Spec.TargetPendingRequests = pointer.Int32(1)

			tm = tm.Remember(&httpso)

			ret9 := tm.Route(url1Key)
			Expect(ret9).To(Equal(&httpso))
		})
	})
})