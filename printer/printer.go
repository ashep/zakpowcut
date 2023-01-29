package printer

import (
	"fmt"

	"zakpowcut/parser"
)

var (
	sun  = []byte{226, 152, 128, 239, 184, 143}
	moon = []byte{240, 159, 140, 145}
)

func PrintTimeTable(tt parser.TimeTable) string {
	var s string

	s = "  "
	for hn := 0; hn < 24; hn++ {
		s += fmt.Sprintf(" %02d", hn)
	}
	s += fmt.Sprintf("\n")

	for qn := 0; qn < 4; qn++ {
		s += fmt.Sprintf("%d: ", qn+1)
		for hn := 0; hn < 24; hn++ {
			if tt[qn][hn] {
				s += fmt.Sprintf(" * ")
			} else {
				s += fmt.Sprintf(" - ")
			}
		}
		s += fmt.Sprintf("\n")
	}

	return s
}

func PrintTimeRanges(trs parser.TimeRanges, surround string) string {
	var s string

	for _, tr := range trs {
		v := moon
		if tr.On {
			v = sun
		}

		s += fmt.Sprintf("%s %s%02d-%02d%s\n", v, surround, tr.Start, tr.End, surround)
	}

	return s
}
