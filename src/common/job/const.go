package job

const (
	// ImageScanJob is name of scan job it will be used as key to register to job service.
	ImageScanJob = "IMAGE_SCAN"
	// ImageScanAllJob is the name of "scanall" job in job service
	ImageScanAllJob = "IMAGE_SCAN_ALL"
	// ImageTransfer : the name of image transfer job in job service
	ImageTransfer = "IMAGE_TRANSFER"
	// ImageDelete : the name of image delete job in job service
	ImageDelete = "IMAGE_DELETE"
	// ImageReplicate : the name of image replicate job in job service
	ImageReplicate = "IMAGE_REPLICATE"
	// ImageGC the name of image garbage collection job in job service
	ImageGC = "IMAGE_GC"
	// ImageWebhook the name of image webhook job in job service
	ImageWebhook = "IMAGE_WEBHOOK"
	// ImageP2PPreheat the name of image p2p preheat job in job service
	ImageP2PPreheat = "IMAGE_P2P_PREHEAT"

	// Job Priority define (1-100000).see https://github.com/gocraft/work#scheduling-algorithm
	JobPriorityHigh = 50000
	JobPriorityNormal = 500
	JobPriorityLow = 5

	// JobKindGeneric : Kind of generic job
	JobKindGeneric = "Generic"
	// JobKindScheduled : Kind of scheduled job
	JobKindScheduled = "Scheduled"
	// JobKindPeriodic : Kind of periodic job
	JobKindPeriodic = "Periodic"

	// JobServiceStatusPending   : job status pending
	JobServiceStatusPending = "Pending"
	// JobServiceStatusRunning   : job status running
	JobServiceStatusRunning = "Running"
	// JobServiceStatusStopped   : job status stopped
	JobServiceStatusStopped = "Stopped"
	// JobServiceStatusCancelled : job status cancelled
	JobServiceStatusCancelled = "Cancelled"
	// JobServiceStatusError     : job status error
	JobServiceStatusError = "Error"
	// JobServiceStatusSuccess   : job status success
	JobServiceStatusSuccess = "Success"
	// JobServiceStatusScheduled : job status scheduled
	JobServiceStatusScheduled = "Scheduled"

	// JobActionStop : the action to stop the job
	JobActionStop = "stop"
)
