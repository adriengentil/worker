package worker

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/travis-ci/worker/backend"
	"github.com/travis-ci/worker/context"
	gocontext "golang.org/x/net/context"
)

// FileJobQueue is a JobQueue that uses directories for input, state, and output
type FileJobQueue struct {
	queue string

	baseDir     string
	createdDir  string
	receivedDir string
	startedDir  string
	finishedDir string
	logDir      string
}

// NewFileJobQueue creates a *FileJobQueue from a base directory and queue name
func NewFileJobQueue(baseDir, queue string) (*FileJobQueue, error) {
	_, err := os.Stat(baseDir)
	if err != nil {
		return nil, err
	}

	fd, err := os.Create(filepath.Join(baseDir, ".write-test"))
	if err != nil {
		return nil, err
	}

	defer fd.Close()

	createdDir := filepath.Join(baseDir, queue, "created")
	receivedDir := filepath.Join(baseDir, queue, "received")
	startedDir := filepath.Join(baseDir, queue, "started")
	finishedDir := filepath.Join(baseDir, queue, "finished")
	logDir := filepath.Join(baseDir, queue, "log")

	for _, dirname := range []string{createdDir, receivedDir, startedDir, finishedDir} {
		err := os.MkdirAll(dirname, os.FileMode(0755))
		if err != nil {
			return nil, err
		}
	}

	return &FileJobQueue{
		queue: queue,

		baseDir:     baseDir,
		createdDir:  createdDir,
		receivedDir: receivedDir,
		startedDir:  startedDir,
		finishedDir: finishedDir,
		logDir:      logDir,
	}, nil
}

// Jobs returns a channel of jobs from the created directory
func (f *FileJobQueue) Jobs(ctx gocontext.Context) (<-chan Job, error) {
	buildJobChan := make(chan Job)
	go f.pollInDirForJobs(ctx, buildJobChan)
	return buildJobChan, nil
}

func (f *FileJobQueue) pollInDirForJobs(ctx gocontext.Context, buildJobChan chan Job) {
	for {
		f.pollInDirTick(ctx, buildJobChan)
		time.Sleep(1 * time.Second)
	}
}

func (f *FileJobQueue) pollInDirTick(ctx gocontext.Context, buildJobChan chan Job) {
	entries, err := ioutil.ReadDir(f.createdDir)
	if err != nil {
		context.LoggerFromContext(ctx).WithField("err", err).Error("input directory read error")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		buildJob := &fileJob{
			createdFile:     filepath.Join(f.createdDir, entry.Name()),
			payload:         &JobPayload{},
			startAttributes: &backend.StartAttributes{},
		}
		startAttrs := &jobPayloadStartAttrs{Config: &backend.StartAttributes{}}

		fb, err := ioutil.ReadFile(buildJob.createdFile)
		if err != nil {
			context.LoggerFromContext(ctx).WithField("err", err).Error("input file read error")
			continue
		}

		err = json.Unmarshal(fb, buildJob.payload)
		if err != nil {
			context.LoggerFromContext(ctx).WithField("err", err).Error("payload JSON parse error")
			continue
		}

		err = json.Unmarshal(fb, &startAttrs)
		if err != nil {
			context.LoggerFromContext(ctx).WithField("err", err).Error("start attributes JSON parse error")
			continue
		}

		buildJob.rawPayload, err = simplejson.NewJson(fb)
		if err != nil {
			context.LoggerFromContext(ctx).WithField("err", err).Error("raw payload JSON parse error")
			continue
		}

		buildJob.startAttributes = startAttrs.Config
		buildJob.receivedFile = filepath.Join(f.receivedDir, entry.Name())
		buildJob.startedFile = filepath.Join(f.startedDir, entry.Name())
		buildJob.finishedFile = filepath.Join(f.finishedDir, entry.Name())
		buildJob.logFile = filepath.Join(f.logDir, strings.Replace(entry.Name(), ".json", ".log", -1))
		buildJob.bytes = fb

		buildJobChan <- buildJob
	}
}

// Cleanup is a no-op
func (f *FileJobQueue) Cleanup() error {
	return nil
}
