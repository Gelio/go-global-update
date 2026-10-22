package gobinaries

import "os"

type DirectoryLister interface {
	ListDirectoryEntries(path string) ([]string, error)
}

type FilesystemDirectoryLister struct{}

func (*FilesystemDirectoryLister) ListDirectoryEntries(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}
