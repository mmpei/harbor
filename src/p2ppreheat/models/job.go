package models

import (
	"time"
)

type JobData struct {
	EventType    string            `json:"event_type"`
	Events       []interface{}    `json:"events"`
}

type PushEvent struct {
	Project      string     `json:"project"`
	RepoName     string     `json:"repo_name"`
	Tag          string     `json:"tag"`
	FullName     string     `json:"full_name"`
	TriggerTime  time.Time  `json:"trigger_time"`
	Digest       string     `json:"digest"`
	ProjectType  string     `json:"project_type"`
	ResourceURL  string     `json:"resource_url"`
}
