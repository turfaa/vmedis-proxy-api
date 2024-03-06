package sale

func GenerateDailyHistory(stats []Statistics) []Statistics {
	res := make([]Statistics, 0, len(stats))
	for _, stat := range stats {
		if len(res) == 0 || !IsStatsInSameDay(res[len(res)-1], stat) {
			res = append(res, stat)
		} else {
			res[len(res)-1] = stat
		}
	}

	return res
}

func IsStatsInSameDay(stat1, stat2 Statistics) bool {
	return stat1.PulledAt.Year() == stat2.PulledAt.Year() &&
		stat1.PulledAt.Month() == stat2.PulledAt.Month() &&
		stat1.PulledAt.Day() == stat2.PulledAt.Day()
}
