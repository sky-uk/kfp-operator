package base

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/common"
	"regexp"
	"strings"
)

var FieldMatcher = regexp.MustCompile(`\S+`)

const OutputSeparator = " "
const StdNumFields = 5
const GoNumFields = 6

type CronSchedule struct {
	fields []string
}

func (cs CronSchedule) PrintStandard() string {
	return strings.Join(cs.fields[1:], OutputSeparator)
}

func (cs CronSchedule) PrintGo() string {
	return strings.Join(cs.fields, OutputSeparator)
}

func ParseCron(schedule string) (CronSchedule, error) {
	fields := FieldMatcher.FindAllString(schedule, -1)

	if len(fields) > GoNumFields {
		return CronSchedule{}, fmt.Errorf("too many fields for go cron schedule")
	}
	if len(fields) < StdNumFields {
		return CronSchedule{}, fmt.Errorf("too few fields for standard cron schedule")
	}

	if len(fields) == StdNumFields {
		return CronSchedule{
			fields: append([]string{"0"}, fields...),
		}, nil
	} else {
		return CronSchedule{
			fields: fields,
		}, nil
	}
}

func SanitiseNamespacedName(namespacedName common.NamespacedName) (string, error) {
	return namespacedName.SeparatedString("-")
}
