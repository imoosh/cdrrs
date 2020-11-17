package file

import "os"

const (
    ReadOrderByFileName = iota
    ReadOrderByCreatedTime
)

var (
    ReadPattern = 0
    FilePath    = "/Users/wayne/go/src/VoipSniffer/adapter/file/files"
)

func ReadInOrder() error {

    _, err := os.Stat(FilePath)
    if err != nil {
        return err
    }

    var fileList []string
    for {

    }
}
