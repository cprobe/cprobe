package probe

func makeJobs() map[string]map[JobID]*JobGoroutine {
	return map[string]map[JobID]*JobGoroutine{
		"mysql":         make(map[JobID]*JobGoroutine),
		"redis":         make(map[JobID]*JobGoroutine),
		"elasticsearch": make(map[JobID]*JobGoroutine),
		"postgresql":    make(map[JobID]*JobGoroutine),
		"kafka":         make(map[JobID]*JobGoroutine),
		"mongodb":       make(map[JobID]*JobGoroutine),
	}
}
