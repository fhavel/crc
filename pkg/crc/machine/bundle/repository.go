package bundle

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/code-ready/crc/pkg/crc/constants"
	"github.com/code-ready/crc/pkg/extract"
	crcos "github.com/code-ready/crc/pkg/os"
	"github.com/pkg/errors"
)

const (
	bundleExtension  = ".crcbundle"
	metadataFilename = "crc-bundle-info.json"
)

type Repository struct {
	CacheDir string
	OcBinDir string
}

func (repo *Repository) Get(bundleName string) (*CrcBundleInfo, error) {
	path := filepath.Join(repo.CacheDir, strings.TrimSuffix(bundleName, bundleExtension))
	if _, err := os.Stat(path); err != nil {
		return nil, errors.Wrapf(err, "could not find cached bundle info in %s", path)
	}
	jsonFilepath := filepath.Join(path, metadataFilename)
	content, err := ioutil.ReadFile(jsonFilepath)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading %s file", jsonFilepath)
	}
	var bundleInfo CrcBundleInfo
	if err := json.Unmarshal(content, &bundleInfo); err != nil {
		return nil, errors.Wrap(err, "error Unmarshal the data")
	}
	bundleInfo.cachedPath = path
	return &bundleInfo, nil
}

func (repo *Repository) Use(bundleName string) (*CrcBundleInfo, error) {
	bundleInfo, err := repo.Get(bundleName)
	if err != nil {
		return nil, err
	}
	if err := bundleInfo.createSymlinkOrCopyOpenShiftClient(repo.OcBinDir); err != nil {
		return nil, err
	}
	return bundleInfo, nil
}

func (bundle *CrcBundleInfo) createSymlinkOrCopyOpenShiftClient(ocBinDir string) error {
	ocInBundle := filepath.Join(bundle.cachedPath, constants.OcExecutableName)
	ocInBinDir := filepath.Join(ocBinDir, constants.OcExecutableName)
	if err := os.MkdirAll(ocBinDir, 0750); err != nil {
		return err
	}
	_ = os.Remove(ocInBinDir)
	if runtime.GOOS == "windows" {
		return crcos.CopyFileContents(ocInBundle, ocInBinDir, 0750)
	}
	return os.Symlink(ocInBundle, ocInBinDir)
}

func (repo *Repository) Extract(path string) error {
	_, err := extract.Uncompress(path, repo.CacheDir, true)
	return err
}

var defaultRepo = &Repository{
	CacheDir: constants.MachineCacheDir,
	OcBinDir: constants.CrcOcBinDir,
}

func GetCachedBundleInfo(bundleName string) (*CrcBundleInfo, error) {
	return defaultRepo.Use(bundleName)
}

func Extract(path string) (*CrcBundleInfo, error) {
	if err := defaultRepo.Extract(path); err != nil {
		return nil, err
	}
	return defaultRepo.Use(filepath.Base(path))
}