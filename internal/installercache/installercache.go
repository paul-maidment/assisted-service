package installercache

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	eventsapi "github.com/openshift/assisted-service/internal/events/api"
	"github.com/openshift/assisted-service/internal/metrics"
	"github.com/openshift/assisted-service/internal/oc"
	"github.com/openshift/assisted-service/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	linkPruningGracePeriod time.Duration = 5 * time.Minute
)

// Installers implements a thread safe LRU cache for ocp install binaries
// on the pod's ephermal file system. The number of binaries stored is
// limited by the storageCapacity parameter.
type Installers struct {
	sync.Mutex
	log logrus.FieldLogger
	// parent directory of the binary cache
	eventsHandler   eventsapi.Handler
	diskStatsHelper metrics.DiskStatsHelper
	config          InstallerCacheConfig
	metricsAPI      metrics.API
}

type InstallerCacheConfig struct {
	CacheDir                        string
	MaxCapacity                     int64
	MaxReleaseSize                  int64
	ReleaseFetchRetryInterval       time.Duration
	InstallerCacheEvictionThreshold float64
}

type errorInsufficientCacheCapacity struct {
	Message string
}

func (e *errorInsufficientCacheCapacity) Error() string {
	return e.Message
}

type fileInfo struct {
	path string
	info os.FileInfo
}

func (fi *fileInfo) Compare(other *fileInfo) bool {
	//oldest file will be first in queue
	// Using micoseconds to make sure that the comparison is granular enough
	return fi.info.ModTime().UnixMicro() < other.info.ModTime().UnixMicro()
}

const (
	metricEventInstallerCacheRelease = "installercache.release.metrics"
)

type Release struct {
	Path          string
	eventsHandler eventsapi.Handler
	// startTime is the time at which the request was made to fetch the release.
	startTime time.Time
	// clusterID is the UUID of the cluster for which the release is being fetched.
	clusterID strfmt.UUID
	// releaseId is the release that is being fetched, for example "4.10.67-x86_64".
	releaseID string
	// cached is `true` if the release was found in the cache, otherwise `false`.
	cached bool
	// extractDuration is the amount of time taken to perform extraction, zero if no extraction took place.
	extractDuration float64
}

// Cleanup is called to signal that the caller has finished using the release and that resources may be released.
func (rl *Release) Cleanup(ctx context.Context) {
	logrus.New().Infof("Cleaning up release %s", rl.Path)
	rl.eventsHandler.V2AddMetricsEvent(
		ctx, &rl.clusterID, nil, nil, "", models.EventSeverityInfo,
		metricEventInstallerCacheRelease,
		time.Now(),
		"release_id", rl.releaseID,
		"start_time", rl.startTime.Format(time.RFC3339),
		"end_time", time.Now().Format(time.RFC3339),
		"cached", rl.cached,
		"extract_duration", rl.extractDuration,
	)
	if err := os.Remove(rl.Path); err != nil {
		logrus.New().WithError(err).Errorf("Failed to delete release link %s", rl.Path)
		return
	}
}

// New constructs an installer cache with a given storage capacity
func New(config InstallerCacheConfig, eventsHandler eventsapi.Handler, metricsAPI metrics.API, diskStatsHelper metrics.DiskStatsHelper, log logrus.FieldLogger) (*Installers, error) {
	if config.InstallerCacheEvictionThreshold == 0 {
		return nil, errors.New("config.InstallerCacheEvictionThreshold must not be zero")
	}
	if config.MaxCapacity > 0 && config.MaxReleaseSize == 0 {
		return nil, fmt.Errorf("config.MaxReleaseSize (%d bytes) must not be zero", config.MaxReleaseSize)
	}
	if config.MaxCapacity > 0 && config.MaxReleaseSize > config.MaxCapacity {
		return nil, fmt.Errorf("config.MaxReleaseSize (%d bytes) must not be greater than config.MaxCapacity (%d bytes)", config.MaxReleaseSize, config.MaxCapacity)
	}
	return &Installers{
		log:             log,
		eventsHandler:   eventsHandler,
		diskStatsHelper: diskStatsHelper,
		config:          config,
		metricsAPI:      metricsAPI,
	}, nil
}

// Get returns the path to an openshift-baremetal-install binary extracted from
// the referenced release image. Tries the mirror release image first if it's set. It is safe for concurrent use. A cache of
// binaries is maintained to reduce re-downloading of the same release.
func (i *Installers) Get(ctx context.Context, releaseID, releaseIDMirror, pullSecret string, ocRelease oc.Release, ocpVersion string, clusterID strfmt.UUID) (*Release, error) {
	for {
		select {
		case <-ctx.Done():
			i.metricsAPI.InstallerCacheGetReleaseTimeout()
			return nil, errors.Errorf("context cancelled or timed out while fetching release %s", releaseID)
		default:
			release, err := i.get(releaseID, releaseIDMirror, pullSecret, ocRelease, ocpVersion, clusterID)
			if err == nil {
				i.metricsAPI.InstallerCacheGetReleaseOK()
				return release, nil
			}
			_, isCapacityError := err.(*errorInsufficientCacheCapacity)
			if !isCapacityError {
				i.metricsAPI.InstallerCacheGetReleaseError()
				return nil, errors.Wrapf(err, "failed to get installer path for release %s", releaseID)
			}
			i.log.WithError(err).Errorf("insufficient installer cache capacity for release %s", releaseID)
			time.Sleep(i.config.ReleaseFetchRetryInterval)
		}
	}
}

func (i *Installers) get(releaseID, releaseIDMirror, pullSecret string, ocRelease oc.Release, ocpVersion string, clusterID strfmt.UUID) (*Release, error) {
	i.Lock()
	defer i.Unlock()

	release := &Release{
		eventsHandler: i.eventsHandler,
		clusterID:     clusterID,
		releaseID:     releaseID,
		startTime:     time.Now(),
	}

	workdir, binary, path, err := ocRelease.GetReleaseBinaryPath(releaseID, i.config.CacheDir, ocpVersion)
	if err != nil {
		return nil, err
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {
		i.log.Infof("release %s - not found in cache - attempting extraction", releaseID)
		// Cache eviction and space checks
		evictionAlreadyPerformed := false
		for {
			if i.config.MaxCapacity == 0 {
				i.log.Info("cache eviction is disabled -- moving directly to extraction")
				break
			}
			i.log.Infof("checking for space to extract release %s", releaseID)
			// Determine actual space usage, accounting for hardlinks.
			var usedBytes uint64
			usedBytes, _, err = i.diskStatsHelper.GetDiskUsage(i.config.CacheDir)
			if err != nil {
				if os.IsNotExist(err) {
					// Installer cache directory will not exist prior to first extraction
					i.log.WithError(err).Warnf("skipping capacity check as first extraction will trigger creation of directory")
					break
				}
				return nil, errors.Wrapf(err, "could not determine disk usage information for cache dir %s", i.config.CacheDir)
			}
			shouldEvict, isBlocked := i.checkEvictionStatus(int64(usedBytes))
			// If we have already been around once, we don't want to 'double' evict
			if shouldEvict && !evictionAlreadyPerformed {
				i.metricsAPI.InstallerCacheTryEviction()
				i.evict()
			}
			if !isBlocked {
				break
			}
			if evictionAlreadyPerformed {
				return nil, &errorInsufficientCacheCapacity{
					Message: fmt.Sprintf("insufficient capacity in %s to store release", i.config.CacheDir),
				}
			}
			evictionAlreadyPerformed = true
		}
		i.log.Infof("space available for release %s -- starting extraction", releaseID)
		extractStartTime := time.Now()
		_, err = ocRelease.Extract(i.log, releaseID, releaseIDMirror, i.config.CacheDir, pullSecret, ocpVersion)
		if err != nil {
			return nil, err
		}
		i.metricsAPI.InstallerCacheReleaseExtracted(releaseID)
		release.extractDuration = time.Since(extractStartTime).Seconds()
		i.log.Infof("finished extraction of %s", releaseID)
	} else {
		i.log.Infof("fetching release %s from cache", releaseID)
		i.metricsAPI.InstallerCacheReleaseCached(releaseID)
		release.cached = true
	}

	// update the file mtime to signal it was recently used
	// do this, no matter where we sourced the binary (cache or otherwise) as it would appear that the binary
	// can have a modtime - upon extraction - that is significantly in the past
	// this can lead to premature link pruning in some scenarios.
	err = os.Chtimes(path, time.Now(), time.Now())
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to update release binary %s", path))
	}

	// return a new hard link to the binary file
	// the caller should delete the hard link when
	// it finishes working with the file
	i.log.Infof("attempting hardlink creation for release %s", releaseID)
	for {
		link := filepath.Join(workdir, fmt.Sprintf("ln_%s_%s", uuid.NewString(), binary))
		err = os.Link(path, link)
		if err == nil {
			i.log.Infof("created hardlink %s to release %s at path %s", link, releaseID, path)
			release.Path = link
			return release, nil
		}
		if !os.IsExist(err) {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to create hard link to binary %s", path))
		}
	}
}

func (i *Installers) shouldEvict(totalUsed int64) bool {
	shouldEvict, _ := i.checkEvictionStatus(totalUsed)
	return shouldEvict
}

func (i *Installers) checkEvictionStatus(totalUsed int64) (shouldEvict bool, isBlocked bool) {
	if i.config.MaxCapacity == 0 {
		// The cache eviction is completely disabled.
		return false, false
	}
	if i.config.MaxCapacity-totalUsed < i.config.MaxReleaseSize {
		// We are badly blocked, not enough room for even one release more.
		return true, true
	}
	if (float64(totalUsed) / float64(i.config.MaxCapacity)) >= i.config.InstallerCacheEvictionThreshold {
		// We need to evict some items in order to keep cache efficient and within capacity.
		return true, false
	}
	// We have enough space.
	return false, false
}

// Walk through the cacheDir and list the files recursively.
// If the total volume of the files reaches the capacity, delete
// the oldest ones.
//
// Locking must be done outside evict() to avoid contentions.
func (i *Installers) evict() {
	// store the file paths
	files := NewPriorityQueue(&fileInfo{})
	links := make([]*fileInfo, 0)
	var totalSize int64

	// visit process the file/dir pointed by path and store relevant
	// paths in a priority queue
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}
		//find hard links
		if strings.HasPrefix(info.Name(), "ln_") {
			links = append(links, &fileInfo{path, info})
			return nil
		}

		//save the other files based on their mod time
		files.Add(&fileInfo{path, info})
		totalSize += info.Size()
		return nil
	}

	err := filepath.Walk(i.config.CacheDir, visit)
	if err != nil {
		if !os.IsNotExist(err) { //ignore first invocation where the cacheDir does not exist
			i.log.WithError(err).Errorf("release binary eviction failed to inspect directory %s", i.config.CacheDir)
		}
		return
	}

	// TODO: We might want to consider if we need to do this longer term, in theory every hardlink should be automatically freed.
	// For now, moved to a function so that this can be tested in unit tests.
	i.pruneExpiredHardLinks(links, linkPruningGracePeriod)

	// delete the oldest file if necessary
	// prefer files without hardlinks first because
	// 1: files without hardlinks are likely to be least recently used anyway
	// 2: files without hardlinks will immediately free up storage
	queues := i.splitQueueOnHardLinks(files)
	for _, q := range queues {
		for i.shouldEvict(totalSize) {
			if q.Len() == 0 { // If we have cleaned out this queue then break the loop
				break
			}
			finfo, _ := q.Pop()
			totalSize -= finfo.info.Size()
			if err := i.evictFile(finfo.path); err != nil {
				i.log.WithError(err).Errorf("failed to evict file %s", finfo.path)
			}
		}
	}
}

func (i *Installers) evictFile(filePath string) error {
	// TODO: Count - Release eviction (label with release ID)
	i.log.Infof("evicting binary file %s due to storage pressure", filePath)
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
	i.metricsAPI.InstallerCacheReleaseEvicted()
	// if the parent directory was left empty,
	// remove it to avoid dangling directories
	parentDir := path.Dir(filePath)
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return os.Remove(parentDir)
	}
	return nil
}

// pruneExpiredHardLinks removes any hardlinks that have been around for too long
// the grace period is used to determine which links should be removed.
func (i *Installers) pruneExpiredHardLinks(links []*fileInfo, gracePeriod time.Duration) {
	for idx := 0; idx < len(links); idx++ {
		finfo := links[idx]
		graceTime := time.Now().Add(-1 * gracePeriod)
		grace := graceTime.Unix()
		if finfo.info.ModTime().Unix() < grace {
			i.metricsAPI.InstallerCachePrunedHardLink()
			i.log.Infof("Found expired hardlink -- pruning %s", finfo.info.Name())
			i.log.Infof("Mod time %s", finfo.info.ModTime().Format("2006-01-02 15:04:05"))
			i.log.Infof("Grace time %s", graceTime.Format("2006-01-02 15:04:05"))
			os.Remove(finfo.path)
		}
	}
}

// splitQueueOnHardLinks Splits the provided *PriorityQueue[*fileInfo] into two queues
// withoutHardLinks will present a queue of files that do not have associated hardlinks
// withHardLinks will present a queue of files that do have associated hardlinks
// This is to allow us to prioritize deletion, favouring files without hardlinks first as these will have an immediate impact on storage.
func (i *Installers) splitQueueOnHardLinks(in *PriorityQueue[*fileInfo]) []*PriorityQueue[*fileInfo] {
	withoutHardLinks := &PriorityQueue[*fileInfo]{}
	withHardLinks := &PriorityQueue[*fileInfo]{}
	for {
		if in.Len() == 0 {
			break
		}
		fi, _ := in.Pop()
		stat, ok := fi.info.Sys().(*syscall.Stat_t)
		if !ok {
			// If we do encounter an error while performing stat - let's fall back to the original queue
			// it's not optimal, but we don't break anything.
			i.log.Errorf("encountered error while trying to split queues - using original queue")
			return []*PriorityQueue[*fileInfo]{in}
		}
		if stat.Nlink == 0 {
			withoutHardLinks.Add(fi)
			continue
		}
		withHardLinks.Add(fi)
	}
	return []*PriorityQueue[*fileInfo]{withoutHardLinks, withHardLinks}
}
