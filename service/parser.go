package service

import (
	"bufio"
	"centnet-cdrrs/common/log"
	"centnet-cdrrs/conf"
	"fmt"
	"go.uber.org/atomic"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	defaultParseFunc = func(interface{}) {}
)

type FileInfoX struct {
	name     string
	filepath string
	dtime    time.Time
}

func (fi FileInfoX) String() string {
	return fmt.Sprintf("dtime:%s name:%s", fi.dtime, fi.name)
}

type Parser struct {
	conf *conf.ParserConfig

	CurrentDaysPath   string
	Files             chan FileInfoX
	newestWrittenTime time.Time // 遍历的数据文件中最新写完时间
	parseFunc         func(interface{})
	numWorkers        *atomic.Int32
	numFiles          *atomic.Int32

	mu sync.Mutex
}

func NewParser(c *conf.ParserConfig) *Parser {
	return &Parser{
		conf:              c,
		CurrentDaysPath:   c.StartDatePath,
		Files:             make(chan FileInfoX, 1024),
		newestWrittenTime: time.Time{},
		parseFunc:         defaultParseFunc,
		numWorkers:        atomic.NewInt32(0),
		numFiles:          atomic.NewInt32(0),
	}
}

func fileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

func (p *Parser) sortFileList(fileInfoX []FileInfoX) {
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

func (p *Parser) FlushFiles() {

	// 读目录信息
	currentPath := filepath.Join(p.conf.RootPath, p.CurrentDaysPath)
	dirInfo, err := ioutil.ReadDir(currentPath)
	if err != nil {
		log.Error(err)
		return
	}

	var todayComplete = true
	var fileList []FileInfoX
	for _, info := range dirInfo {
		// 只处理.txt文件
		if !strings.HasSuffix(info.Name(), ".txt") {
			continue
		}
		// 对应.ok文件未生成，标识数据未写完，不处理.txt文件
		if !fileExist(currentPath + "/" + info.Name() + ".ok") {
			todayComplete = false
			continue
		}

		// 之前已经遍历过的文件不处理，只比较时分秒
		dtime, _ := time.Parse("15:04:05.000000", info.ModTime().Format("15:04:05.000000"))
		ltime, _ := time.Parse("15:04:05.000000", p.newestWrittenTime.Format("15:04:05.000000"))
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
		p.sortFileList(fileList)
		p.newestWrittenTime = fileList[len(fileList)-1].dtime

		// 新写完的文件放入队列
		for _, file := range fileList {
			p.Files <- file
		}
	}

	// 202012/20201201，切割出20201201，并解析成时间
	subPath := strings.Split(p.CurrentDaysPath, "/")
	subPathTime, err := time.Parse("20060102", subPath[1])
	if err != nil {
		log.Error(err)
		return
	}

	// 当前系统日期不是当前目录日期
	if time.Now().Day() != subPathTime.Day() {
		// 当天文件还没写完
		if !todayComplete {
			// 不等了。把未生成ok文件对应的txt文件直接入队
			fileList = fileList[:0]
			for _, info := range dirInfo {
				if strings.HasSuffix(info.Name(), ".txt") && !fileExist(currentPath+"/"+info.Name()+".ok") {
					fileList = append(fileList, FileInfoX{
						name:     info.Name(),
						dtime:    info.ModTime(),
						filepath: filepath.Join(currentPath, info.Name()),
					})
				}
			}

			if len(fileList) != 0 {
				// 按修改时间排序，并记录最新那个文件写完时间
				p.sortFileList(fileList)
				//p.newestWrittenTime = fileList[len(fileList)-1].dtime

				for _, file := range fileList {
					p.Files <- file
				}
			}
		}

		// 立即滚动到下一天日期目录
		p.newestWrittenTime = time.Time{}
		p.CurrentDaysPath = getNextDaysReadPath(subPathTime)
	}
}

func getNextDaysReadPath(t time.Time) string {
	return t.Add(time.Hour * 24).Format("200601/20060102")
}

func (p *Parser) Run() {

	go func() {
		// 定时刷新文件信息，新写完的文件放入队列
		ticker := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-ticker.C:
				p.FlushFiles()
			}
		}
	}()

	// 开启多通道读文件，最少24线程处理
	if p.conf.MaxParseProcs <= 0 && runtime.NumCPU() <= 12 {
		p.conf.MaxParseProcs = 24
	}
	for i := 0; i < p.conf.MaxParseProcs; i++ {
		go p.parseLoop()
	}
}

func (p *Parser) parseLoop() {
	var fix FileInfoX
	bufReader := bufio.NewReaderSize(nil, 10<<20)
	for {
		// 阻塞获取一个读文件任务
		select {
		case fix = <-p.Files:
			p.parseFile(fix, bufReader)
		}
	}
}

func (p *Parser) SetParseFunc(fun func(interface{})) {
	p.mu.Lock()
	p.parseFunc = fun
	p.mu.Unlock()
}

func (p *Parser) parseFile(x FileInfoX, bufReader *bufio.Reader) {

	p.numWorkers.Add(1)

	// 处理文件流程
	fr, err := os.OpenFile(x.filepath, os.O_RDONLY, 0)
	defer fr.Close()
	if err != nil {
		log.Errorf("os.OpenFile '%s' error: %s", x.filepath, err)
		return
	}

	t := time.Now()
	log.Debugf("[%2d/%d] - '%s' opened", p.numWorkers.Load(), p.conf.MaxParseProcs,
		x.filepath[len(p.conf.RootPath)+1:])

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
		p.parseFunc(line)
	}
	p.numWorkers.Sub(1)
	log.Debugf("[%d/%d] - '%s' parsed done. %.1fs used, %d files in total", p.numWorkers.Load(),
		p.conf.MaxParseProcs, x.filepath[len(p.conf.RootPath)+1:], time.Since(t).Seconds(), p.numFiles.Add(1))
}

//定义一个通用的结构体
type Bucket struct {
	Slice []FileInfoX                 //承载以任意结构体为元素构成的Slice
	By    func(a, b interface{}) bool //排序规则函数,当需要对新的结构体slice进行排序时，只需定义这个函数即可
}

/*
定义三个必须方法的准则：接收者不能为指针
*/
func (this Bucket) Len() int { return len(this.Slice) }

func (this Bucket) Swap(i, j int) { this.Slice[i], this.Slice[j] = this.Slice[j], this.Slice[i] }

func (this Bucket) Less(i, j int) bool { return this.By(this.Slice[i], this.Slice[j]) }
