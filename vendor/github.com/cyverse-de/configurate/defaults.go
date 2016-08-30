package configurate

// JobServicesDefaults contains the default configuration for the DE job execution services.
const JobServicesDefaults = `
amqp:
  uri: amqp://guest:guest@rabbit:5672/jobs

apps:
  callbacks_uri: "http://apps:60000/callbacks/de-job"

condor:
  condor_config: /etc/condor/condor_config
  path_env_var: /opt/condor/bin/:/usr/bin/:/usr/local/bin/:/bin/
  log_path: /tmp/
  request_disk: 0
  filter_files: ".job.ad,.machine.ad,_condor_stderr,_condor_stdout,condor_exec.exe,.chirp_config,\
.chirp.config,logs/logs-stderr-output,logs/logs-stdout-output,config,job,iplant.cmd"

db:
  uri: postgresql://guest:guest@dedb:5432/de?sslmode=disable

irods:
  user: "rods"
  pass: "notprod"
  host: "irods"
  port: "1247"
  base: "/iplant/home"
  resc: ""
  zone: "iplant"

porklock:
  image: discoenv/porklock
  tag: "dev"
`
