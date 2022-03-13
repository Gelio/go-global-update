package gobinariestest

type TestSuccessDirectoryLister struct{ Entries []string }

func (l *TestSuccessDirectoryLister) ListDirectoryEntries(path string) ([]string, error) {
	return l.Entries, nil
}
