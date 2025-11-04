package gcrontask

import "cloud_store/cron"

var (
	CleanTask     *cron.Consumers
	OSSUploadTask *cron.Consumers
)
