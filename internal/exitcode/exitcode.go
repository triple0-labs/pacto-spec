package exitcode

import "pacto/internal/model"

func Evaluate(failOn string, report model.StatusReport) int {
	switch failOn {
	case "none", "":
		return 0
	case "unverified":
		for _, p := range report.Plans {
			if p.Verification == "unverified" {
				return 1
			}
		}
	case "partial":
		for _, p := range report.Plans {
			if p.Verification == "partial" || p.Verification == "unverified" {
				return 1
			}
		}
	case "blocked":
		for _, p := range report.Plans {
			if p.BlockedTasks > 0 {
				return 1
			}
		}
	}
	return 0
}
