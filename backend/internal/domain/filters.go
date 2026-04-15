package domain

type FeedType string

const (
	FeedTypeRecent  FeedType = "recent"
	FeedTypePopular FeedType = "popular"
	FeedTypeShelved FeedType = "shelved"
	FeedTypeMargin  FeedType = "margin"
	FeedTypeSemble  FeedType = "semble"
)

type NoteFilter struct {
	Motivations []string
	AuthorDID   string
	TargetHash  string
	Tag         string
	FeedType    FeedType
	Query       string
	Limit       int
	Offset      int
}
