// Package contract is the Gap A activation checklist protocol. Pure data + interfaces.
package contract

const (
	StepWait    = "wait"
	StepHealthz = "healthz"
	StepPing    = "ping"
	StepIngest  = "ingest"
	StepVerify  = "verify"
	StepSmoke   = "smoke"
	StepGold    = "gold"
	StepEgress  = "egress"
	StepCronJob = "cronjob"
)

// Endpoints are the local minikube NodePorts from the README.
type Endpoints struct {
	API    string
	Bolt   string
	Qdrant string
}

func NodePorts(host string) Endpoints {
	if host == "" {
		host = "127.0.0.1"
	}
	return Endpoints{
		API:    "http://" + host + ":30080",
		Bolt:   "bolt://" + host + ":30687",
		Qdrant: "http://" + host + ":30333",
	}
}

type Step struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

type Report struct {
	Status string `json:"status"`
	Steps  []Step `json:"steps"`
}

func (r Report) AllOK() bool {
	if len(r.Steps) == 0 {
		return false
	}
	for _, step := range r.Steps {
		if !step.OK {
			return false
		}
	}
	return true
}

// Checker is the HTTP half of Gap A. Impl is a child; tests stub this.
type Checker interface {
	Health(api string) Step
	Smoke(api string) Report
	Gold(api string) Report
}
