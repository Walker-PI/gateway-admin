package cronloader

import (
	"time"

	"github.com/go-co-op/gocron"
)

func InitCronLoader() {

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.TagsUnique()

	// your code ...
	scheduler.Every(10).Seconds().Tag("update_route_expiration").Do(updateRouteConfigExpiration)

	scheduler.StartAsync()
	scheduler.StartBlocking()
}
