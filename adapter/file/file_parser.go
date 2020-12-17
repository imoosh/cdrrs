package file

import (
	"bufio"
	"centnet-cdrrs/library/log"
	"fmt"
	"go.uber.org/atomic"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Config struct {
	RootPath      string // 规则顶层路径
	StartDatePath string // 开始日期路径 202012/20201201
	MaxParseProcs int
}

type FileInfoX struct {
	name     string
	filepath string
	dtime    time.Time
}

func (fi FileInfoX) String() string {
	return fmt.Sprintf("dtime:%s name:%s", fi.dtime, fi.name)
}

type RawFileParser struct {
	conf *Config

	CurrentDaysPath   string
	Files             chan FileInfoX
	newestWrittenTime time.Time // 遍历的数据文件中最新写完时间
	parseLineFunc     func(interface{})
	numWorkers        *atomic.Int32
	numFiles          *atomic.Int32
}

func NewRawFileParser(c *Config) *RawFileParser {
	return &RawFileParser{
		conf:              c,
		CurrentDaysPath:   c.StartDatePath,
		Files:             make(chan FileInfoX, 1024),
		newestWrittenTime: time.Time{},
		numWorkers:        atomic.NewInt32(0),
		numFiles:          atomic.NewInt32(0),
	}
}

func fileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

func (rfp *RawFileParser) sortFileList(fileInfoX []FileInfoX) {
	if len(fileInfoX) == 0 {
		return
	}

	timeBy := func(a, b interface{}) bool {
		return a.(FileInfoX).dtime.Before(b.(FileInfoX).dtime)
	}

	results := Bucket{
		Slice: fileInfoX,
		By:    timeBy,
	}

	sort.Sort(results)
}

func (rfp *RawFileParser) FlushFileList() {

	// 读目录信息
	currentPath := filepath.Join(rfp.conf.RootPath, rfp.CurrentDaysPath)
	dirInfo, err := ioutil.ReadDir(currentPath)
	if err != nil {
		log.Error(err)
		return
	}

	var todayComplete = true
	var fileList []FileInfoX
	for _, info := range dirInfo {
		// 只处理.txt文件
		if !(strings.HasSuffix(info.Name(), ".txt")) {
			continue
		}
		// 对应.ok文件未生成，标识数据未写完，不处理.txt文件
		if !fileExist(currentPath + "/" + info.Name() + ".ok") {
			todayComplete = false
			continue
		}

		// 之前已经遍历过的文件不处理，只比较时分秒
		dtime, _ := time.Parse("15:04:05.000000", info.ModTime().Format("15:04:05.000000"))
		ltime, _ := time.Parse("15:04:05.000000", rfp.newestWrittenTime.Format("15:04:05.000000"))
		if !dtime.After(ltime) {
			continue
		}
		fileList = append(fileList, FileInfoX{
			name:     info.Name(),
			dtime:    info.ModTime(),
			filepath: filepath.Join(currentPath, info.Name()),
		})
	}

	if len(fileList) != 0 {
		// 按修改时间排序，并记录最新那个文件写完时间
		rfp.sortFileList(fileList)
		rfp.newestWrittenTime = fileList[len(fileList)-1].dtime

		// 新写完的文件放入队列
		for _, file := range fileList {
			rfp.Files <- file
		}
	}

	// 202012/20201201，切割出20201201，并解析成时间
	subPath := strings.Split(rfp.CurrentDaysPath, "/")
	subPathTime, err := time.Parse("20060102", subPath[1])
	if err != nil {
		log.Error(err)
		return
	}

	// 当天文件都已写完，并且现在时间日期不是最新写完文件时间日期，立即切换到下一天日期目录
	if todayComplete && time.Now().Day() != subPathTime.Day() {
		rfp.newestWrittenTime = time.Time{}
		rfp.CurrentDaysPath = getNextDaysReadPath(subPathTime)
	}
}

func getNextDaysReadPath(t time.Time) string {
	return t.Add(time.Hour * 24).Format("200601/20060102")
}

func (rfp *RawFileParser) Run(fun func(interface{})) {

	rfp.parseLineFunc = fun
	go func() {
		// 定时刷新文件信息，新写完的文件放入队列
		ticker := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-ticker.C:
				rfp.FlushFileList()
			}
		}
	}()

	// 开启多通道读文件
	for i := 0; i < rfp.conf.MaxParseProcs; i++ {
		go rfp.parseLoop()
	}
}

func (rfp *RawFileParser) parseLoop() {
	var fix FileInfoX
	bufReader := bufio.NewReaderSize(nil, 10<<20)
	for {
		// 阻塞获取一个读文件任务
		select {
		case fix = <-rfp.Files:
			rfp.parseFile(fix, bufReader)
		}
	}
}

func (rfp *RawFileParser) parseFile(x FileInfoX, bufReader *bufio.Reader) {

	rfp.numWorkers.Add(1)

	// 处理文件流程
	fr, err := os.OpenFile(x.filepath, os.O_RDONLY, 0)
	defer fr.Close()
	if err != nil {
		log.Errorf("os.OpenFile '%s' error: %s", x.filepath, err)
		return
	}

	t := time.Now()
	log.Debugf("[%2d/%d] - '%s' opened", rfp.numWorkers.Load(), rfp.conf.MaxParseProcs,
		x.filepath[len(rfp.conf.RootPath)+1:])

	bufReader.Reset(fr)
	for {
		line, _, err := bufReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("bufio.Readline error:%s", err)
		}

		// 调用model.DoLine
		rfp.parseLineFunc(line)
	}
	rfp.numWorkers.Sub(1)
	log.Debugf("[%d/%d] - '%s' parsed done. %.1fs used, file count: %d", rfp.numWorkers.Load(),
		rfp.conf.MaxParseProcs, x.filepath[len(rfp.conf.RootPath)+1:], time.Since(t).Seconds(), rfp.numFiles.Add(1))
}
