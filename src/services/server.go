package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/astaxie/beego/logs"
	"github.com/hpcloud/tail"
)

// TailObj is TailMgr's instance
type TailObj struct {
	tail     *tail.Tail
	offset   int64
	logConf  LogConfig
	secLimit *SecondLimit
	exitChan chan bool
}

var tailMgr *TailMgr

//TailMgr to manage tailObj
type TailMgr struct {
	tailObjMap map[string]*TailObj
	lock       sync.Mutex
}

// NewTailMgr init TailMgr obj
func NewTailMgr() *TailMgr {
	return &TailMgr{
		tailObjMap: make(map[string]*TailObj, 16),
	}
}

//AddLogFile to Add tail obj
func (t *TailMgr) AddLogFile(conf LogConfig) (err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	fmt.Println("add log file:", conf)
	_, ok := t.tailObjMap[conf.LogPath]
	if ok {
		err = fmt.Errorf("duplicate filename:%s", conf.LogPath)
		return
	}

	tail, err := tail.TailFile(conf.LogPath, tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // read to tail
		MustExist: false,  //file does not exist, it does not return an error 
		Poll:      true,
	})
	if err != nil {
		fmt.Println("tail file err:", err)
		return
	}

	tailObj := &TailObj{
		tail:     tail,
		offset:   0,
		logConf:  conf,
		secLimit: NewSecondLimit(int32(conf.SendRate)),
		exitChan: make(chan bool, 1),
	}
	t.tailObjMap[conf.LogPath] = tailObj

	waitGroup.Add(1)
	go tailObj.readLog()
	return
}

func (t *TailMgr) reloadConfig(logConfArr []LogConfig) (err error) {
	for _, conf := range logConfArr {
		tailObj, ok := t.tailObjMap[conf.LogPath]

		if !ok {
			err = t.AddLogFile(conf)
			if err != nil {
				logs.Error("add log file failed:%v", err)
				continue
			}
			continue
		}
		tailObj.logConf = conf
		tailObj.secLimit.limit = int32(conf.SendRate)
		t.tailObjMap[conf.LogPath] = tailObj
		fmt.Println("tailObj:", tailObj)
	}

	for key, tailObj := range t.tailObjMap {
		var found = false
		for _, newValue := range logConfArr {
			if key == newValue.LogPath {
				found = true
				break
			}
		}
		if found == false {
			logs.Warn("log path :%s is remove", key)
			tailObj.exitChan <- true
			delete(t.tailObjMap, key)
		}
	}
	return
}

// Process hava two func get new log conf and reload conf
func (t *TailMgr) Process() {
	for conf := range GetEtcdConfChan() {
		logs.Debug("log conf: %v", conf)
		fmt.Printf("log conf: %v", conf)

		var logConfArr []LogConfig

		err := json.Unmarshal([]byte(conf), &logConfArr)
		fmt.Println("logConfArr: ", logConfArr)

		var logConfArr1 LogConfig

		 json.Unmarshal([]byte(conf), &logConfArr1)
		fmt.Println("logConfArr: ", logConfArr1)

		if err != nil {
			logs.Error("unmarshal failed, err: %v conf :%s", err, conf)
			fmt.Println("unmarshal failed, err: %v conf :%s", err, conf)
			continue
		}

		err = t.reloadConfig(logConfArr)
		if err != nil {
			logs.Error("reload config from etcd failed: %v", err)
			continue
		}
		//logs.Debug("reload config from etcd success")
		fmt.Printf("reload config from etcd success")
	}
}

func (t *TailObj) readLog() {

	for line := range t.tail.Lines {
		if line.Err != nil {
			logs.Error("read line error:%v ", line.Err)
			continue
		}

		lineStr := strings.TrimSpace(line.Text)
		fmt.Println("readLog :", lineStr)

		if len(lineStr) == 0 || lineStr[0] == '\n' {
			continue
		}

		kafkaSend.addMessage(line.Text, t.logConf.Topic)
		t.secLimit.Add(1)
		t.secLimit.Wait()

		select {
		case <-t.exitChan:
			logs.Warn("tail obj is exited: config:", t.logConf)
			return
		default:
		}
	}
	waitGroup.Done()
}

func runServer() {
	tailMgr = NewTailMgr()
	tailMgr.Process()
	waitGroup.Wait()
}
