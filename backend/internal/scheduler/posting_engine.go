package scheduler

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"fbscheduler/internal/db"
	"fbscheduler/internal/facebook"
)

// ============================================
// CONSTANTS
// ============================================

const (
	// Cooldown sau mỗi bài cùng nick (giây)
	CooldownAfterPostSeconds = 30

	// Retry delays (phút)
	RetryDelay1Minutes = 2
	RetryDelay2Minutes = 5

	// Số request song song tối đa mỗi nick
	MaxConcurrentPerAccount = 3
)

// ============================================
// POSTING ENGINE
// ============================================

// PostingEngine xử lý việc đăng bài với rate limiting và retry
type PostingEngine struct {
	store    *db.Store
	fbClient *facebook.Client

	// Track last post time per account for cooldown
	accountLastPost map[string]time.Time
	accountMu       sync.RWMutex

	// Semaphore per account for concurrent limit
	accountSem map[string]chan struct{}
	semMu      sync.Mutex
}

// NewPostingEngine tạo posting engine mới
func NewPostingEngine(store *db.Store) *PostingEngine {
	return &PostingEngine{
		store:           store,
		fbClient:        facebook.NewClient(),
		accountLastPost: make(map[string]time.Time),
		accountSem:      make(map[string]chan struct{}),
	}
}

// PublishPost đăng 1 bài với rate limiting và retry
func (e *PostingEngine) PublishPost(sp db.ScheduledPost) error {
	// Lấy account để đăng bài
	account, accessToken, err := e.getAccountForPost(sp)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	accountID := ""
	if account != nil {
		accountID = account.ID

		// Acquire semaphore (giới hạn concurrent)
		sem := e.getAccountSemaphore(accountID)
		sem <- struct{}{}
		defer func() { <-sem }()

		// Wait for cooldown
		e.waitForCooldown(accountID)
	}

	// Update status to processing
	if err := e.store.UpdateScheduledPostStatus(sp.ID, "processing"); err != nil {
		log.Printf("❌ Error updating status: %v", err)
		return err
	}

	// Post to Facebook
	fbPostID, err := e.fbClient.PostToPage(
		sp.Page.PageID,
		accessToken,
		sp.Post.Content,
		sp.Post.MediaURLs,
		sp.Post.MediaType,
	)

	// Create log entry
	logEntry := &db.PostLog{
		ScheduledPostID: sp.ID,
		PostID:          sp.PostID,
		PageID:          sp.PageID,
	}

	if err != nil {
		return e.handlePostError(sp, account, logEntry, err)
	}

	// Success
	return e.handlePostSuccess(sp, account, logEntry, fbPostID)
}

// getAccountForPost lấy account và access token để đăng bài
func (e *PostingEngine) getAccountForPost(sp db.ScheduledPost) (*db.FacebookAccount, string, error) {
	// Thử lấy account từ scheduled_post (nếu đã được assign)
	// TODO: Cần thêm account_id vào ScheduledPost struct

	// Fallback: Lấy best account cho page
	account, err := e.store.GetBestAccountForPage(sp.PageID)
	if err == nil && account != nil {
		// Lấy access token từ page (vì page token khác user token)
		page, err := e.store.GetPageByID(sp.PageID)
		if err != nil {
			return account, "", err
		}
		return account, page.AccessToken, nil
	}

	// Fallback: Dùng access token của page trực tiếp
	if sp.Page != nil && sp.Page.AccessToken != "" {
		return nil, sp.Page.AccessToken, nil
	}

	return nil, "", fmt.Errorf("no access token available for page %s", sp.PageID)
}

// getAccountSemaphore lấy hoặc tạo semaphore cho account
func (e *PostingEngine) getAccountSemaphore(accountID string) chan struct{} {
	e.semMu.Lock()
	defer e.semMu.Unlock()

	if sem, ok := e.accountSem[accountID]; ok {
		return sem
	}

	sem := make(chan struct{}, MaxConcurrentPerAccount)
	e.accountSem[accountID] = sem
	return sem
}

// waitForCooldown chờ cooldown nếu cần
func (e *PostingEngine) waitForCooldown(accountID string) {
	e.accountMu.RLock()
	lastPost, ok := e.accountLastPost[accountID]
	e.accountMu.RUnlock()

	if !ok {
		return
	}

	elapsed := time.Since(lastPost)
	cooldown := time.Duration(CooldownAfterPostSeconds) * time.Second

	if elapsed < cooldown {
		waitTime := cooldown - elapsed
		log.Printf("⏳ Waiting %.1f seconds for cooldown (account: %s)", waitTime.Seconds(), accountID[:8])
		time.Sleep(waitTime)
	}
}

// updateLastPostTime cập nhật thời gian post cuối
func (e *PostingEngine) updateLastPostTime(accountID string) {
	e.accountMu.Lock()
	e.accountLastPost[accountID] = time.Now()
	e.accountMu.Unlock()
}

// handlePostSuccess xử lý khi đăng bài thành công
func (e *PostingEngine) handlePostSuccess(sp db.ScheduledPost, account *db.FacebookAccount, logEntry *db.PostLog, fbPostID string) error {
	log.Printf("✅ Successfully posted to page %s: %s", sp.Page.PageID, fbPostID)

	// Update scheduled post status
	e.store.UpdateScheduledPostStatus(sp.ID, "completed")

	// Update account_id for tracking who posted
	if account != nil {
		if err := e.store.UpdateScheduledPostAccount(sp.ID, account.ID); err != nil {
			log.Printf("⚠️ Error updating account_id: %v", err)
		}
	}

	// Update log
	logEntry.Status = "success"
	logEntry.FacebookPostID = fbPostID
	if err := e.store.CreatePostLog(logEntry); err != nil {
		log.Printf("❌ Error creating log: %v", err)
	}

	// Update account stats
	if account != nil {
		e.updateLastPostTime(account.ID)
		if err := e.store.RecordSuccessfulPost(account.ID, sp.PageID); err != nil {
			log.Printf("⚠️ Error recording successful post: %v", err)
		}

		// Check warning threshold (80%)
		e.checkWarningThreshold(account)
	}

	return nil
}

// handlePostError xử lý khi đăng bài thất bại
func (e *PostingEngine) handlePostError(sp db.ScheduledPost, account *db.FacebookAccount, logEntry *db.PostLog, postErr error) error {
	log.Printf("❌ Failed to post to page %s: %v", sp.Page.PageID, postErr)

	// Check if rate limit error
	isRateLimit := e.isRateLimitError(postErr)

	// Update account stats
	if account != nil {
		if err := e.store.RecordPostFailure(account.ID, isRateLimit); err != nil {
			log.Printf("⚠️ Error recording post failure: %v", err)
		}

		// Create notification if rate limit
		if isRateLimit {
			e.store.NotifyRateLimit(account.ID, account.FbUserName)
		}
	}

	// Determine retry strategy
	retryDelay := e.getRetryDelay(sp.RetryCount)

	if retryDelay > 0 {
		// Schedule retry
		e.store.IncrementRetryCount(sp.ID)
		e.store.UpdateScheduledPostStatus(sp.ID, "pending")

		// Update scheduled_time for retry
		newTime := time.Now().Add(retryDelay)
		e.updateScheduledTime(sp.ID, newTime)

		log.Printf("🔄 Retry %d/3 scheduled in %v for post %s",
			sp.RetryCount+1, retryDelay, sp.ID)
	} else {
		// Max retries reached
		e.store.UpdateScheduledPostStatus(sp.ID, "failed")
		log.Printf("💀 Max retries reached for post %s", sp.ID)

		// Create notification
		if account != nil {
			pageName := ""
			if sp.Page != nil {
				pageName = sp.Page.PageName
			}
			e.store.NotifyPostFailed(account.ID, account.FbUserName, pageName, postErr.Error())
		}
	}

	// Save log
	logEntry.Status = "failed"
	logEntry.ErrorMessage = postErr.Error()
	if err := e.store.CreatePostLog(logEntry); err != nil {
		log.Printf("❌ Error creating log: %v", err)
	}

	return postErr
}

// getRetryDelay trả về delay cho retry tiếp theo
func (e *PostingEngine) getRetryDelay(currentRetryCount int) time.Duration {
	switch currentRetryCount {
	case 0:
		return time.Duration(RetryDelay1Minutes) * time.Minute
	case 1:
		return time.Duration(RetryDelay2Minutes) * time.Minute
	default:
		return 0 // No more retries
	}
}

// isRateLimitError kiểm tra có phải lỗi rate limit không
func (e *PostingEngine) isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many") ||
		strings.Contains(errStr, "limit reached") ||
		strings.Contains(errStr, "code: 4") ||
		strings.Contains(errStr, "code: 17") ||
		strings.Contains(errStr, "code: 32")
}

// updateScheduledTime cập nhật thời gian schedule
func (e *PostingEngine) updateScheduledTime(spID string, newTime time.Time) {
	query := `UPDATE scheduled_posts SET scheduled_time = $1 WHERE id = $2`
	e.store.DB().Exec(query, newTime, spID)
}

// checkWarningThreshold kiểm tra và tạo cảnh báo nếu đạt 80%
func (e *PostingEngine) checkWarningThreshold(account *db.FacebookAccount) {
	// Refresh account data
	refreshed, err := e.store.GetAccountByID(account.ID)
	if err != nil {
		return
	}

	percentage := float64(refreshed.PostsToday) / float64(refreshed.MaxPostsPerDay) * 100

	// Check 80% threshold
	if percentage >= 80 && percentage < 100 {
		// Check if we already sent warning today (simple check)
		if refreshed.PostsToday == int(float64(refreshed.MaxPostsPerDay)*0.8) {
			e.store.NotifyWarningThreshold(account.ID, account.FbUserName,
				refreshed.PostsToday, refreshed.MaxPostsPerDay)
		}
	}

	// Check 100% threshold
	if percentage >= 100 {
		if refreshed.PostsToday == refreshed.MaxPostsPerDay {
			e.store.NotifyDailyLimit(account.ID, account.FbUserName)
		}
	}
}
