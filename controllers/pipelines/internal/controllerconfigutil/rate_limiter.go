package controllerconfigutil

import (
	"k8s.io/client-go/util/workqueue"
	"time"
)

// This rate limiter is used to limit the rate of retry requests.
// It is set to requeue after 5 minutes for the first 10 retries, and then every 30 minutes thereafter.
var RateLimiter = workqueue.NewItemFastSlowRateLimiter(1*time.Minute, 30*time.Minute, 10)
