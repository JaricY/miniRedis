package timewheel

import (
	"container/list"
	"miniRedis/lib/logger"
	"time"
)

// location 记录一个任务的定时器位置
type location struct {
	slot  int           //任务所在的时间论的槽
	etask *list.Element //在槽内的位置
}

// TimeWheel 时间轮，用于实现在固定时间间隔内执行任务
type TimeWheel struct {
	interval time.Duration //时间间隔，例如设置为1秒
	ticker   *time.Ticker  //用于触发时间论的计时器
	slots    []*list.List  //时间轮的每个时间槽，每个时间槽存储了在该时间槽期间需要执行的任务

	timer             map[string]*location // timer存储了时间轮中已添加任务的信息，用于查询和移除任务
	currentPos        int                  //currentPos是时间轮的当前位置
	slotNum           int                  //时间轮的时间槽数量
	addTaskChannel    chan task            //添加任务的通道
	removeTaskChannel chan string          //删除任务的通道
	stopChannel       chan bool            //停止计时器的通道
}

// 表示时间轮中的任务
type task struct {
	delay  time.Duration //延迟时间，即任务需要延迟多长时间之后执行。
	circle int           //时间轮转动的圈数，即任务需要在时间轮转动多少圈之后执行
	key    string        //任务关联的键
	job    func()        //任务的具体执行函数
}

// New 创建一个新的时间轮，interval表示时间轮的槽间隔时间，slotNum表示槽的数量
func New(interval time.Duration, slotNum int) *TimeWheel {
	if interval <= 0 || slotNum <= 0 {
		return nil
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		timer:             make(map[string]*location),
		currentPos:        0,
		slotNum:           slotNum,
		addTaskChannel:    make(chan task),
		removeTaskChannel: make(chan string),
		stopChannel:       make(chan bool),
	}
	tw.initSlots()

	return tw
}

func (tw *TimeWheel) initSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

// Start 启动时间轮
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

func (tw *TimeWheel) Stop() {
	tw.stopChannel <- true
}

// AddJob 添加一个任务
func (tw *TimeWheel) AddJob(delay time.Duration, key string, job func()) {
	if delay < 0 {
		return
	}
	tw.addTaskChannel <- task{delay: delay, key: key, job: job}
}

func (tw *TimeWheel) RemoveJob(key string) {
	if key == "" {
		return
	}
	tw.removeTaskChannel <- key
}

func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(&task)
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheel) addTask(task *task) {
	// 获取当前任务需要在时间轮中的位置
	pos, circle := tw.getPositionAndCircle(task.delay)
	task.circle = circle

	// 将当前元素从插入到槽后返回在这个槽中的位置
	e := tw.slots[pos].PushBack(task)

	loc := &location{
		slot:  pos,
		etask: e,
	}

	if task.key != "" {
		_, ok := tw.timer[task.key]
		// 如果已经存在，则覆盖这个任务
		if ok {
			tw.removeTask(task.key)
		}
	}
	tw.timer[task.key] = loc
}

// getPositionAndCircle 该函数的作用是计算定时任务在时间轮上的位置和所在圈数。
func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos int, circle int) {
	//
	delaySeconds := int(d.Seconds())

	// 获取时间轮的时间间隔
	intervalSeconds := int(tw.interval.Seconds())

	// 计算应该转多少圈
	circle = int(delaySeconds / intervalSeconds / tw.slotNum)

	// 得到在时间轮上的槽的位置
	pos = int(tw.currentPos+delaySeconds/intervalSeconds) % tw.slotNum

	return
}

func (tw *TimeWheel) tickHandler() {
	// 获取到当前时刻所在的槽
	l := tw.slots[tw.currentPos]
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
	// 扫描当前槽内的任务并执行
	go tw.scanAndRunTask(l)
}

func (tw *TimeWheel) scanAndRunTask(l *list.List) {

	// 开始遍历这个槽
	for e := l.Front(); e != nil; {
		task := e.Value.(*task)
		// 如果当前任务需要循环执行，则将其循环次数减一
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}
		// 如果当前任务剩余循环圈数为0，则异步执行其job方法
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(err)
				}
			}()
			job := task.job
			job()
		}()

		next := e.Next()

		//移除这个任务
		l.Remove(e)
		if task.key != "" {
			// 从timer中删除这个任务
			delete(tw.timer, task.key)
		}
		e = next
	}
}

// removeTask 移除任务
func (tw *TimeWheel) removeTask(key string) {
	pos, ok := tw.timer[key]
	if !ok {
		return
	}
	l := tw.slots[pos.slot]
	l.Remove(pos.etask)
	delete(tw.timer, key)
}
