package main

import (
	"bytes"
	"context"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type SQLData struct {
	Lock                  int     `db:"lock"`
	ChargingEnable        int     `db:"charging_enable"`
	MaxChargingCurrent    int     `db:"max_charging_current"`
	HaloBrightness        int     `db:"halo_brightness"`
	CumulativeAddedEnergy float64 `db:"cumulative_added_energy"`
	AddedRange            float64 `db:"added_range"`
}

type RedisState struct {
	SessionState   int     `redis:"session.state"`
	ControlPilot   int     `redis:"ctrlPilot"`
	S2open         int     `redis:"S2open"`
	ScheduleEnergy float64 `redis:"scheduleEnergy"`
}

type RedisM2W struct {
	ChargerStatus int     `redis:"tms.charger_status"`
	Line1Power    float64 `redis:"tms.line1.power_watt.value"`
	Line2Power    float64 `redis:"tms.line2.power_watt.value"`
	Line3Power    float64 `redis:"tms.line3.power_watt.value"`
}

type DataCache struct {
	SQL        SQLData
	RedisState RedisState
	RedisM2W   RedisM2W
}

type Wallbox struct {
	redisClient *redis.Client
	sqlClient   *sqlx.DB
	Data        DataCache
}

func NewWallbox() *Wallbox {
	var w Wallbox

	var err error
	w.sqlClient, err = sqlx.Connect("mysql", "root:fJmExsJgmKV7cq8H@tcp(127.0.0.1:3306)/wallbox")
	if err != nil {
		panic(err)
	}

	w.redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &w
}

func (w *Wallbox) UpdateCache() {
	ctx := context.Background()

	stateRes := w.redisClient.HGetAll(ctx, "state")
	if stateRes.Err() != nil {
		panic(stateRes.Err())
	}

	if err := stateRes.Scan(&w.Data.RedisState); err != nil {
		panic(err)
	}

	m2wRes := w.redisClient.HGetAll(ctx, "m2w")
	if m2wRes.Err() != nil {
		panic(m2wRes.Err())
	}

	if err := m2wRes.Scan(&w.Data.RedisM2W); err != nil {
		panic(err)
	}

	query := "SELECT " +
		"  `wallbox_config`.`charging_enable`," +
		"  `wallbox_config`.`lock`," +
		"  `wallbox_config`.`max_charging_current`," +
		"  `wallbox_config`.`halo_brightness`," +
		"  `power_outage_values`.`charged_energy` AS cumulative_added_energy," +
		"  IF(`active_session`.`unique_id` != 0," +
		"    `active_session`.`charged_range`," +
		"    `latest_session`.`charged_range`) AS added_range " +
		"FROM `wallbox_config`," +
		"    `active_session`," +
		"    `power_outage_values`," +
		"    (SELECT * FROM `session` ORDER BY `id` DESC LIMIT 1) AS latest_session"
	w.sqlClient.Get(&w.Data.SQL, query)
}

func (w *Wallbox) GetSerialNumber() string {
	var serialNumber string
	w.sqlClient.Get(&serialNumber, "SELECT `serial_num` FROM charger_info")
	return serialNumber
}

func (w *Wallbox) GetUserId() string {
	var userId string
	w.sqlClient.QueryRow("SELECT `user_id` FROM `users` WHERE `user_id` != 1 ORDER BY `user_id` DESC LIMIT 1").Scan(&userId)
	return userId
}

func (w *Wallbox) GetAvailableCurrent() string {
	var availableCurrent string
	w.sqlClient.QueryRow("SELECT `max_avbl_current` FROM `state_values` ORDER BY `id` DESC LIMIT 1").Scan(&availableCurrent)
	return availableCurrent
}

func SendToPosixQueue(path, data string) {
	pathBytes := append([]byte(path), 0)
	mq := mq_open(pathBytes)

	event := []byte(data)
	eventPaddedBytes := append(event, bytes.Repeat([]byte{0x00}, 1024-len(event))...)

	mq_timedsend(mq, eventPaddedBytes)
	mq_close(mq)
}

func (w *Wallbox) SetLocked(lockStr string) {
	lock, _ := strconv.Atoi(lockStr)
	w.UpdateCache()
	if lock == w.Data.SQL.Lock {
		return
	}
	if lock == 1 {
		SendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_LOGIN", "EVENT_REQUEST_LOCK")
	} else {
		userId := w.GetUserId()
		SendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_LOGIN", "EVENT_REQUEST_LOGIN#"+userId+".000000")
	}
}

func (w *Wallbox) SetChargingEnable(enableStr string) {
	enable, _ := strconv.Atoi(enableStr)
	w.UpdateCache()
	if enable == w.Data.SQL.ChargingEnable {
		return
	}
	if enable == 1 {
		SendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_STATEMACHINE", "EVENT_REQUEST_USER_ACTION#1.000000")
	} else {
		SendToPosixQueue("WALLBOX_MYWALLBOX_WALLBOX_STATEMACHINE", "EVENT_REQUEST_USER_ACTION#2.000000")
	}
}

func (w *Wallbox) SetMaxChargingCurrent(current string) {
	w.sqlClient.MustExec("UPDATE `wallbox_config` SET `max_charging_current`=?", current)
}

func (w *Wallbox) SetHaloBrightness(brightness string) {
	w.sqlClient.MustExec("UPDATE `wallbox_config` SET `halo_brightness`=?", brightness)
}

func (w *Wallbox) GetCableConnected() interface{} {
	if w.Data.RedisM2W.ChargerStatus == 0 || w.Data.RedisM2W.ChargerStatus == 6 {
		return 0
	}
	return 1
}

func (w *Wallbox) GetEffectiveStatus() interface{} {
	tmsStatus := w.Data.RedisM2W.ChargerStatus
	state := w.Data.RedisState.SessionState

	// The wallbox app shows "Locked" for longer than the TMS status
	if state == 210 { // Wait unlock
		tmsStatus = 6 // Locked
	}

	return wallboxStatusCodes[tmsStatus]
}
