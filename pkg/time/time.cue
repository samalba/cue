package time

#Parts: {
	Year:       int64
	Month:      >=1 & <=12
	Day:        >=0 & <=6
	Hour:       >=0 & <=23
	Minute:     >=0 & <=59
	Second:     Nanosecond / 1_000_000_000
	Nanosecond: int
}
