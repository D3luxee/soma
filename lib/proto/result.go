/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto

const (
	StatusOK             = 200
	StatusAccepted       = 202
	StatusPartial        = 206
	StatusBadRequest     = 400
	StatusUnauthorized   = 401
	StatusForbidden      = 403
	StatusNotFound       = 404
	StatusConflict       = 406
	StatusError          = 500
	StatusNotImplemented = 501
	StatusUnavailable    = 503
	StatusGatewayTimeout = 504
)

// Display text for status codes
var DisplayStatus = map[int]string{
	200: "OK",
	202: "Accepted",
	206: "Partial result",
	400: "Bad Request",
	401: "Unauthorized",
	403: "Forbidden",
	404: "Not found",
	406: "Readonly instance",
	500: "Server error",
	501: "Not implemented",
	503: "Service unavailable",
	504: "Gateway timeout",
}

type Result struct {
	StatusCode uint16 `json:"statusCode"`
	StatusText string `json:"statusText"`
	RequestID  string `json:"requestID"`

	// Errors is set for StatusCode >399
	Errors *[]string `json:"errors,omitempty"`
	// JobID is set for StatusCode 202 (async processing)
	JobID   string `json:"jobId,omitempty"`
	JobType string `json:"jobType,omitempty"`
	// List of (outstanding) deployment IDs
	DeploymentsList *[]string `json:"deploymentsList,omitempty"`

	// Request dependent data
	Actions          *[]Action          `json:"actions,omitempty"`
	Admins           *[]Admin           `json:"admins,omitempty"`
	Attributes       *[]Attribute       `json:"attributes,omitempty"`
	Buckets          *[]Bucket          `json:"buckets,omitempty"`
	Capabilities     *[]Capability      `json:"capability,omitempty"`
	Categories       *[]Category        `json:"categories,omitempty"`
	CheckConfigs     *[]CheckConfig     `json:"checkConfigs,omitempty"`
	Clusters         *[]Cluster         `json:"clusters,omitempty"`
	DatacenterGroups *[]DatacenterGroup `json:"datacenterGroups,omitempty"`
	Datacenters      *[]Datacenter      `json:"datacenter,omitempty"`
	Deployments      *[]Deployment      `json:"deployments,omitempty"`
	Entities         *[]Entity          `json:"entities,omitempty"`
	Environments     *[]Environment     `json:"environment,omitempty"`
	Grants           *[]Grant           `json:"grants,omitempty"`
	Groups           *[]Group           `json:"groups,omitempty"`
	HostDeployments  *[]HostDeployment  `json:"hostDeployments,omitempty"`
	Instances        *[]Instance        `json:"instances,omitempty"`
	JobResults       *[]JobResult       `json:"jobResults,omitempty"`
	JobStatus        *[]JobStatus       `json:"jobStatus,omitempty"`
	JobTypes         *[]JobType         `json:"jobTypes,omitempty"`
	Jobs             *[]Job             `json:"jobs,omitempty"`
	Levels           *[]Level           `json:"levels,omitempty"`
	Metrics          *[]Metric          `json:"metrics,omitempty"`
	Modes            *[]Mode            `json:"modes,omitempty"`
	Monitorings      *[]Monitoring      `json:"monitorings,omitempty"`
	Nodes            *[]Node            `json:"nodes,omitempty"`
	Oncalls          *[]Oncall          `json:"oncall,omitempty"`
	Permissions      *[]Permission      `json:"permissions,omitempty"`
	Predicates       *[]Predicate       `json:"predicates,omitempty"`
	Properties       *[]Property        `json:"properties,omitempty"`
	Providers        *[]Provider        `json:"providers,omitempty"`
	Repositories     *[]Repository      `json:"repositories,omitempty"`
	Sections         *[]Section         `json:"sections,omitempty"`
	Servers          *[]Server          `json:"servers,omitempty"`
	States           *[]State           `json:"states,omitempty"`
	Status           *[]Status          `json:"status,omitempty"`
	Systems          *[]System          `json:"system,omitempty"`
	Teams            *[]Team            `json:"teams,omitempty"`
	Tree             *Tree              `json:"tree,omitempty"`
	Units            *[]Unit            `json:"units,omitempty"`
	Users            *[]User            `json:"users,omitempty"`
	Validities       *[]Validity        `json:"validities,omitempty"`
	Views            *[]View            `json:"views,omitempty"`
	Workflows        *[]Workflow        `json:"workflows,omitempty"`
}

func (r *Result) Error(err error) bool {
	if err != nil {
		r.StatusCode = StatusError
		r.StatusText = DisplayStatus[StatusError]
		r.Errors = &[]string{err.Error()}
		return true
	}
	return false
}

func (r *Result) Conflict() {
	r.StatusCode = StatusConflict
	r.StatusText = DisplayStatus[StatusConflict]
}

func (r *Result) NotImplemented() {
	r.StatusCode = StatusNotImplemented
	r.StatusText = DisplayStatus[StatusNotImplemented]
}

func (r *Result) NotFound() {
	r.StatusCode = StatusNotFound
	r.StatusText = DisplayStatus[StatusNotFound]
}

func (r *Result) NotFoundErr(err error) {
	r.StatusCode = StatusNotFound
	r.StatusText = DisplayStatus[StatusNotFound]
	if err != nil {
		r.Errors = &[]string{err.Error()}
	}
}

func (r *Result) Accepted() {
	r.StatusCode = StatusAccepted
	r.StatusText = DisplayStatus[StatusAccepted]
}

func (r *Result) Unavailable() {
	r.StatusCode = StatusUnavailable
	r.StatusText = DisplayStatus[StatusUnavailable]
}

func (r *Result) OK() {
	if r.Errors == nil || *r.Errors == nil || len(*r.Errors) == 0 {
		r.StatusCode = StatusOK
		r.StatusText = DisplayStatus[StatusOK]
		return
	}
	r.Partial()
}

func (r *Result) Partial() {
	r.StatusCode = StatusPartial
	r.StatusText = DisplayStatus[StatusPartial]
}

func (r *Result) Clean() {
	if r.Errors == nil || len(*r.Errors) == 0 {
		r.Errors = nil
	}

	if r.DeploymentsList == nil || len(*r.DeploymentsList) == 0 {
		r.DeploymentsList = nil
	}
}

func (r *Result) BadRequest(err error) {
	r.StatusCode = StatusBadRequest
	r.StatusText = DisplayStatus[StatusBadRequest]
	if err != nil {
		r.Errors = &[]string{err.Error()}
	}
}

func (r *Result) Forbidden(err error) {
	r.StatusCode = StatusForbidden
	r.StatusText = DisplayStatus[StatusForbidden]
	if err != nil {
		r.Errors = &[]string{err.Error()}
	}
}

// NewResult returns a blank proto.Result
func NewResult() Result {
	return Result{
		Errors: &[]string{},
	}
}

// DataClean ensures that r contains no leaky application data except
// for tracking and correlation metadata IDs required to match the
// result with server logs
func (r *Result) DataClean() {
	r.Errors = &[]string{`Internal server error forced empty result`}
	r.DeploymentsList = nil
	r.Actions = nil
	r.Admins = nil
	r.Attributes = nil
	r.Buckets = nil
	r.Capabilities = nil
	r.Categories = nil
	r.CheckConfigs = nil
	r.Clusters = nil
	r.DatacenterGroups = nil
	r.Datacenters = nil
	r.Deployments = nil
	r.Entities = nil
	r.Environments = nil
	r.Grants = nil
	r.Groups = nil
	r.HostDeployments = nil
	r.Instances = nil
	r.JobResults = nil
	r.JobStatus = nil
	r.JobTypes = nil
	r.Jobs = nil
	r.Levels = nil
	r.Metrics = nil
	r.Modes = nil
	r.Monitorings = nil
	r.Nodes = nil
	r.Oncalls = nil
	r.Permissions = nil
	r.Predicates = nil
	r.Properties = nil
	r.Providers = nil
	r.Repositories = nil
	r.Sections = nil
	r.Servers = nil
	r.States = nil
	r.Status = nil
	r.Systems = nil
	r.Teams = nil
	r.Tree = nil
	r.Units = nil
	r.Users = nil
	r.Validities = nil
	r.Views = nil
	r.Workflows = nil
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
