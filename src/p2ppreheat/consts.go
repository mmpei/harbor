package p2ppreheat

const (
	// ImagePushEvent : webhook type for pushing image
	ImagePushEvent = "pushImage"
	// UploadChartEvent : webhook type for uploading chart
	UploadChartEvent = "uploadChart"
	// ReplicateFinishedEvent : webhook type for replicating finished
	ReplicateFinishedEvent = "replicateFinished"
	// ScanFinishedEvent : webhook type for scanning images finished
	ScanFinishedEvent = "scanFinished"
)

var HookTypes = []string{ImagePushEvent, UploadChartEvent, ReplicateFinishedEvent, ScanFinishedEvent}