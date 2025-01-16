package installercache

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	eventsapi "github.com/openshift/assisted-service/internal/events/api"
	"github.com/openshift/assisted-service/internal/metrics"
	"github.com/openshift/assisted-service/internal/oc"
	"github.com/openshift/assisted-service/models"
	"github.com/sirupsen/logrus"
)

var _ = Describe("release event", func() {

	var (
		ctx           context.Context
		ctrl          *gomock.Controller
		eventsHandler *eventsapi.MockHandler
	)

	BeforeEach(func() {
		ctx = context.TODO()
		ctrl = gomock.NewController(GinkgoT())
		eventsHandler = eventsapi.NewMockHandler(ctrl)
	})

	It("should send correct fields when release event is sent", func() {
		startTime, err := time.Parse(time.RFC3339, "2025-01-07T16:51:10Z")
		Expect(err).ToNot(HaveOccurred())
		clusterID := strfmt.UUID(uuid.NewString())
		releaseID := "quay.io/openshift-release-dev/ocp-release:4.16.28-x86_64"
		r := &Release{
			eventsHandler:   eventsHandler,
			startTime:       startTime,
			clusterID:       clusterID,
			releaseID:       releaseID,
			cached:          false,
			extractDuration: 19.5,
		}
		eventsHandler.EXPECT().V2AddMetricsEvent(
			ctx, &clusterID, nil, nil, "", models.EventSeverityInfo, metricEventInstallerCacheRelease, gomock.Any(),
			"release_id", r.releaseID,
			"start_time", "2025-01-07T16:51:10Z",
			"end_time", gomock.Any(),
			"cached", r.cached,
			"extract_duration", r.extractDuration,
		).Times(1)
		r.Cleanup(ctx)
	})
})

var _ = Describe("installer cache", func() {
	var (
		ctrl            *gomock.Controller
		mockRelease     *oc.MockRelease
		manager         *Installers
		cacheDir        string
		eventsHandler   *eventsapi.MockHandler
		ctx             context.Context
		diskStatsHelper metrics.DiskStatsHelper
		cacheSize       int64
		maxReleaseSize  int64
	)

	BeforeEach(func() {

		ctrl = gomock.NewController(GinkgoT())
		diskStatsHelper = metrics.NewOSDiskStatsHelper()
		mockRelease = oc.NewMockRelease(ctrl)
		eventsHandler = eventsapi.NewMockHandler(ctrl)

		var err error
		cacheDir, err = os.MkdirTemp("/tmp", "cacheDir")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Mkdir(filepath.Join(cacheDir, "quay.io"), 0755)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(filepath.Join(cacheDir, "quay.io"), "release-dev"), 0755)).To(Succeed())
		cacheSize = int64(12)
		maxReleaseSize = int64(5)
		manager, err = New(cacheDir, cacheSize, maxReleaseSize, eventsHandler, logrus.New(), diskStatsHelper)
		Expect(err).NotTo(HaveOccurred())
		ctx = context.TODO()
	})

	AfterEach(func() {
		os.RemoveAll(cacheDir)
	})

	expectEventsSent := func() {
		eventsHandler.EXPECT().V2AddMetricsEvent(
			ctx, gomock.Any(), nil, nil, "", models.EventSeverityInfo, metricEventInstallerCacheRelease, gomock.Any(),
			gomock.Any(),
		).AnyTimes()
	}

	mockReleaseCalls := func(releaseID string, version string) {
		workdir := filepath.Join(cacheDir, "quay.io", "release-dev")
		fname := filepath.Join(workdir, releaseID)

		mockRelease.EXPECT().GetReleaseBinaryPath(
			gomock.Any(), gomock.Any(), version).
			Return(workdir, releaseID, fname, nil).AnyTimes()

		mockRelease.EXPECT().Extract(gomock.Any(), releaseID,
			gomock.Any(), cacheDir, gomock.Any(), version).
			DoAndReturn(func(log logrus.FieldLogger, releaseImage string, releaseImageMirror string, cacheDir string, pullSecret string, version string) (string, error) {
				dir, _ := filepath.Split(fname)
				err := os.MkdirAll(dir, 0700)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(10 * time.Millisecond) // Add a small amount of latency for file extraction
				err = os.WriteFile(fname, []byte("abcde"), 0600)
				return "", err
			}).AnyTimes()
	}

	testGet := func(releaseID, version string, clusterID strfmt.UUID, expectCached bool) (string, string) {
		workdir := filepath.Join(cacheDir, "quay.io", "release-dev")
		fname := filepath.Join(workdir, releaseID)
		mockReleaseCalls(releaseID, version)
		expectEventsSent()
		l, err := manager.Get(ctx, releaseID, "mirror", "pull-secret", mockRelease, version, clusterID)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(l.releaseID).To(Equal(releaseID))
		Expect(l.clusterID).To(Equal(clusterID))
		Expect(l.startTime).ShouldNot(BeZero())
		Expect(l.cached).To(Equal(expectCached))
		Expect(l.eventsHandler).To(Equal(eventsHandler))
		if !expectCached {
			Expect(l.extractDuration).ShouldNot(BeZero())
		}
		Expect(l.Path).ShouldNot(BeEmpty())

		time.Sleep(10 * time.Millisecond)
		Expect(l.startTime.Before(time.Now())).To(BeTrue())
		l.Cleanup(context.TODO())
		return fname, l.Path
	}

	type launchParams struct {
		releaseID          string
		version            string
		clusterID          strfmt.UUID
		autoCleanupRelease bool
	}

	// launchParallel launches a batch of fetches from the installer cache in order to simulate multiple requests to the same node
	// releases are automatically released as they are gathered.
	runParallelTest := func(params []launchParams) error {
		var wg sync.WaitGroup
		var firstError error
		var errMutex sync.Mutex
		for _, param := range params {
			mockReleaseCalls(param.releaseID, param.version)
			wg.Add(1)
			go func(param launchParams) {
				defer wg.Done()
				release, err := manager.Get(ctx, param.releaseID, "mirror", "pull-secret", mockRelease, param.version, param.clusterID)
				if err != nil {
					errMutex.Lock()
					if firstError == nil {
						firstError = err
					}
					errMutex.Unlock()
					return
				}
				// Simulate calls to release cleanup to ensure that we can test capacity limits
				// Add a very small delay so that we can simulate usage of the release prior to release.
				if param.autoCleanupRelease && release != nil {
					time.Sleep(2 * time.Millisecond)
					release.Cleanup(context.TODO())
				}
			}(param)
		}
		wg.Wait()
		return firstError
	}

	getUsedBytesForDirectory := func(directory string) uint64 {
		var totalBytes uint64
		seenInodes := make(map[uint64]bool)
		err := filepath.Walk(directory, func(path string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				if _, ok := err.(*os.PathError); ok {
					// Something deleted the file before we could walk it
					// count this as zero bytes
					return nil
				}
				return err
			}
			stat, ok := fileInfo.Sys().(*syscall.Stat_t)
			Expect(ok).To(BeTrue())
			if !fileInfo.IsDir() && !seenInodes[stat.Ino] {
				totalBytes += uint64(fileInfo.Size())
				seenInodes[stat.Ino] = true
			}
			return nil
		})
		Expect(err).ToNot(HaveOccurred())
		return totalBytes
	}

	It("Should raise error on construction if max release size is larger than cache and cache is enabled", func() {
		cacheSize = int64(5)
		maxReleaseSize = int64(10)
		var err error
		manager, err = New(cacheDir, cacheSize, maxReleaseSize, eventsHandler, logrus.New(), diskStatsHelper)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("maxReleaseSize (10 bytes) must not be less than storageCapacity (5 bytes)"))
	})

	It("Should not raise error on construction if max release size is larger than cache and cache is disabled", func() {
		cacheSize = int64(0)
		maxReleaseSize = int64(10)
		var err error
		manager, err = New(cacheDir, cacheSize, maxReleaseSize, eventsHandler, logrus.New(), diskStatsHelper)
		Expect(err).ToNot(HaveOccurred())
	})

	It("when cache limit is zero - eviction is skipped", func() {
		manager.storageCapacity = 0
		clusterId := strfmt.UUID(uuid.New().String())
		r1, _ := testGet("4.8", "4.8.0", clusterId, false)
		r2, _ := testGet("4.9", "4.9.0", clusterId, false)
		r3, _ := testGet("4.10", "4.10.0", clusterId, false)

		By("verify that the no file was deleted")
		_, err := os.Stat(r1)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r2)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r3)
		Expect(os.IsNotExist(err)).To(BeFalse())
	})

	It("exising files access time is updated", func() {
		clusterId := strfmt.UUID(uuid.New().String())
		_, _ = testGet("4.8", "4.8.0", clusterId, false)
		r2, _ := testGet("4.9", "4.9.0", clusterId, false)
		r1, _ := testGet("4.8", "4.8.0", clusterId, true)
		r3, _ := testGet("4.10", "4.10.0", clusterId, false)

		By("verify that the oldest file was deleted")
		_, err := os.Stat(r1)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r2)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(r3)
		Expect(os.IsNotExist(err)).To(BeFalse())
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(cacheDir)).To(BeNumerically("<=", cacheSize))
	})

	It("evicts the oldest file", func() {
		clusterId := strfmt.UUID(uuid.New().String())
		r1, l1 := testGet("4.8", "4.8.0", clusterId, false)
		r2, l2 := testGet("4.9", "4.9.0", clusterId, false)
		r3, l3 := testGet("4.10", "4.10.0", clusterId, false)

		By("verify that the oldest file was deleted")
		_, err := os.Stat(r1)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(r2)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r3)
		Expect(os.IsNotExist(err)).To(BeFalse())

		By("verify that the links were purged")
		manager.evict(manager.maxReleaseSize)
		_, err = os.Stat(l1)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(l2)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(l3)
		Expect(os.IsNotExist(err)).To(BeTrue())
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(cacheDir)).To(BeNumerically("<=", cacheSize))
	})

	It("extracts from the mirror", func() {
		releaseID := "4.10-orig"
		releaseMirrorID := "4.10-mirror"
		clusterID := strfmt.UUID(uuid.New().String())
		version := "4.10.0"
		mockReleaseCalls(releaseID, version)
		l, err := manager.Get(ctx, releaseID, releaseMirrorID, "pull-secret", mockRelease, version, clusterID)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(l.releaseID).To(Equal(releaseID))
		Expect(l.clusterID).To(Equal(clusterID))
		Expect(l.startTime).ShouldNot(BeZero())
		Expect(l.cached).To(BeFalse())
		Expect(l.extractDuration).ShouldNot(BeZero())
		Expect(l.Path).ShouldNot(BeEmpty())
		expectEventsSent()
		l.Cleanup(ctx)
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(cacheDir)).To(BeNumerically("<=", cacheSize))
	})

	It("extracts without a mirror", func() {
		releaseID := "4.10-orig"
		releaseMirrorID := ""
		version := "4.10.0"
		clusterID := strfmt.UUID(uuid.NewString())
		mockReleaseCalls(releaseID, version)
		l, err := manager.Get(ctx, releaseID, releaseMirrorID, "pull-secret", mockRelease, version, clusterID)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(l.releaseID).To(Equal(releaseID))
		Expect(l.clusterID).To(Equal(clusterID))
		Expect(l.startTime).ShouldNot(BeZero())
		Expect(l.cached).To(BeFalse())
		Expect(l.extractDuration).ShouldNot(BeZero())
		Expect(l.Path).ShouldNot(BeEmpty())
		expectEventsSent()
		l.Cleanup(ctx)
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(cacheDir)).To(BeNumerically("<=", cacheSize))
	})

	It("should correctly handle multiple requests for the same release at the same time", func() {
		cacheSize = int64(20)
		maxReleaseSize = int64(5)
		var err error
		manager, err = New(cacheDir, cacheSize, maxReleaseSize, eventsHandler, logrus.New(), diskStatsHelper)
		Expect(err).NotTo(HaveOccurred())
		params := []launchParams{}
		for i := 0; i < 10; i++ {
			params = append(params, launchParams{releaseID: "4.17.11-x86_64", version: "4.17.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
		}
		expectEventsSent()
		err = runParallelTest(params)
		Expect(err).ToNot(HaveOccurred())
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(cacheDir)).To(BeNumerically("<=", cacheSize))
	})

	It("should consistently handle multiple requests for different releases at the same time", func() {
		cacheSize = int64(25)
		maxReleaseSize = int64(5)
		var err error
		manager, err = New(cacheDir, cacheSize, maxReleaseSize, eventsHandler, logrus.New(), diskStatsHelper)
		Expect(err).NotTo(HaveOccurred())
		for i := 0; i < 100; i++ {
			params := []launchParams{}
			params = append(params, launchParams{releaseID: "4.17.11-x86_64", version: "4.17.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
			params = append(params, launchParams{releaseID: "4.18.11-x86_64", version: "4.18.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
			params = append(params, launchParams{releaseID: "4.19.11-x86_64", version: "4.19.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
			params = append(params, launchParams{releaseID: "4.20.11-x86_64", version: "4.20.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
			params = append(params, launchParams{releaseID: "4.21.11-x86_64", version: "4.21.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
			expectEventsSent()
			err = runParallelTest(params)
			Expect(err).ToNot(HaveOccurred())
		}
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(cacheDir)).To(BeNumerically("<=", cacheSize))
	})

	It("should raise an error if unable to write within storage limit", func() {
		cacheSize = int64(5)
		maxReleaseSize = int64(5)
		var err error
		manager, err = New(cacheDir, cacheSize, maxReleaseSize, eventsHandler, logrus.New(), diskStatsHelper)
		Expect(err).NotTo(HaveOccurred())
		expectEventsSent()
		// Force a scenario where one of the requests will fail because of a pending release cleanup
		params := []launchParams{}
		params = append(params, launchParams{releaseID: "4.17.11-x86_64", version: "4.17.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
		params = append(params, launchParams{releaseID: "4.18.11-x86_64", version: "4.18.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
		err = runParallelTest(params)
		Expect(err).To(BeAssignableToTypeOf(&ErrorInsufficientCacheCapacity{}))
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(cacheDir)).To(BeNumerically("<=", cacheSize))

		// After we have collected all results from the last parallel test, the cleanup should have occurred as part of the test
		// Now assert that a retry would work, there should be enough space for another release
		// Using a brand new release ID to prove we are not hitting cache here.
		params = []launchParams{}
		params = append(params, launchParams{releaseID: "4.19.11-x86_64", version: "4.19.11", clusterID: strfmt.UUID(uuid.NewString()), autoCleanupRelease: true})
		err = runParallelTest(params)
		Expect(err).ToNot(HaveOccurred())
	})
})

func TestInstallerCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "installercache tests")
}
