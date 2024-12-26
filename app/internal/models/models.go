package models

// структура напоминания
type Remind struct {
	ID      string `json:"id"` // классно обманули с типом id :) поставил uint64 в итоге put падал из-за невозможности декода
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
