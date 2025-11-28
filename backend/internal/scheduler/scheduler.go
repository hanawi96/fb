package scheduler

import (
	"fbscheduler/internal/db"
	"log"
	"sync"
	"time"
)

// ============================================
// SCHEDULER
// Chạy background job để xử lý scheduled posts
// ============================================

type Scheduler struct {
	store         *db.Store
	postingEngine *PostingEngine
	ticker        *time.Ticker
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func NewScheduler(store *db.Store) *Scheduler {
	return &Scheduler{
		store:         store,
		postingEngine: NewPostingEngine(store),
		ticker:        time.NewTicker(30 * time.Second),
		stopChan:      make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	log.Println("📅 Scheduler: Checking for pending posts every 30 seconds...")

	// Run daily reset job
	go s.runDailyResetJob()

	for {
		select {
		case <-s.ticker.C:
			s.processPendingPosts()
		case <-s.stopChan:
			log.Println("📅 Scheduler: Stopped")
			return
		}
	}
}

func (s *Scheduler) Stop() {
	s.ticker.Stop()
	close(s.stopChan)
	s.wg.Wait()
}

func (s *Scheduler) processPendingPosts() {
	posts, err := s.store.GetPendingScheduledPosts()
	if err != nil {
		log.Printf("❌ Scheduler: Error fetching pending posts: %v", err)
		return
	}

	if len(posts) == 0 {
		return
	}

	log.Printf("📤 Scheduler: Found %d posts to publish", len(posts))

	// Group posts by account to respect rate limits
	postsByAccount := s.groupPostsByAccount(posts)

	// Process each account's posts
	for accountID, accountPosts := range postsByAccount {
		s.wg.Add(1)
		go func(accID string, posts []db.ScheduledPost) {
			defer s.wg.Done()
			s.processAccountPosts(accID, posts)
		}(accountID, accountPosts)
	}
}

// groupPostsByAccount nhóm posts theo account
func (s *Scheduler) groupPostsByAccount(posts []db.ScheduledPost) map[string][]db.ScheduledPost {
	result := make(map[string][]db.ScheduledPost)

	for _, sp := range posts {
		// Lấy account cho page này
		account, _ := s.store.GetBestAccountForPage(sp.PageID)

		accountID := "default" // Fallback nếu không có account
		if account != nil {
			accountID = account.ID
		}

		result[accountID] = append(result[accountID], sp)
	}

	return result
}

// processAccountPosts xử lý posts của 1 account (tuần tự với cooldown)
func (s *Scheduler) processAccountPosts(accountID string, posts []db.ScheduledPost) {
	for _, sp := range posts {
		// Check if scheduler is stopping
		select {
		case <-s.stopChan:
			return
		default:
		}

		// Publish post (PostingEngine sẽ xử lý cooldown và retry)
		err := s.postingEngine.PublishPost(sp)
		if err != nil {
			log.Printf("⚠️ Post %s failed: %v", sp.ID, err)
		}
	}
}

// runDailyResetJob chạy job reset counter hàng ngày
func (s *Scheduler) runDailyResetJob() {
	for {
		// Tính thời gian đến 00:00 ngày mai
		now := time.Now()
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		duration := nextMidnight.Sub(now)

		log.Printf("📅 Daily reset scheduled in %v", duration)

		select {
		case <-time.After(duration):
			log.Println("🔄 Running daily reset...")
			if err := s.store.ResetDailyPostCounts(); err != nil {
				log.Printf("❌ Error resetting daily counts: %v", err)
			} else {
				log.Println("✅ Daily post counts reset")
			}
		case <-s.stopChan:
			return
		}
	}
}
