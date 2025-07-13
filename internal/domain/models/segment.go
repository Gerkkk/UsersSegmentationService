package models

// Segment - Структура для сегмента, на которые делятся пользователи
type Segment struct {
	Id          string `json:"id"`
	Description string `json:"description"`
}
