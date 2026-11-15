package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

/*
  آموزش همزمانی (Concurrency) در Go

  اجرا:
    go run ./cmd/ss -e 1   # Goroutine
    go run ./cmd/ss -e 2   # Channel بدون بافر
    go run ./cmd/ss -e 3   # Channel با بافر
    go run ./cmd/ss -e 4   # WaitGroup
    go run ./cmd/ss -e 5   # Mutex (race بدون قفل)
    go run ./cmd/ss -e 6   # Mutex (امن)
    go run ./cmd/ss -e 7   # Select
    go run ./cmd/ss -e 8   # Worker Pool (کاربردی)
    go run ./cmd/ss -e 9   # Pipeline (کاربردی)
*/

func main() {
	example := flag.Int("e", 1, "شماره مثال (1-9)")
	flag.Parse()

	fmt.Printf("\n=== مثال %d ===\n\n", *example)

	switch *example {
	case 1:
		ex1Goroutine()
	case 2:
		ex2UnbufferedChannel()
	case 3:
		ex3BufferedChannel()
	case 4:
		ex4WaitGroup()
	case 5:
		ex5RaceWithoutMutex()
	case 6:
		ex6SafeWithMutex()
	case 7:
		ex7Select()
	case 8:
		ex8WorkerPool()
	case 9:
		ex9Pipeline()
	default:
		fmt.Println("مثال معتبر نیست. از 1 تا 9 انتخاب کن.")
	}
}

// ---------------------------------------------------------------------------
// 1) Goroutine: اجرای همزمان توابع با کلمه کلیدی go
// ---------------------------------------------------------------------------
// Goroutine یک thread سبک است که runtime گولنگ مدیریتش می‌کند.
// بدون sync (مثل WaitGroup یا channel) main ممکن است زودتر تمام شود.

func ex1Goroutine() {
	go say("goroutine A")
	go say("goroutine B")
	say("main") // این روی همان goroutine اصلی اجرا می‌شود

	time.Sleep(500 * time.Millisecond) // فقط برای دیدن خروجی؛ در عمل از WaitGroup استفاده کن
}

func say(who string) {
	for i := 1; i <= 3; i++ {
		fmt.Printf("[%s] tick %d\n", who, i)
		time.Sleep(80 * time.Millisecond)
	}
}

// ---------------------------------------------------------------------------
// 2) Unbuffered Channel: ارسال و دریافت همزمان (handshake)
// ---------------------------------------------------------------------------
// ch <- x  بلاک می‌شود تا کسی از کانال بخواند.
// <-ch     بلاک می‌شود تا کسی چیزی بفرستد.
// close(ch) یعنی دیگر مقداری ارسال نمی‌شود؛ range تمام می‌شود.

func ex2UnbufferedChannel() {
	ch := make(chan string) // بدون بافر

	go func() {
		for _, msg := range []string{"task-1", "task-2", "task-3"} {
			fmt.Println("ارسال:", msg)
			ch <- msg // منتظر می‌ماند تا consumer بخواند
			time.Sleep(100 * time.Millisecond)
		}
		close(ch) // خیلی مهم: وگرنه range در consumer تا ابد بلاک می‌ماند
	}()

	for msg := range ch {
		fmt.Println("دریافت:", msg)
	}
	fmt.Println("کانال بسته شد، کار تمام.")
}

// ---------------------------------------------------------------------------
// 3) Buffered Channel: ارسال بدون منتظر ماندن تا پر شدن بافر
// ---------------------------------------------------------------------------
// make(chan T, n) تا n مقدار را بدون گیرنده نگه می‌دارد.
// وقتی بافر پر است، ارسال بلاک می‌شود.

func ex3BufferedChannel() {
	ch := make(chan int, 3) // ظرفیت 3

	ch <- 10
	ch <- 20
	ch <- 30
	// ch <- 40 // این خط بلاک می‌شود چون بافر پر است و گیرنده‌ای نیست

	fmt.Println("طول بافر الان:", len(ch), "/ ظرفیت:", cap(ch))

	fmt.Println(<-ch)
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}

// ---------------------------------------------------------------------------
// 4) WaitGroup: منتظر ماندن تا همه goroutineها تمام شوند
// ---------------------------------------------------------------------------
// Add(n) قبل از go، Done() داخل goroutine (معمولاً با defer)، Wait() در main.

func ex4WaitGroup() {
	var wg sync.WaitGroup
	jobs := []string{"download", "resize", "upload"}

	for _, job := range jobs {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			fmt.Println("شروع:", name)
			time.Sleep(200 * time.Millisecond)
			fmt.Println("پایان:", name)
		}(job) // job را به عنوان آرگومان پاس بده (جلوگیری از closure bug)
	}

	wg.Wait()
	fmt.Println("همه jobها تمام شدند.")
}

// ---------------------------------------------------------------------------
// 5) Race Condition بدون Mutex
// ---------------------------------------------------------------------------
// چند goroutine همزمان روی یک متغیر می‌نویسند → نتیجه اشتباه.
// برای دیدن race واقعی: go run -race ./cmd/ss -e 5

func ex5RaceWithoutMutex() {
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // ناامن!
		}()
	}

	wg.Wait()
	fmt.Println("انتظار: 1000 | واقعیت:", counter)
	fmt.Println("(معمولاً کمتر از 1000 است چون race داریم)")
}

// ---------------------------------------------------------------------------
// 6) Mutex: قفل برای دسترسی امن به حافظه مشترک
// ---------------------------------------------------------------------------
// Lock / Unlock دور بخش بحرانی (critical section).
// همیشه Unlock را با defer بگذار تا در panic هم آزاد شود.
// RWMutex برای خواندن زیاد/نوشتن کم بهتر است.

func ex6SafeWithMutex() {
	counter := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println("انتظار: 1000 | واقعیت:", counter)
}

// ---------------------------------------------------------------------------
// 7) Select: منتظر ماندن روی چند کانال همزمان
// ---------------------------------------------------------------------------
// مثل switch برای کانال‌ها. case آماده را اجرا می‌کند.
// default یعنی non-blocking. timeout با time.After رایج است.

func ex7Select() {
	fast := make(chan string, 1)
	slow := make(chan string, 1)

	go func() {
		time.Sleep(50 * time.Millisecond)
		fast <- "جواب سریع"
	}()
	go func() {
		time.Sleep(200 * time.Millisecond)
		slow <- "جواب کند"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg := <-fast:
			fmt.Println("از fast:", msg)
		case msg := <-slow:
			fmt.Println("از slow:", msg)
		case <-time.After(300 * time.Millisecond):
			fmt.Println("timeout")
		}
	}
}

// ---------------------------------------------------------------------------
// 8) Worker Pool — الگوی کاربردی واقعی
// ---------------------------------------------------------------------------
// چند worker موازی job می‌گیرند، نتیجه می‌فرستند.
// مناسب: پردازش تصویر، ارسال ایمیل، فراخوانی API، ...

type job struct {
	ID   int
	Data string
}

type result struct {
	JobID int
	Out   string
}

func ex8WorkerPool() {
	const workers = 3
	jobs := make(chan job, 10)
	results := make(chan result, 10)

	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go worker(w, jobs, results, &wg)
	}

	// تولید کار
	go func() {
		for i := 1; i <= 6; i++ {
			jobs <- job{ID: i, Data: fmt.Sprintf("item-%d", i)}
		}
		close(jobs)
	}()

	// بستن results وقتی همه workerها تمام شدند
	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		fmt.Printf("نتیجه job#%d => %s\n", r.JobID, r.Out)
	}
}

func worker(id int, jobs <-chan job, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		fmt.Printf("worker-%d شروع job#%d\n", id, j.ID)
		time.Sleep(120 * time.Millisecond) // شبیه‌سازی کار سنگین
		results <- result{
			JobID: j.ID,
			Out:   fmt.Sprintf("processed(%s) by worker-%d", j.Data, id),
		}
	}
}

// ---------------------------------------------------------------------------
// 9) Pipeline — جریان داده مرحله‌به‌مرحله با کانال
// ---------------------------------------------------------------------------
// stage1 → stage2 → stage3
// هر مرحله goroutine خودش را دارد و فقط از کانال‌ها حرف می‌زند.
// مناسب: ETL، تبدیل داده، فیلتر + map + aggregate

func ex9Pipeline() {
	nums := generate(1, 2, 3, 4, 5)
	squared := square(nums)
	filtered := filterEven(squared)

	for v := range filtered {
		fmt.Println("خروجی pipeline:", v)
	}
}

func generate(values ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, v := range values {
			out <- v
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			out <- v * v
		}
	}()
	return out
}

func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := range in {
			if v%2 == 0 {
				out <- v
			}
		}
	}()
	return out
}
