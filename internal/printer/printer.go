package printer

import (
	"fmt"

	"zakpowcut/internal/parser"
)

var (
	sun      = []byte{226, 152, 128, 239, 184, 143}
	moon     = []byte{240, 159, 140, 145}
	question = []byte{226, 157, 147}
)

func PrintTimeTable(tt parser.TimeTable) string {
	var s string

	s = "  "
	for hn := 0; hn < 24; hn++ {
		s += fmt.Sprintf(" %02d", hn)
	}
	s += fmt.Sprintf("\n")

	for qn := 0; qn < len(tt); qn++ {
		s += fmt.Sprintf("%d: ", qn+1)
		for hn := 0; hn < 24; hn++ {
			switch tt[qn][hn] {
			case parser.PowerOn:
				s += fmt.Sprintf(" * ")
			case parser.PowerPerhaps:
				s += fmt.Sprintf(" ? ")
			default:
				s += fmt.Sprintf(" - ")
			}
		}
		s += fmt.Sprintf("\n")
	}

	return s
}

func PrintTimeRanges(trs parser.TimeRanges) string {
	var (
		v []byte
		s string
	)

	for _, tr := range trs {
		switch tr.State {
		case parser.PowerOn:
			v = sun
		case parser.PowerPerhaps:
			v = question
		default:
			v = moon
		}

		diff := tr.End - tr.Start + 1
		suffix := "годин"
		switch {
		case diff == 1 || diff == 21:
			suffix += "а"
		case diff > 1 && diff < 5 || diff > 21 && diff < 25:
			suffix += "и"
		}

		s += fmt.Sprintf("%s `%02d-%02d` — %d %s\n", v, tr.Start, tr.End, diff, suffix)
	}

	return s
}
