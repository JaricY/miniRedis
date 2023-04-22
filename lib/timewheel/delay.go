package timewheel

import "time"

var tw = New(time.Second, 3600)

func init() {
	tw.Start()
}

// Delay 添加一个任务到时间轮中，传入过期时间，key以及处理函数
func Delay(duration time.Duration, key string, job func()) {
	tw.AddJob(duration, key, job)
}

// At 添加一个任务到时间轮中，传入过期的时刻，key以及处理函数
func At(at time.Time, key string, job func()) {
	// at.Sub(time.Now()) 计算执行时间与当前时间的差值，也就是延迟删除的时间
	tw.AddJob(at.Sub(time.Now()), key, job)
}

// Cancel 移除一个任务
func Cancel(key string) {
	tw.RemoveJob(key)
}
