package sign_model

import "time"

type SignRecordJSON struct {
	Time  time.Time            `json:"time"`
	Roles map[string]RolesJSON `json:"roles"`
}

type RolesJSON struct {
	UID          string    `json:"uid"`
	Name         string    `json:"name"`
	Time         time.Time `json:"time"`
	TotalSignDay int8      `json:"total_sign_day"`
}
