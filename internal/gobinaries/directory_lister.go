package gobinaries

import "io/ioutil"

type DirectoryLister interface {
	ListDirectoryEntries(path string) ([]string, error)
}

type FilesystemDirectoryLister struct{}

func (_ *FilesystemDirectoryLister) ListDirectoryEntries(path string) ([]string, error) {
	fileinfos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, fileinfo := range fileinfos {
		names = append(names, fileinfo.Name())
	}

	return names, nil
}
