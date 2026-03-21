package activitylog

import (
	"context"
	"fmt"

	domainactivitylog "github.com/AbenezerWork/ProcureFlow/internal/domain/activitylog"
)

func NotifyHooks(ctx context.Context, entry domainactivitylog.Entry, hooks ...Hook) error {
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if err := hook.HandleActivityLog(ctx, entry); err != nil {
			return fmt.Errorf("notify activity log hook: %w", err)
		}
	}

	return nil
}
