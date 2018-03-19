package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v1"
)

type CondorJobSubmissionBuilder struct {
	cfg *viper.Viper
}

func (b CondorJobSubmissionBuilder) build(submission *model.Job, dirPath string) (string, error) {

}

func newCondorJobSubmissionBuilder(cfg *viper.Viper) JobSubmissionBuilder {
	return CondorJobSubmissionBuilder{cfg: cfg}
}
