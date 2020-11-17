package file

import "os"


type FileReader struct {
    Filepath string
    File     *os.File
}

func NewFileReader(filepath string) (*FileReader, error) {
    _, err := os.Stat(filepath)
    if err != nil {
        return nil, err
    }

    file, err := os.OpenFile(filepath, os.O_RDONLY, 0)
    if err != nil {
        return nil, err
    }

    return &FileReader{
        Filepath: filepath,
        File:     file,
    }, nil
}

func (fr *FileReader) Read(b []byte) (int, error) {
    return fr.File.Read(b)
}

func (fr *FileReader) Close() error {
    return fr.File.Close()
}
