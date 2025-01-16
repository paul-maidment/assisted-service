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
	eventsapi "github.com/openshift/assisted-service/internal/events/api"
	"github.com/openshift/assisted-service/internal/metrics"
	"github.com/openshift/assisted-service/internal/oc"
	"github.com/openshift/assisted-service/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// Note: This is the deletion period for the 'automated cleanup' of hardlinks
	// we have yet to establish if this functionality is needed.
	// Leaving this here for now as it's a candidate for deletion.
	DeleteGracePeriod time.Duration = 5 * time.Minute
)

// Installers implements a thread safe LRU cache for ocp install binaries
// on the pod's ephermal file system. The number of binaries stored is
// limited by the storageCapacity parameter.

type Installers struct {
	sync.Mutex
	log logrus.FieldLogger
	// total capcity of the allowed storage (in bytes)
	storageCapacity int64
	// maxReleaseSize is the minimum free space that must be present in the storage directory in order for a write to take place.
	// any attempt to call `installers.Get(` when there is not enough space will result in error,
	// the caller is expected to retry in these cases.
	maxReleaseSize int64
	// parent directory of the binary cache
	cacheDir        string
	eventsHandler   eventsapi.Handler
	diskStatsHelper metrics.DiskStatsHelper
}

//go:generate mockgen -source=installercache.go -package=installercache -destination=mock_installercache.go
type InstallerCache interface {
	Get(ctx context.Context, releaseID, releaseIDMirror, pullSecret string, ocRelease oc.Release, ocpVersion string, clusterID strfmt.UUID) (*Release, error)
}

type fileInfo struct {
	path string
	info os.FileInfo
}

func (fi *fileInfo) Compare(other *fileInfo) bool {
	//oldest file will be first in queue
	//By gathering this in micoseconds, we make sure that the comparison is granular enough
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

func NewRelease(eventsHandler eventsapi.Handler, clusterID strfmt.UUID, releaseID string, startTime time.Time) *Release {
	return &Release{
		eventsHandler: eventsHandler,
		clusterID:     clusterID,
		releaseID:     releaseID,
		startTime:     startTime,
	}
}

// Cleanup is called to signal that the caller has finished using the release and that resources may be released.
func (rl *Release) Cleanup(ctx context.Context) {
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
	if len(rl.Path) == 0 {
		return
	}
	logrus.New().Infof("Cleaning up release %s", rl.Path)
	if err := os.Remove(rl.Path); err != nil {
		logrus.New().WithError(err).Errorf("Failed to delete release link %s", rl.Path)
	}

}

// New constructs an installer cache with a given storage capacity
func New(cacheDir string, storageCapacity int64, maxReleaseSize int64, eventsHandler eventsapi.Handler, log logrus.FieldLogger, diskStatsHelper metrics.DiskStatsHelper) (*Installers, error) {
	if storageCapacity > 0 && maxReleaseSize > storageCapacity {
		return nil, fmt.Errorf("maxReleaseSize (%d bytes) must not be less than storageCapacity (%d bytes)", maxReleaseSize, storageCapacity)
	}
	return &Installers{
		log:             log,
		storageCapacity: storageCapacity,
		cacheDir:        cacheDir,
		eventsHandler:   eventsHandler,
		diskStatsHelper: diskStatsHelper,
		maxReleaseSize:  maxReleaseSize,
	}, nil
}

type ErrorInsufficientCacheCapacity struct {
	Message string
}

func (e *ErrorInsufficientCacheCapacity) Error() string {
	return e.Message
}

// Get returns the path to an openshift-baremetal-install binary extracted from
// the referenced release image. Tries the mirror release image first if it's set. It is safe for concurrent use. A cache of
// binaries is maintained to reduce re-downloading of the same release.
func (i *Installers) Get(ctx context.Context, releaseID, releaseIDMirror, pullSecret string, ocRelease oc.Release, ocpVersion string, clusterID strfmt.UUID) (*Release, error) {
	i.Lock()
	defer i.Unlock()

	release := NewRelease(i.eventsHandler, clusterID, releaseID, time.Now())

	var workdir, binary, path string
	var err error

	workdir, binary, path, err = ocRelease.GetReleaseBinaryPath(releaseID, i.cacheDir, ocpVersion)
	if err != nil {
		return nil, err
	}
	if _, err = os.Stat(path); os.IsNotExist(err) {

		// Ensure that we have sufficient space to store at least one release
		// based on i.maxReleaseSize and the available capacity in the cache
		evictionPerformed := false
		for {
			// If the cache eviction is disabled, skip this step.
			if i.storageCapacity == 0 {
				break
			}
			var usedBytes uint64
			usedBytes, _, err = i.diskStatsHelper.GetDiskUsage(i.cacheDir)
			if err != nil {
				return nil, errors.Wrapf(err, "could not determine disk usage information for cache dir %s", i.cacheDir)
			}
			// Do we have enough space? If so then we can proceed...
			availableBytes := i.storageCapacity - int64(usedBytes)
			if i.storageCapacity > 0 && availableBytes >= i.maxReleaseSize {
				break
			}
			// If we are here, we have insufficient space
			if evictionPerformed {
				// We already performed an eviction, there is nothing more can be done until this is resolved.
				// exit with a *ErrorInsufficientCacheCapacity so that the caller may attempt retry
				message := fmt.Sprintf("insufficient capacity in %s to store release, need %d bytes but only have %d bytes", i.cacheDir, i.maxReleaseSize, availableBytes)
				i.log.Errorf(message)
				return nil, &ErrorInsufficientCacheCapacity{Message: message}
			}
			// Send an eviction and go back around for another try to see if space was freed
			i.evict(i.maxReleaseSize)
			evictionPerformed = true
		}
		// If we are here, we have sufficient space to store the release.

		extractStartTime := time.Now()
		//extract the binary
		_, err = ocRelease.Extract(i.log, releaseID, releaseIDMirror, i.cacheDir, pullSecret, ocpVersion)
		if err != nil {
			return nil, err
		}
		release.extractDuration = time.Since(extractStartTime).Seconds()
	} else {
		release.cached = true
		//update the file mtime to signal it was recently used
		err = os.Chtimes(path, time.Now(), time.Now())
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Failed to update release binary %s", path))
		}
	}
	// return a new hard link to the binary file
	// the caller should delete the hard link when
	// it finishes working with the file
	// Filename should be generated using micosecond timing to prevent collision.
	link := filepath.Join(workdir, fmt.Sprintf("ln_%d_%s", time.Now().UnixMicro(), binary))
	err = os.Link(path, link)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Failed to create hard link to binary %s", path))
	}
	release.Path = link
	return release, nil
}

// Walk through the cacheDir and list the files recursively.
// If the total volume of the files reaches the capacity, delete
// the oldest ones.
//
// Locking must be done outside evict() to avoid contentions.
func (i *Installers) evict(capacityNeeded int64) {
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

	err := filepath.Walk(i.cacheDir, visit)
	if err != nil {
		if !os.IsNotExist(err) { //ignore first invocation where the cacheDir does not exist
			i.log.WithError(err).Errorf("release binary eviction failed to inspect directory %s", i.cacheDir)
		}
		return
	}

	// ####
	// PM 27 January 2025 - Honestly, not sure if this code is even needed, all links should be purged as they are part of calls with defer
	// maybe we can add a counter for this block to metrics and if we don't get any calls to this code in a while, we can delete this block?
	// #####
	//prune the hard links just in case the deletion of resources
	//in ignition.go did not succeeded as expected
	for idx := 0; idx < len(links); idx++ {
		finfo := links[idx]
		//Allow a grace period of 5 minutes from the link creation time
		//to ensure the link is not being used.
		if finfo.info.ModTime().Add(DeleteGracePeriod).UnixMicro() <= time.Now().UnixMicro() {
			os.Remove(finfo.path)
		}
	}

	// delete the oldest file if necessary
	// prefer files without hardlinks first because
	// 1: files without hardlinks are likely to be least recently used anyway
	// 2: files without hardlinks will immediately free up storage
	queues := []*PriorityQueue[*fileInfo]{}
	withoutHardLinks, withHardLinks := i.splitQueueOnHardLinksCount(files)
	// Only process the queues if they have items
	if withoutHardLinks.Len() > 0 {
		queues = append(queues, withoutHardLinks)
	}
	if withHardLinks.Len() > 0 {
		queues = append(queues, withHardLinks)
	}
	for _, q := range queues {
		for totalSize+capacityNeeded > i.storageCapacity {
			if q.Len() == 0 { // If we have cleaned out this queue then break the loop
				break
			}
			finfo, _ := q.Pop()
			totalSize -= finfo.info.Size()
			//remove the file
			if err := i.evictFile(finfo.path); err != nil {
				i.log.WithError(err).Errorf("failed to evict file %s", finfo.path)
			}
		}
	}
}

// splitQueueOnHardLinksCount Splits the provided *PriorityQueue[*fileInfo] into two queues
// withoutHardLinks will present a queue of files that do not have associated hardlinks
// withHardLinks will present a queue of files that do have associated hardlinks
// This is to allow us to prioritize deletion, favouring files without hardlinks first as these will have an immediate impact on storage.
func (i *Installers) splitQueueOnHardLinksCount(in *PriorityQueue[*fileInfo]) (*PriorityQueue[*fileInfo], *PriorityQueue[*fileInfo]) {
	withoutHardLinks := &PriorityQueue[*fileInfo]{}
	withHardLinks := &PriorityQueue[*fileInfo]{}
	for {
		if in.Len() == 0 {
			break
		}
		fileInfo, _ := in.Pop()
		stat, ok := fileInfo.info.Sys().(*syscall.Stat_t)
		if !ok {
			i.log.Errorf("unable to determine stat information for file with path %s -- skipping", fileInfo.path)
			continue
		}
		if stat.Nlink == 0 {
			withHardLinks.Add(fileInfo)
			continue
		}
		withHardLinks.Add(fileInfo)
	}
	return withoutHardLinks, withHardLinks
}

func (i *Installers) evictFile(filePath string) error {
	i.log.Infof("evicting binary file %s due to storage pressure", filePath)
	err := os.Remove(filePath)
	if err != nil {
		return err
	}
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
