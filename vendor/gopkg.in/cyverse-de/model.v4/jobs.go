package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/cyverse-de/model/submitfile"
	"github.com/spf13/viper"
)

var (
	validName = regexp.MustCompile(`-\d{4}(?:-\d{2}){5}\.\d+$`) // this isn't included in the Dirname() function so it isn't re-evaluated a lot
	quoteStr  = regexp.MustCompile(`^''|''$`)
)

const (
	nowfmt = "2006-01-02-15-04-05.000" // appears in file and directory names.

	//DockerLabelKey is the key for the labels applied to all containers associated with a job.
	DockerLabelKey = "org.iplantc.analysis"
)

// naivelyquote single-quotes a string that will be placed on the command line
// using plain string substitution.  This works, but may leave extra pairs
// of leading or trailing quotes if there was a leading or trailing quote
// in the original string, which is valid, but may be confusing to human
// readers.
func naivelyquote(s string) string {
	return fmt.Sprintf("'%s'", strings.Replace(s, "'", "''", -1))
}

// quote quotes and escapes a string that is supposed to be passed in to a tool on
// the command line.
func quote(s string) string {
	return quoteStr.ReplaceAllString(naivelyquote(s), "")
}

// ExtractJobID pulls the job id from the given []byte, if it exists. Returns
// an empty []byte if it doesn't.
func ExtractJobID(output []byte) []byte {
	extractor := regexp.MustCompile(`submitted to cluster ((\d+)+)`)
	matches := extractor.FindAllSubmatch(output, -1)
	var thematch []byte
	if len(matches) > 0 {
		if len(matches[0]) > 1 {
			thematch = matches[0][1]
		}
	}
	return thematch
}

// Job is a type that contains info that goes into the jobs table.
type Job struct {
	AppDescription     string         `json:"app_description"`
	AppID              string         `json:"app_id"`
	AppName            string         `json:"app_name"`
	ArchiveLogs        bool           `json:"archive_logs"`
	ID                 string         `json:"id"`
	BatchID            string         `json:"batch_id"`
	CondorID           string         `json:"condor_id"`
	CondorLogPath      string         `json:"condor_log_path"` //comes from config, not upstream service
	CreateOutputSubdir bool           `json:"create_output_subdir"`
	DateSubmitted      time.Time      `json:"date_submitted"`
	DateStarted        time.Time      `json:"date_started"`
	DateCompleted      time.Time      `json:"date_completed"`
	Description        string         `json:"description"`
	Email              string         `json:"email"`
	ExecutionTarget    string         `json:"execution_target"`
	ExitCode           int            `json:"exit_code"`
	FailureCount       int64          `json:"failure_count"`
	FailureThreshold   int64          `json:"failure_threshold"`
	FileMetadata       []FileMetadata `json:"file-metadata"`
	FilterFiles        []string       `json:"filter_files"`      //comes from config, not upstream service
	Group              string         `json:"group"`             //untested for now
	InputPathListFile  string         `json:"input_path_list"`   //path to a list of inputs (not from upstream).
	InputTicketsFile   string         `json:"input_ticket_list"` //path to a list of inputs with tickets (not from upstream).
	InvocationID       string         `json:"uuid"`
	IRODSBase          string         `json:"irods_base"`
	Name               string         `json:"name"`
	NFSBase            string         `json:"nfs_base"`
	Notify             bool           `json:"notify"`
	NowDate            string         `json:"now_date"`
	OutputDir          string         `json:"output_dir"`         //the value parsed out of the JSON. Use OutputDirectory() instead.
	OutputDirTicket    string         `json:"output_dir_ticket"`  //the write ticket for output_dir (assumes output_dir is set correctly).
	OutputTicketFile   string         `json:"output_ticket_list"` //path to the file of the output dest with ticket (not from upstream).
	RequestType        string         `json:"request_type"`
	RunOnNFS           bool           `json:"run-on-nfs"`
	SkipParentMetadata bool           `json:"skip-parent-meta"`
	Steps              []Step         `json:"steps"`
	SubmissionDate     string         `json:"submission_date"`
	Submitter          string         `json:"username"`
	Type               string         `json:"type"`
	UserID             string         `json:"user_id"`
	UserGroups         []string       `json:"user_groups"`
	WikiURL            string         `json:"wiki_url"`
	ConfigFile         string         `json:"config_file"` //path to the job configuration file (not from upstream)
}

// New returns a pointer to a newly instantiated Job with NowDate set.
// Accesses the following configuration settings:
//  * condor.log_path
//  * condor.filter_files
//  * irods.base
func New(cfg *viper.Viper) *Job {
	n := time.Now().Format(nowfmt)
	lp := cfg.GetString("condor.log_path")
	var paths []string
	filterFiles := cfg.GetString("condor.filter_files")
	for _, filter := range strings.Split(filterFiles, ",") {
		paths = append(paths, filter)
	}
	irodsBase := cfg.GetString("irods.base")
	return &Job{
		NowDate:        n,
		SubmissionDate: n,
		ArchiveLogs:    true,
		CondorLogPath:  lp,
		FilterFiles:    paths,
		IRODSBase:      irodsBase,
	}
}

// NewFromData creates a new submission and populates it by parsing the passed
// in []byte as JSON.
func NewFromData(cfg *viper.Viper, data []byte) (*Job, error) {
	var err error
	s := New(cfg)
	err = json.Unmarshal(data, s)
	if err != nil {
		return nil, err
	}
	s.Sanitize()
	s.AddRequiredMetadata()
	return s, err
}

// sanitize replaces @ and spaces with _, making a string safe to use as a
// part of a path. Mostly to keep things from getting really confusing when
// a path is passed to Condor.
func sanitize(s string) string {
	step := strings.Replace(s, "@", "_", -1)
	step = strings.Replace(step, " ", "_", -1)
	return step
}

// Sanitize makes sure the fields in a submission are ready to be used in things
// like file names.
func (job *Job) Sanitize() {
	job.Submitter = sanitize(job.Submitter)

	if job.Type == "" {
		job.Type = "analysis"
	}

	job.Name = sanitize(job.Name)

	for i, step := range job.Steps {
		step.Component.Container.Image.Name = strings.TrimSpace(step.Component.Container.Image.Name)
		step.Component.Container.Image.Tag = strings.TrimSpace(step.Component.Container.Image.Tag)
		step.Component.Container.Image.OSGImagePath = strings.TrimSpace(step.Component.Container.Image.OSGImagePath)
		step.Component.Container.Name = strings.TrimSpace(step.Component.Container.Name)

		for j, vf := range step.Component.Container.VolumesFrom {
			vf.Name = strings.TrimSpace(vf.Name)
			vf.Tag = strings.TrimSpace(vf.Tag)
			vf.NamePrefix = strings.TrimSpace(vf.NamePrefix)
			vf.HostPath = strings.TrimSpace(vf.HostPath)
			vf.ContainerPath = strings.TrimSpace(vf.ContainerPath)
			step.Component.Container.VolumesFrom[j] = vf
		}
		job.Steps[i] = step
	}
}

// DirectoryName creates a directory name for an analysis. Used when the submission
// doesn't specify an output directory.  Some types of jobs, for example
// Foundational API jobs, include a timestamp in the job name, so a timestamp
// will not be appended to the directory name in those cases.
func (job *Job) DirectoryName() string {
	if validName.MatchString(job.Name) {
		return job.Name
	}
	return fmt.Sprintf("%s-%s", job.Name, job.NowDate)
}

// UserIDForSubmission returns the cleaned up user ID for use in the iplant.cmd file. This
// is dumb. Very, very dumb.
func (job *Job) UserIDForSubmission() string {
	var retval string
	if job.UserID == "" {
		hash := sha256.New()
		hash.Write([]byte(job.Submitter))
		md := hash.Sum(nil)
		retval = hex.EncodeToString(md)
	} else {
		retval = job.UserID
	}
	return fmt.Sprintf("_%s", strings.Replace(retval, "-", "", -1))
}

// CondorLogDirectory returns the path to the directory containing condor logs on the
// submission node. This a computed value, so it isn't in the struct.
func (job *Job) CondorLogDirectory() string {
	return fmt.Sprintf("%s/", path.Join(job.CondorLogPath, job.Submitter, job.DirectoryName()))
}

// IRODSConfig returns the path to iRODS config inside the working directory.
func (job *Job) IRODSConfig() string {
	return path.Join("logs", "irods-config")
}

// OutputDirectory returns the path to the output directory in iRODS. It's
// computed, which is why it isn't in the struct. Use this instead of directly
// accessing the OutputDir field.
func (job *Job) OutputDirectory() string {
	if job.OutputDir == "" {
		return path.Join(job.IRODSBase, job.Submitter, "analyses", job.DirectoryName())
	} else if job.OutputDir != "" && job.CreateOutputSubdir {
		return path.Join(job.OutputDir, job.DirectoryName())
	} else if job.OutputDir != "" && !job.CreateOutputSubdir {
		return strings.TrimSuffix(job.OutputDir, "/")
	}
	//probably won't ever reach this, but just in case...
	return path.Join(job.IRODSBase, job.Submitter, "analyses", job.DirectoryName())
}

// DataContainers returns a list of VolumesFrom that describe the data
// containers associated with the job submission.
func (job *Job) DataContainers() []VolumesFrom {
	var vfs []VolumesFrom
	for _, step := range job.Steps {
		for _, vf := range step.Component.Container.VolumesFrom {
			vfs = append(vfs, vf)
		}
	}
	return vfs
}

// ContainerImages returns a []ContainerImage of all of the images associated
// with this submission.
func (job *Job) ContainerImages() []ContainerImage {
	var ci []ContainerImage
	for _, step := range job.Steps {
		ci = append(ci, step.Component.Container.Image)
	}
	return ci
}

// Inputs returns all of the StepInputs associated with the submission,
// regardless of what step they're associated with.
func (job *Job) Inputs() []StepInput {
	var inputs []StepInput
	for _, step := range job.Steps {
		for _, input := range step.Config.Inputs {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// Outputs returns all of the StepOutputs associated with the submission,
// regardless of what step they're associated with.
func (job *Job) Outputs() []StepOutput {
	var outputs []StepOutput
	for _, step := range job.Steps {
		for _, output := range step.Config.Outputs {
			outputs = append(outputs, output)
		}
	}
	return outputs
}

// ExcludeArguments returns a list of paths that should not upload as outputs.
func (job *Job) ExcludeArguments() []string {
	var paths []string
	for _, input := range job.Inputs() {
		if !input.Retain && input.Value != "" {
			paths = append(paths, input.Source())
		}
	}
	for _, output := range job.Outputs() {
		if !output.Retain {
			paths = append(paths, output.Source())
		}
	}
	for _, ff := range job.FilterFiles {
		paths = append(paths, ff)
	}
	if !job.ArchiveLogs {
		paths = append(paths, "logs")
	}

	return paths
}

// AddRequiredMetadata adds any required AVUs that are required but are missing
// from Job.FileMetadata. This should be called after both of the New*()
// functions and after the Job has been initialized from JSON.
func (job *Job) AddRequiredMetadata() {
	foundAnalysis := false
	foundExecution := false
	for _, md := range job.FileMetadata {
		if md.Attribute == "ipc-analysis-id" {
			foundAnalysis = true
		}
		if md.Attribute == "ipc-execution-id" {
			foundExecution = true
		}
	}
	if !foundAnalysis {
		job.FileMetadata = append(
			job.FileMetadata,
			FileMetadata{
				Attribute: "ipc-analysis-id",
				Value:     job.AppID,
				Unit:      "UUID",
			},
		)
	}
	if !foundExecution {
		job.FileMetadata = append(
			job.FileMetadata,
			FileMetadata{
				Attribute: "ipc-execution-id",
				Value:     job.InvocationID,
				Unit:      "UUID",
			},
		)
	}
}

// FinalOutputArguments returns a string containing the arguments passed to
// porklock for the final output operation, which transfers all files back into
// iRODS.
func (job *Job) FinalOutputArguments(excludeFilePath string) []string {
	dest := job.OutputDirectory()
	retval := []string{
		"put",
		"--user", job.Submitter,
		"--destination", dest,
	}
	for _, m := range MetadataArgs(job.FileMetadata).FileMetadataArguments() {
		retval = append(retval, m)
	}
	if excludeFilePath != "" {
		retval = append(retval, "--exclude", excludeFilePath)
	}
	if job.SkipParentMetadata {
		retval = append(retval, "--skip-parent-meta")
	}
	return retval
}

// FormatUserGroups converts the list of user groups to the list format used by the
// HTCondor job submission file.
func (job *Job) FormatUserGroups() string {
	return submitfile.FormatList(job.UserGroups)
}

// UsesVolumes returns a boolean value which indicates if any step of a job uses host-mounted volumes
func (job *Job) UsesVolumes() bool {
	for _, step := range job.Steps {
		if step.UsesVolumes() {
			return true
		}
	}
	return false
}

// FilterInputsWithoutTickets returns a list of inputs that do not have download tickets.
func (job *Job) FilterInputsWithoutTickets() []StepInput {
	var inputs []StepInput
	for _, input := range job.Inputs() {
		if input.Ticket == "" {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// FilterInputsWithTickets returns a list of inputs that have download tickets.
func (job *Job) FilterInputsWithTickets() []StepInput {
	var inputs []StepInput
	for _, input := range job.Inputs() {
		if input.Ticket != "" {
			inputs = append(inputs, input)
		}
	}
	return inputs
}

// CPURequest calculates the highest maximum CPU among the steps of a job (i.e.
// the largest slot size the job will need), or 0 if no steps have maximum CPUs
// set
func (job *Job) CPURequest() float32 {
	var cpu float32

	for _, step := range job.Steps {
		if step.Component.Container.MaxCPUCores > cpu {
			cpu = step.Component.Container.MaxCPUCores
		}
	}

	return cpu
}

// MemoryRequest calculates the highest maximum memory among the steps of a job
// (i.e.  the largest slot size the job will need), or 0 if no steps have
// maximum memory set
func (job *Job) MemoryRequest() int64 {
	var mem int64

	for _, step := range job.Steps {
		if step.Component.Container.MemoryLimit > mem {
			mem = step.Component.Container.MemoryLimit
		}
	}

	return mem
}

// DiskRequest calculates the highest disk need among the steps of a job
// (i.e.  the largest slot size the job will need), or 0 if no steps have
// disk set. As we only track minimum disk space, it uses that number.
func (job *Job) DiskRequest() int64 {
	var disk int64

	for _, step := range job.Steps {
		if step.Component.Container.MinDiskSpace > disk {
			disk = step.Component.Container.MinDiskSpace
		}
	}

	return disk
}

// FileMetadata describes a unit of metadata that should get associated with
// all of the files associated with the job submission.
type FileMetadata struct {
	Attribute string `json:"attr"`
	Value     string `json:"value"`
	Unit      string `json:"unit"`
}

// Argument returns a string containing the command-line settings for the
// file transfer tool.
func (m *FileMetadata) Argument() []string {
	return []string{"-m", fmt.Sprintf("%s,%s,%s", m.Attribute, m.Value, m.Unit)}
}

// MetadataArgs is a list of FileMetadata
type MetadataArgs []FileMetadata

// FileMetadataArguments returns a string containing the command-line arguments
// for porklock that sets all of the metadata triples.
func (m MetadataArgs) FileMetadataArguments() []string {
	retval := []string{}
	for _, fm := range m {
		for _, a := range fm.Argument() {
			retval = append(retval, a)
		}
	}
	return retval
}
