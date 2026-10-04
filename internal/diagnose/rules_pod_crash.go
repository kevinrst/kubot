package diagnose

import (
	"fmt"

	"kubot/internal/k8s"
	"kubot/internal/model"
)

type PodCrashLoopRule struct{}

func (PodCrashLoopRule) Name() string { return "pod_crashloop_backoff" }

func (PodCrashLoopRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		for _, cs := range pod.Status.ContainerStatuses {
			waiting := ""
			waitMsg := ""
			if cs.State.Waiting != nil {
				waiting = cs.State.Waiting.Reason
				waitMsg = cs.State.Waiting.Message
			}
			lastExit := -1
			lastReason := ""
			if cs.LastTerminationState.Terminated != nil {
				lastExit = int(cs.LastTerminationState.Terminated.ExitCode)
				lastReason = cs.LastTerminationState.Terminated.Reason
			}
			crashloop := waiting == "CrashLoopBackOff"
			// Between backoffs the container sits in Error/Terminated with a
			// non-zero exit. Same loop, different snapshot instant; 3+ restarts
			// keeps one-off failures silent.
			curExit := -1
			curReason := ""
			if cs.State.Terminated != nil {
				curExit = int(cs.State.Terminated.ExitCode)
				curReason = cs.State.Terminated.Reason
				if lastExit == -1 {
					lastExit, lastReason = curExit, cs.State.Terminated.Reason
				}
			}
			// OOMKills loop too, but that rule owns the signal — don't report twice.
			if lastReason == "OOMKilled" || curReason == "OOMKilled" {
				continue
			}
			repeated := !crashloop && cs.RestartCount >= 3 && (lastExit != 0 || curExit != 0) &&
				waiting != "ContainerCreating" && waiting != "PodInitializing"
			if !crashloop && !repeated {
				continue
			}
			msg := fmt.Sprintf("Container %q is in CrashLoopBackOff", cs.Name)
			if repeated {
				msg = fmt.Sprintf("Container %q keeps crashing (%d restarts, last exit %d)", cs.Name, cs.RestartCount, lastExit)
			}
			out = append(out, model.Finding{
				Severity:  "critical",
				Resource:  fmt.Sprintf("pod/%s", pod.Name),
				Namespace: pod.Namespace,
				Reason:    "pod_crashloop_backoff",
				Message:   msg,
				Evidence: map[string]any{
					"container":       cs.Name,
					"restart_count":   cs.RestartCount,
					"last_exit_code":  lastExit,
					"last_reason":     lastReason,
					"waiting_reason":  waiting,
					"waiting_message": waitMsg,
				},
				Recommendation: "Check container logs and startup command; verify probes, env, and mounted config.",
			})
		}
	}
	return out
}

type PodOOMKilledRule struct{}

func (PodOOMKilledRule) Name() string { return "pod_oom_killed" }

func (PodOOMKilledRule) Run(s *k8s.Snapshot) []model.Finding {
	var out []model.Finding
	for _, pod := range s.Pods {
		for _, cs := range pod.Status.ContainerStatuses {
			term := oomTerm(cs)
			if term == nil {
				continue
			}
			var limit string
			for _, c := range pod.Spec.Containers {
				if c.Name == cs.Name {
					if q, ok := c.Resources.Limits["memory"]; ok {
						limit = q.String()
					}
				}
			}
			ev := map[string]any{
				"container":     cs.Name,
				"restart_count": cs.RestartCount,
				"exit_code":     term.ExitCode,
				"reason":        "OOMKilled",
			}
			if limit != "" {
				ev["memory_limit"] = limit
			} else {
				ev["memory_limit"] = "unset"
			}
			out = append(out, model.Finding{
				Severity:  "critical",
				Resource:  fmt.Sprintf("pod/%s", pod.Name),
				Namespace: pod.Namespace,
				Reason:    "pod_oom_killed",
				Message:   fmt.Sprintf("Container %q was OOMKilled (exit 137)", cs.Name),
				Evidence:  ev,
				Recommendation: "Increase the container memory limit or investigate memory usage. " +
					"If no limit is set, the node may be under memory pressure.",
			})
		}
	}
	return out
}
