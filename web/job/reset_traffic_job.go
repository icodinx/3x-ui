package job

import (
	"github.com/goccy/go-json"
	"x-ui/database"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/web/service"
)

type ResetTrafficJob struct {
	xrayService     service.XrayService
	inboundService  service.InboundService
	outboundService service.OutboundService
}

func NewResetTrafficJob() *ResetTrafficJob {
	return new(ResetTrafficJob)
}

func (j *ResetTrafficJob) Run() {
	if !j.xrayService.IsXrayRunning() {
		return
	}

	logger.Info("do reset job")

	needRestart, err := j.resetInboundTraffic()
	if err != nil {
		logger.Warning("reset inbound traffic failed:", err)
		return
	}

	logger.Info("Traffic has been reset", needRestart, err)
	if needRestart {
		j.xrayService.SetToNeedRestart()
	}
}

func (j *ResetTrafficJob) resetInboundTraffic() (bool, error) {

	db := database.GetDB()
	var inbounds []*model.Inbound
	err := db.Model(model.Inbound{}).
		//Where("settings like '%reset%'").
		Where("1=1").
		Find(&inbounds).
		Error
	if err != nil {
		return false, err
	}

	needRestart := false

	for _, inbound := range inbounds {

		logger.Info("Inbound {}", inbound)

		setting := new(map[string]interface{})
		err = json.Unmarshal([]byte(inbound.Settings), &setting)
		if nil == err {

			//resetType := setting.AutoResetInfo.AutoResetTraffic.ResetType
			//minute := time.Now().Minute()
			//hour := time.Now().Hour()
			//day := time.Now().Day()

			reset := true

			//reset := resetType == model.Hourly && 0 == minute
			//reset = reset || resetType == model.Daily && 0 == hour && 0 == minute
			//reset = reset || resetType == model.Monthly && 0 == day && 0 == hour && 0 == minute

			if reset {
				err0 := j.inboundService.ResetInboundTraffics(inbound.Id)
				if err0 != nil {
					err = err0
				} else {
					needRestart = true
				}
			}
		} else {
			logger.Error("inbound settings:", err)
		}
	}

	return needRestart, err
}
