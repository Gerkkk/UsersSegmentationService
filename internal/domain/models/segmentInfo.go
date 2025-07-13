package models

// SegmentInfo - Структура для статистики сегмента, которая получается при запросе GetSegmentInfo
type SegmentInfo struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	UsersNum    int64  `json:"users_num"`
}
