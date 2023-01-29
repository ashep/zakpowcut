package printer

import (
	"fmt"
	"time"

	"zakpowcut/parser"
)

var (
	sun  = []byte{226, 152, 128, 239, 184, 143}
	moon = []byte{240, 159, 140, 145}
)

func PrintTimeTable(tt parser.TimeTable) string {
	var s []byte

	s = fmt.Append(s, "  ")
	for hn := 0; hn < 24; hn++ {
		s = fmt.Appendf(s, " %02d", hn)
	}
	s = fmt.Appendln(s)

	for qn := 0; qn < 4; qn++ {
		s = fmt.Appendf(s, "%d: ", qn+1)
		for hn := 0; hn < 24; hn++ {
			if tt[qn][hn] {
				s = fmt.Append(s, " * ")
			} else {
				s = fmt.Append(s, " - ")
			}
		}
		s = fmt.Appendln(s)
	}

	return string(s)
}

func PrintTimeRanges(trs parser.TimeRanges, surround string) string {
	var s []byte

	for _, tr := range trs {
		v := moon
		if tr.On {
			v = sun
		}

		s = fmt.Appendf(s, "%s %s%02d-%02d%s\n", v, surround, tr.Start, tr.End, surround)
	}

	return string(s)
}

func PrintAllTimeRanges(trs [4]parser.TimeRanges, dt time.Time, qTitle, surround string) string {
	var s []byte

	for qn := 0; qn < 4; qn++ {
		if qTitle != "" {
			s = fmt.Appendf(s, "%s %d\n", qTitle, qn+1)
		}

		if dt.Year() > 2022 {
			s = fmt.Appendf(s, "**Графік на %s**\n\n", dt.Format("02.01.2006"))
		}

		s = fmt.Appendf(s, "%s", PrintTimeRanges(trs[qn], surround))
		s = fmt.Appendln(s)
	}

	return string(s)
}
