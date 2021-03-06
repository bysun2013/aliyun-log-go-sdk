package main

import (
	"fmt"
	aliyun_log "github.com/aliyun/aliyun-log-go-sdk"
	log_producer "github.com/aliyun/aliyun-log-go-sdk/log_producer"
	"github.com/gogo/protobuf/proto"
	"log"
	"math/rand"
	"sync"
	"time"
)

var project = &aliyun_log.LogProject{
	Name:            "test-project-for-go-producer",
	Endpoint:        "cn-beijing.log.aliyuncs.com",
	AccessKeyID:     "LTAIuVFxy8FHaGWD",
	AccessKeySecret: "8QzfYMDxxiyFZDzJ2m1W2YtOJcVi0c",
}

func main() {
	log.Println("loghub sample begin")

	begin_time := uint32(time.Now().Unix())
	rand.Seed(int64(begin_time))
	logstore_name := "test-logstore"
	// logstore_name2 := "sls-test-2"

	project_pool := &log_producer.ProjectPool{}
	project_pool.UpdateProject(project)

	producer := log_producer.LogProducer{}
	producer.Init(project_pool, &log_producer.DefaultGlobalProducerConfig)
	defer producer.Destroy()

	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go sendLogs(&producer, logstore_name, fmt.Sprintf("test-log-%d", i), fmt.Sprintf("10.10.10.%d", i), "", wg)
	}

	wg.Wait()
	// time.Sleep(5 * time.Second)
	log.Println("loghub sample end")
}

func sendLogs(producer *log_producer.LogProducer, logstore_name string, topic string, source string, shardHash string, wg *sync.WaitGroup) {
	defer wg.Done()

	for loggroupIdx := 0; loggroupIdx < 500; loggroupIdx++ {
		logs := []*aliyun_log.Log{}
		for logIdx := 0; logIdx < 3; logIdx++ {
			content := []*aliyun_log.LogContent{}
			for colIdx := 0; colIdx < 3; colIdx++ {
				content = append(content, &aliyun_log.LogContent{
					Key:   proto.String(fmt.Sprintf("col_%d", colIdx)),
					Value: proto.String(fmt.Sprintf("loggroup idx: %d, log idx: %d, col idx: %d, value: %d", loggroupIdx, logIdx, colIdx, rand.Intn(10000000))),
				})
			}
			log := &aliyun_log.Log{
				Time:     proto.Uint32(uint32(time.Now().Unix())),
				Contents: content,
			}
			logs = append(logs, log)
		}
		loggroup := &aliyun_log.LogGroup{
			Topic:  proto.String(topic),
			Source: proto.String(source),
			Logs:   logs,
		}

		callback := &log_producer.DefaultLogCallback{
			Producer:     producer,
			ProjectName:  project.Name,
			LogstoreName: logstore_name,
			ShardHash:    shardHash,
			Loggroup:     loggroup,
		}
		err := producer.Send(project.Name, logstore_name, shardHash, loggroup, callback)
		if err != nil {
			log.Println(err)
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	log.Println("[test] send log routine end, ", topic, source)
}
