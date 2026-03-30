package pipeline

import (
	"fmt"
	"os"
	"time"

	"github.com/chandrasekar-r/bitbucket-cli/pkg/api"
	"github.com/chandrasekar-r/bitbucket-cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// exitCodeSuccess is returned when the pipeline completed successfully.
const exitCodeSuccess = 0

// exitCodePipelineFailed is returned when the pipeline itself failed/errored/stopped.
// This lets `set -e` CI scripts detect a failed pipeline run.
const exitCodePipelineFailed = 1

// exitCodeWatchError is returned when the watch command itself encountered an error
// (network failure, rate limit, etc.) — distinct from a pipeline that ran and failed.
const exitCodeWatchError = 2

func newCmdWatch(f *cmdutil.Factory) *cobra.Command {
	var pollInterval int
	var timeout int

	cmd := &cobra.Command{
		Use:   "watch [uuid]",
		Short: "Follow pipeline logs in real time (polling)",
		Long: `Watch a pipeline run and stream its log output.

If no UUID is given, triggers a new pipeline on the current branch and watches it.

Exit codes:
  0  Pipeline completed successfully
  1  Pipeline failed, errored, or was stopped
  2  Watch itself failed (network error, timeout, etc.)

Note: Bitbucket Pipelines has no streaming API. Logs are fetched via polling
with HTTP byte-range requests at the configured interval.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Enforce minimum poll interval
			if pollInterval < 5 {
				fmt.Fprintf(f.IOStreams.ErrOut,
					"Warning: --poll-interval below 5s may hit rate limits; using 5s\n")
				pollInterval = 5
			}

			workspace, slug, err := repoCtx(f)
			if err != nil {
				return err
			}
			httpClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			client := api.New(httpClient, f.BaseURL)

			var pipelineUUID string

			if len(args) == 1 {
				pipelineUUID = args[0]
				fmt.Fprintf(f.IOStreams.Out, "Watching pipeline %s...\n\n", pipelineUUID)
			} else {
				// No UUID — trigger on current branch and watch
				branch := currentBranch()
				if branch == "" {
					return &cmdutil.FlagError{Err: fmt.Errorf("cannot detect current branch: use `bb pipeline watch <uuid>` instead")}
				}
				fmt.Fprintf(f.IOStreams.Out, "Triggering pipeline on branch %s...\n", branch)
				p, err := client.TriggerPipeline(workspace, slug, api.TriggerPipelineOptions{Branch: branch})
				if err != nil {
					return fmt.Errorf("triggering pipeline: %w", err)
				}
				pipelineUUID = p.UUID
				fmt.Fprintf(f.IOStreams.Out, "Pipeline #%d started (%s)\n\n", p.BuildNumber, pipelineUUID)
			}

			exitCode := watchPipeline(f, client, workspace, slug, pipelineUUID, pollInterval, timeout)
			os.Exit(exitCode)
			return nil // unreachable
		},
	}

	cmd.Flags().IntVar(&pollInterval, "poll-interval", 10,
		"Seconds between status polls (minimum 5)")
	cmd.Flags().IntVar(&timeout, "timeout", 0,
		"Abort watch after N minutes (0 = no timeout)")
	return cmd
}

// watchPipeline polls the pipeline until completion and prints log output incrementally.
// Returns the exit code: 0 = success, 1 = pipeline failed, 2 = watch error.
func watchPipeline(
	f *cmdutil.Factory,
	client *api.Client,
	workspace, slug, uuid string,
	pollIntervalSecs, timeoutMins int,
) int {
	ticker := time.NewTicker(time.Duration(pollIntervalSecs) * time.Second)
	defer ticker.Stop()

	var deadline <-chan time.Time
	if timeoutMins > 0 {
		t := time.NewTimer(time.Duration(timeoutMins) * time.Minute)
		defer t.Stop()
		deadline = t.C
	}

	// Track byte offset per step UUID to fetch only new log bytes each poll
	stepOffsets := map[string]int64{}
	requestCount := 0

	for {
		select {
		case <-deadline:
			fmt.Fprintf(f.IOStreams.ErrOut, "\nWatch timed out after %d minutes\n", timeoutMins)
			return exitCodeWatchError

		case <-ticker.C:
			// 1. Fetch overall pipeline state
			requestCount++
			p, err := client.GetPipeline(workspace, slug, uuid)
			if err != nil {
				fmt.Fprintf(f.IOStreams.ErrOut, "\nError polling pipeline: %v\n", err)
				return exitCodeWatchError
			}

			// 2. Fetch steps and print any new log bytes
			requestCount++
			steps, err := client.ListSteps(workspace, slug, uuid)
			if err != nil {
				fmt.Fprintf(f.IOStreams.ErrOut, "\nError fetching steps: %v\n", err)
				return exitCodeWatchError
			}

			for _, step := range steps {
				if step.State.Name == "PENDING" {
					continue // step hasn't started yet
				}
				offset := stepOffsets[step.UUID]
				requestCount++
				data, err := client.GetStepLog(workspace, slug, uuid, step.UUID, offset)
				if err != nil {
					// Non-fatal: log fetch may fail transiently during step startup
					continue
				}
				if len(data) > 0 {
					fmt.Fprint(f.IOStreams.Out, string(data))
					stepOffsets[step.UUID] = offset + int64(len(data))
				}
			}

			// Warn when approaching rate limit (1,000 req/hr ≈ 278 reqs per 10-min window)
			if requestCount > 0 && requestCount%200 == 0 {
				fmt.Fprintf(f.IOStreams.ErrOut,
					"\nNote: %d API requests made this session. Bitbucket rate limit is 1,000/hr.\n",
					requestCount)
			}

			// 3. Check if pipeline has finished
			if p.IsComplete() {
				result := p.ResultName()
				fmt.Fprintf(f.IOStreams.ErrOut, "\nPipeline #%d %s (%s)\n",
					p.BuildNumber,
					statusColor(p.State.Name, result, f.IOStreams.ColorEnabled()),
					formatDuration(p.DurationInSeconds),
				)
				if result == "SUCCESSFUL" {
					return exitCodeSuccess
				}
				return exitCodePipelineFailed
			}
		}
	}
}
