package installercache

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
		metricsAPI      *metrics.MockAPI
		ctx             context.Context
		diskStatsHelper metrics.DiskStatsHelper
	)

	BeforeEach(func() {

		ctrl = gomock.NewController(GinkgoT())
		diskStatsHelper = metrics.NewOSDiskStatsHelper(logrus.New())
		mockRelease = oc.NewMockRelease(ctrl)
		eventsHandler = eventsapi.NewMockHandler(ctrl)
		metricsAPI = metrics.NewMockAPI(ctrl)
		var err error
		cacheDir, err = os.MkdirTemp("/tmp", "cacheDir")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Mkdir(filepath.Join(cacheDir, "quay.io"), 0755)).To(Succeed())
		Expect(os.Mkdir(filepath.Join(filepath.Join(cacheDir, "quay.io"), "release-dev"), 0755)).To(Succeed())
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(12),
			MaxReleaseSize:                  int64(5),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 1,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
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
		workdir := filepath.Join(manager.config.CacheDir, "quay.io", "release-dev")
		fname := filepath.Join(workdir, releaseID)

		mockRelease.EXPECT().GetReleaseBinaryPath(
			gomock.Any(), gomock.Any(), version).
			Return(workdir, releaseID, fname, nil).AnyTimes()

		mockRelease.EXPECT().Extract(gomock.Any(), releaseID,
			gomock.Any(), manager.config.CacheDir, gomock.Any(), version).
			DoAndReturn(func(log logrus.FieldLogger, releaseImage string, releaseImageMirror string, cacheDir string, pullSecret string, version string) (string, error) {
				dir, _ := filepath.Split(fname)
				err := os.MkdirAll(dir, 0700)
				Expect(err).ToNot(HaveOccurred())
				time.Sleep(10 * time.Millisecond) // Add a small amount of latency to simulate extraction
				err = os.WriteFile(fname, []byte("abcde"), 0600)
				return "", err
			}).AnyTimes()
	}

	testGet := func(releaseID, version string, clusterID strfmt.UUID, expectCached bool) (string, string) {
		workdir := filepath.Join(manager.config.CacheDir, "quay.io", "release-dev")
		fname := filepath.Join(workdir, releaseID)
		mockReleaseCalls(releaseID, version)
		expectEventsSent()
		metricsAPI.EXPECT().InstallerCacheReleaseExtracted(releaseID).Times(1)
		metricsAPI.EXPECT().InstallerCacheGetReleaseOK().Times(1)
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
		releaseID string
		version   string
		clusterID strfmt.UUID
	}

	// runParallelTest launches a batch of fetches from the installer cache in order to simulate multiple requests to the same node
	// releases are automatically cleaned up as they are gathered
	// returns the first error encountered or nil if no error encountered.
	runParallelTest := func(params []launchParams) error {
		var wg sync.WaitGroup
		var firstError error
		var errMutex sync.Mutex
		for _, param := range params {
			mockReleaseCalls(param.releaseID, param.version)
			metricsAPI.EXPECT().InstallerCacheReleaseCached(param.releaseID).AnyTimes()
			metricsAPI.EXPECT().InstallerCacheReleaseExtracted(param.releaseID).AnyTimes()
			metricsAPI.EXPECT().InstallerCacheReleaseExtracted(param.releaseID).AnyTimes()
			wg.Add(1)
			go func(param launchParams) {
				defer wg.Done()
				metricsAPI.EXPECT().InstallerCacheGetReleaseOK().Times(1)
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
				if release != nil {
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
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(5),
			MaxReleaseSize:                  int64(10),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("config.MaxReleaseSize (10 bytes) must not be greater than config.MaxCapacity (5 bytes)"))
	})

	It("Should raise error on construction if max release size is zero and cache is enabled", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(5),
			MaxReleaseSize:                  int64(0),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("config.MaxReleaseSize (0 bytes) must not be zero"))
	})

	It("Should not raise error on construction if max release size is larger than cache and cache eviction is disabled", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(0),
			MaxReleaseSize:                  int64(10),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should not raise error on construction if max release size is zero and cache eviction is disabled", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(0),
			MaxReleaseSize:                  int64(0),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should not accept InstallerCacheEvictionThreshold of zero", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(0),
			MaxReleaseSize:                  int64(0),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("config.InstallerCacheEvictionThreshold must not be zero"))
	})

	It("when cache limit is zero - eviction is skipped", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(0),
			MaxReleaseSize:                  int64(5),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).ToNot(HaveOccurred())
		clusterId := strfmt.UUID(uuid.New().String())
		r1, l1 := testGet("4.8", "4.8.0", clusterId, false)
		r2, l2 := testGet("4.9", "4.9.0", clusterId, false)
		r3, l3 := testGet("4.10", "4.10.0", clusterId, false)

		By("verify that the no file was deleted")
		_, err = os.Stat(r1)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r2)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r3)
		Expect(os.IsNotExist(err)).To(BeFalse())

		By("verify that the links were purged")
		manager.evict()
		_, err = os.Stat(l1)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(l2)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(l3)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("exising files access time is updated", func() {
		clusterId := strfmt.UUID(uuid.New().String())
		_, _ = testGet("4.8", "4.8.0", clusterId, false)
		r2, _ := testGet("4.9", "4.9.0", clusterId, false)
		metricsAPI.EXPECT().InstallerCacheReleaseCached("4.8").Times(1)
		r1, _ := testGet("4.8", "4.8.0", clusterId, true)
		metricsAPI.EXPECT().InstallerCacheTryEviction().Times(1)
		metricsAPI.EXPECT().InstallerCacheReleaseEvicted().Times(1)
		r3, _ := testGet("4.10", "4.10.0", clusterId, false)

		By("verify that the oldest file was deleted")
		_, err := os.Stat(r1)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r2)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(r3)
		Expect(os.IsNotExist(err)).To(BeFalse())
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", manager.config.MaxCapacity))
	})

	It("evicts the oldest file", func() {
		clusterId := strfmt.UUID(uuid.New().String())
		r1, l1 := testGet("4.8", "4.8.0", clusterId, false)
		r2, l2 := testGet("4.9", "4.9.0", clusterId, false)
		metricsAPI.EXPECT().InstallerCacheTryEviction().Times(1)
		metricsAPI.EXPECT().InstallerCacheReleaseEvicted().Times(1)
		r3, l3 := testGet("4.10", "4.10.0", clusterId, false)

		By("verify that the oldest file was deleted")
		_, err := os.Stat(r1)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(r2)
		Expect(os.IsNotExist(err)).To(BeFalse())
		_, err = os.Stat(r3)
		Expect(os.IsNotExist(err)).To(BeFalse())

		By("verify that the links were purged")
		metricsAPI.EXPECT().InstallerCacheTryEviction().Times(1)
		metricsAPI.EXPECT().InstallerCacheReleaseEvicted().Times(1)
		manager.evict()
		_, err = os.Stat(l1)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(l2)
		Expect(os.IsNotExist(err)).To(BeTrue())
		_, err = os.Stat(l3)
		Expect(os.IsNotExist(err)).To(BeTrue())
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", manager.config.MaxCapacity))
	})

	It("extracts from the mirror", func() {
		releaseID := "4.10-orig"
		releaseMirrorID := "4.10-mirror"
		clusterID := strfmt.UUID(uuid.New().String())
		version := "4.10.0"
		mockReleaseCalls(releaseID, version)
		metricsAPI.EXPECT().InstallerCacheReleaseExtracted(releaseID).Times(1)
		metricsAPI.EXPECT().InstallerCacheGetReleaseOK().Times(1)
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
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", manager.config.MaxCapacity))
	})

	It("extracts without a mirror", func() {
		releaseID := "4.10-orig"
		releaseMirrorID := ""
		version := "4.10.0"
		clusterID := strfmt.UUID(uuid.NewString())
		mockReleaseCalls(releaseID, version)
		metricsAPI.EXPECT().InstallerCacheReleaseExtracted(releaseID).Times(1)
		metricsAPI.EXPECT().InstallerCacheGetReleaseOK().Times(1)
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
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", manager.config.MaxCapacity))
	})

	It("should correctly handle multiple requests for the same release at the same time", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(10),
			MaxReleaseSize:                  int64(5),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).ToNot(HaveOccurred())
		params := []launchParams{}
		for i := 0; i < 10; i++ {
			params = append(params, launchParams{releaseID: "4.17.11-x86_64", version: "4.17.11", clusterID: strfmt.UUID(uuid.NewString())})
		}
		expectEventsSent()
		err = runParallelTest(params)
		Expect(err).ToNot(HaveOccurred())
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", manager.config.MaxCapacity))
	})

	It("should consistently handle multiple requests for different releases at the same time", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(25),
			MaxReleaseSize:                  int64(5),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		metricsAPI.EXPECT().InstallerCacheTryEviction().AnyTimes()
		metricsAPI.EXPECT().InstallerCacheReleaseEvicted().AnyTimes()
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).ToNot(HaveOccurred())
		for i := 0; i < 10; i++ {
			params := []launchParams{}
			params = append(params, launchParams{releaseID: "4.17.11-x86_64", version: "4.17.11", clusterID: strfmt.UUID(uuid.NewString())})
			params = append(params, launchParams{releaseID: "4.18.11-x86_64", version: "4.18.11", clusterID: strfmt.UUID(uuid.NewString())})
			params = append(params, launchParams{releaseID: "4.19.11-x86_64", version: "4.19.11", clusterID: strfmt.UUID(uuid.NewString())})
			params = append(params, launchParams{releaseID: "4.20.11-x86_64", version: "4.20.11", clusterID: strfmt.UUID(uuid.NewString())})
			params = append(params, launchParams{releaseID: "4.21.11-x86_64", version: "4.21.11", clusterID: strfmt.UUID(uuid.NewString())})
			expectEventsSent()
			err = runParallelTest(params)
			Expect(err).ToNot(HaveOccurred())
		}
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", manager.config.MaxCapacity))
	})

	It("should maintain cache within threshold", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(500),
			MaxReleaseSize:                  int64(5),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		metricsAPI.EXPECT().InstallerCacheTryEviction().AnyTimes()
		metricsAPI.EXPECT().InstallerCacheReleaseEvicted().AnyTimes()
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).ToNot(HaveOccurred())
		for i := 0; i < 85; i++ { // Deliberately generate a number of requests that will be above the percentage
			params := []launchParams{}
			params = append(params, launchParams{releaseID: fmt.Sprintf("release-%d", i), version: fmt.Sprintf("release-%d.0.1", i), clusterID: strfmt.UUID(uuid.NewString())})
			expectEventsSent()
			err = runParallelTest(params)
			Expect(err).ToNot(HaveOccurred())
		}
		// Ensure that we hold within the correct percentage of the cache size
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", float64(manager.config.MaxCapacity)*installerCacheConfig.InstallerCacheEvictionThreshold))
	})

	It("should stay within the cache limit where there is only sufficient space for one release", func() {
		var err error
		installerCacheConfig := InstallerCacheConfig{
			CacheDir:                        cacheDir,
			MaxCapacity:                     int64(5),
			MaxReleaseSize:                  int64(5),
			ReleaseFetchRetryInterval:       1 * time.Millisecond,
			InstallerCacheEvictionThreshold: 0.8,
		}
		metricsAPI.EXPECT().InstallerCacheTryEviction().AnyTimes()
		metricsAPI.EXPECT().InstallerCacheReleaseEvicted().AnyTimes()
		manager, err = New(installerCacheConfig, eventsHandler, metricsAPI, diskStatsHelper, logrus.New())
		Expect(err).ToNot(HaveOccurred())
		expectEventsSent()
		// Force a scenario where one of the requests will fail because of a pending release cleanup
		params := []launchParams{}
		params = append(params, launchParams{releaseID: "4.17.11-x86_64", version: "4.17.11", clusterID: strfmt.UUID(uuid.NewString())})
		params = append(params, launchParams{releaseID: "4.18.11-x86_64", version: "4.18.11", clusterID: strfmt.UUID(uuid.NewString())})
		err = runParallelTest(params)
		Expect(err).ToNot(HaveOccurred())
		// Now measure disk usage, we should be under the cache size
		Expect(getUsedBytesForDirectory(manager.config.CacheDir)).To(BeNumerically("<=", manager.config.MaxCapacity))

		// After we have collected all results from the last parallel test, the cleanup should have occurred as part of the test.
		// Now assert that a retry would work, there should be enough space for another release
		// use a brand new release ID to prove we are not hitting cache here.
		params = []launchParams{}
		params = append(params, launchParams{releaseID: "4.19.11-x86_64", version: "4.19.11", clusterID: strfmt.UUID(uuid.NewString())})
		err = runParallelTest(params)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should remove expired links while leaving non expired links intact", func() {

		numberOfLinks := 10
		numberOfExpiredLinks := 5

		metricsAPI.EXPECT().InstallerCachePrunedHardLink().Times(numberOfExpiredLinks)

		directory, err := os.MkdirTemp("", "testPruneExpiredHardLinks")
		Expect(err).ToNot(HaveOccurred())

		defer os.RemoveAll(directory)

		for i := 0; i < numberOfLinks; i++ {
			var someFile *os.File
			someFile, err = os.CreateTemp(directory, "somefile")
			Expect(err).ToNot(HaveOccurred())
			linkPath := filepath.Join(directory, fmt.Sprintf("ln_%s", uuid.NewString()))
			err = os.Link(someFile.Name(), linkPath)
			Expect(err).ToNot(HaveOccurred())
			if i > numberOfExpiredLinks-1 {
				err = os.Chtimes(linkPath, time.Now().Add(-10*time.Minute), time.Now().Add(-10*time.Minute))
				Expect(err).ToNot(HaveOccurred())
			}
		}

		links := make([]*fileInfo, 0)
		err = filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
			if strings.HasPrefix(info.Name(), "ln_") {
				links = append(links, &fileInfo{path, info})
			}
			return nil
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(links)).To(Equal(10))

		manager.pruneExpiredHardLinks(links, linkPruningGracePeriod)

		linkCount := 0
		err = filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
			if strings.HasPrefix(info.Name(), "ln_") {
				linkCount++
			}
			return nil
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(linkCount).To(Equal(numberOfLinks - numberOfExpiredLinks))
	})
})

func TestInstallerCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "installercache tests")
}
