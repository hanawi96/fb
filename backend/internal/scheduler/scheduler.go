package scheduler

import (
	"fbscheduler/internal/db"
	"fbscheduler/internal/facebook"
	"log"
	"time"
)

type Scheduler struct {
	store    *db.Store
	fbClient *facebook.Client
	ticker   *time.Ticker
}

func NewScheduler(store *db.Store) *Scheduler {
	return &Scheduler{
		store:    store,
		fbClient: facebook.NewClient(),
		ticker:   time.NewTicker(30 * time.Second),
	}
}

func (s *Scheduler) Start() {
	log.Println("📅 Scheduler: Checking for pending posts every 30 seconds...")
	
	for range s.ticker.C {
		s.processPendingPosts()
	}
}

func (s *Scheduler) Stop() {
	s.ticker.Stop()
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
	
	for _, sp := range posts {
		go s.publishPost(sp)
	}
}

func (s *Scheduler) publishPost(sp db.ScheduledPost) {
	// Update status to processing
	if err := s.store.UpdateScheduledPostStatus(sp.ID, "processing"); err != nil {
		log.Printf("❌ Error updating status: %v", err)
		return
	}
	
	// Post to Facebook
	fbPostID, err := s.fbClient.PostToPage(
		sp.Page.PageID,
		sp.Page.AccessToken,
		sp.Post.Content,
		sp.Post.MediaURLs,
	)
	
	logEntry := &db.PostLog{
		ScheduledPostID: sp.ID,
		PostID:          sp.PostID,
		PageID:          sp.PageID,
	}
	
	if err != nil {
		log.Printf("❌ Failed to post to page %s: %v", sp.Page.PageID, err)
		
		// Check if should retry
		if sp.RetryCount < sp.MaxRetries {
			s.store.IncrementRetryCount(sp.ID)
			s.store.UpdateScheduledPostStatus(sp.ID, "pending")
			log.Printf("🔄 Retry %d/%d for post %s", sp.RetryCount+1, sp.MaxRetries, sp.ID)
		} else {
			s.store.UpdateScheduledPostStatus(sp.ID, "failed")
			log.Printf("💀 Max retries reached for post %s", sp.ID)
		}
		
		logEntry.Status = "failed"
		logEntry.ErrorMessage = err.Error()
	} else {
		log.Printf("✅ Successfully posted to page %s: %s", sp.Page.PageID, fbPostID)
		s.store.UpdateScheduledPostStatus(sp.ID, "completed")
		
		logEntry.Status = "success"
		logEntry.FacebookPostID = fbPostID
	}
	
	// Save log
	if err := s.store.CreatePostLog(logEntry); err != nil {
		log.Printf("❌ Error creating log: %v", err)
	}
}
