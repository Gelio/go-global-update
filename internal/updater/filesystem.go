package updater

import "os"

type FilesystemUtils interface {
	Chdir(dir string) error
}

type Filesystem struct{}

func (fs *Filesystem) Chdir(dir string) error {
	return os.Chdir(dir)
}
