package gateway

const (
	EventServiceGet = iota
	EventServiceAdd
	EventServiceUpdate
	EventServiceDelete
)

type Event struct {
	EventType int
	TimeStamp string
	Content   interface{}
}
