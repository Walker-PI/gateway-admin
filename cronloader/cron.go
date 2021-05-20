package cronloader

import (
	"time"

	"github.com/go-co-op/gocron"
)

func InitCronLoader() {

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.TagsUnique()

	// your code ...
	scheduler.Every(10).Seconds().Tag("update_api_expiration").Do(updateAPIConfigExpiration)

	scheduler.StartAsync()
	scheduler.StartBlocking()
}
