package daemon

import (
	"context"
	"log/slog"
	"time"

	telemetryadapter "github.com/aoagents/agent-orchestrator/backend/internal/adapters/telemetry"
	"github.com/aoagents/agent-orchestrator/backend/internal/config"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
	"github.com/aoagents/agent-orchestrator/backend/internal/storage/sqlite"
)

// aggregatedEventNames are the event names prone to bursts (crash loops,
// retry storms) that get folded into one rollup event per minute instead of
// one PostHog event per occurrence, and correspondingly get the generous
// daily rate-limit tier since their per-occurrence cost is already gone (see
// RateLimitedSink's eventsPerNamePerDayAggregated). Must match the event
// names as actually emitted (see httpd/router.go and httpd/recover.go /
// httpd/log.go) exactly, including underscores - there is no compile-time
// check tying this list to the emit sites.
//
// Deliberately does NOT include ao.cli.invoked or ao.app.active: those are
// already deduped to at most once per command-path/day (ao.cli.invoked) or
// once per day (ao.app.active) at the httpd layer before they ever reach a
// sink, including for ao hooks/ao pty-host and high-frequency polling
// commands like ao status/ao session ls - adding them here would double up
// on dedup logic that already owns that job and gains nothing.
var aggregatedEventNames = []string{
	"ao.http.5xx",
	"ao.daemon.panic",
	"ao.cli.usage_errors",
}


// newTelemetrySink also returns the live telemetry identity (nil when the data
// dir is unknown), shared with the remote sink so a sign-in or opt-out seen by
// the identity routes is what the exporter stamps and honors.
func newTelemetrySink(cfg config.Config, store *sqlite.Store, log *slog.Logger) (ports.EventSink, ports.TelemetryIdentityStore) {
	if !cfg.Telemetry.Events {
		return telemetryadapter.NoopSink{}, standaloneIdentity(cfg, log)
	}
	return telemetryadapter.NewLocalSQLiteSink(store, log), standaloneIdentity(cfg, log)
}

// standaloneIdentity serves the identity routes when nothing is exported, so the
// phone still learns the opt-out state and the install ID.
func standaloneIdentity(cfg config.Config, log *slog.Logger) ports.TelemetryIdentityStore {
	if cfg.DataDir == "" {
		return nil
	}
	ident, err := telemetryadapter.NewIdentity(cfg.DataDir)
	if err != nil {
		log.Warn("telemetry identity unavailable", "error", err)
		return nil
	}
	return ident
}

// gitHubLoginRetryDelays space the attempts to resolve the operator's GitHub
// login after daemon start. A fresh install usually has no token yet, so the
// later attempts catch a PAT or `gh auth login` done shortly after launch.
// ponytail: fixed schedule, then give up until the next daemon start.
var gitHubLoginRetryDelays = []time.Duration{0, 30 * time.Second, 2 * time.Minute, 10 * time.Minute}

// startGitHubLoginResolver resolves the authenticated human GitHub login in the
// background and stores it on the telemetry identity, so every exported daemon
// event carries github_actor, not just ao.session.spawned. The provider is
// rebuilt per attempt because it reads the token at construction. Only the
// authenticated GitHub account is used, never git config.
func startGitHubLoginResolver(ctx context.Context, cfg config.Config, ident ports.TelemetryIdentityStore, log *slog.Logger) {
	if ident == nil || !cfg.Telemetry.Events || cfg.Telemetry.Remote != config.TelemetryRemotePostHog {
		return
	}
	go func() {
		for _, delay := range gitHubLoginRetryDelays {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
			scm := newMultiSCMProvider(cfg.GitLab, log)
			if scm == nil {
				continue
			}
			attempt, cancel := context.WithTimeout(ctx, 15*time.Second)
			identity, err := scm.AuthenticatedIdentityForProvider(attempt, "github", "")
			cancel()
			if err == nil && identity.Human && identity.Login != "" {
				ident.SetGitHubLogin(identity.Login)
				return
			}
		}
	}()
}
