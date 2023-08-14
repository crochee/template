package utils

import "os"

type FileOrContent string

func (f FileOrContent) String() string {
	return string(f)
}

// IsPath returns true if the FileOrContent is a file path, otherwise returns false.
func (f FileOrContent) IsPath() bool {
	_, err := os.Stat(f.String())
	return err == nil
}

// Read output content
func (f FileOrContent) Read() ([]byte, error) {
	var content []byte
	if f.IsPath() {
		var err error
		if content, err = os.ReadFile(f.String()); err != nil {
			return nil, err
		}
	} else {
		content = []byte(f)
	}
	return content, nil
}
