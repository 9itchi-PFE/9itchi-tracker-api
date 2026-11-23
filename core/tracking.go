package core

import "time"

// OpenIgnoreWindow: opens this soon after sent_at are mail-provider prefetch/scanner hits and are not counted.
const OpenIgnoreWindow = 10 * time.Second
