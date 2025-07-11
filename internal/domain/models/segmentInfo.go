package models

type SegmentInfo struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	UsersNum    int64  `json:"users_num"`
}
