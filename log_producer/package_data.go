package log_producer

import (
	aliyun_log "github.com/aliyun/aliyun-log-go-sdk"
	"sync"
	"time"
)

type ILogCallback interface {
	OnCompletion(err error)
	SetSendBeginTimeInMillis(t int64)
	SetSendEndTimeInMillis(t int64)
	SetAddToIOQueueBeginTimeInMillis(t int64)
	SetAddToIOQueueEndTimeInMillis(t int64)
	SetCompleteIOBeginTimeInMillis(t int64)
	SetCompleteIOEndTimeInMillis(t int64)
	SetIOQueueSize(size int)
	SetSendBytesPerSecond(bps int)
}

type PackageData struct {
	ProjectName  string
	LogstoreName string
	ShardHash    string

	ArriveTimeInMS int64
	LogLinesCount  int
	PackageBytes   int
	Lock           *sync.Mutex

	Logstore    *aliyun_log.LogStore
	LogGroup    *aliyun_log.LogGroup
	Callbacks   []ILogCallback
	SendToQueue bool
}

func (p *PackageData) addLogs(logs []*aliyun_log.Log, callback ILogCallback) {
	tmp := p.LogGroup.Logs
	for _, log := range logs {
		tmp = append(tmp, log)
	}

	p.LogGroup.Logs = tmp

	if callback != nil {
		if p.Callbacks == nil {
			p.Callbacks = []ILogCallback{}
			p.Callbacks = append(p.Callbacks, callback)
		}
	}
}

func (p *PackageData) Clear() {
	p.ArriveTimeInMS = time.Now().UnixNano() / (1000 * 1000)
	p.LogLinesCount = 0
	p.PackageBytes = 0
	p.LogGroup = nil
	p.Callbacks = []ILogCallback{}
	p.SendToQueue = false
}

func (p *PackageData) Callback(err error, srcOutFlow float32) {
	curr := time.Now().UnixNano() / (1000 * 1000)

	for _, cb := range p.Callbacks {
		cb.SetCompleteIOEndTimeInMillis(curr)
		cb.SetSendBytesPerSecond(int(srcOutFlow))
		cb.OnCompletion(err)
	}
}

func (p *PackageData) MarkAddToIOBeginTime() {
	curr := time.Now().Unix()

	for _, cb := range p.Callbacks {
		cb.SetAddToIOQueueBeginTimeInMillis(curr)
		Info.Println("markAddToIOBeginTime %s %v", p.Logstore, cb)
	}
}

func (p *PackageData) MarkAddToIOEndTime() {
	curr := time.Now().Unix()

	for _, cb := range p.Callbacks {
		cb.SetAddToIOQueueEndTimeInMillis(curr)
		Info.Println("markAddToIOEndTime %s %v", p.Logstore, cb)
	}
}

func (p *PackageData) MarkCompleteIOBeginTimeInMillis(queueSize int) {
	curr := time.Now().Unix()

	for _, cb := range p.Callbacks {
		cb.SetCompleteIOBeginTimeInMillis(curr)
		cb.SetIOQueueSize(queueSize)
		Info.Println("%v markCompleteIOBeginTimeInMillis %s %v", curr, p.Logstore, cb)
	}
}
