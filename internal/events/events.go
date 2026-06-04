package events

const (
	ExchangeName      = "carshop.exchange"
	QueueName         = "carshop.notifications"
	BindingRoutingKey = "car.#"
	CarCreatedKey     = "car.created"
	CarUpdatedKey     = "car.updated"
	CarDeletedKey     = "car.deleted"
)

type CarCreatedEvent struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	Colour    string  `json:"colour"`
	Price     float64 `json:"price"`
	BuildDate string  `json:"build_date"`
}

type CarUpdatedEvent struct {
	ID        uint64   `json:"id"`
	Name      *string  `json:"name,omitempty"`
	Colour    *string  `json:"colour,omitempty"`
	Price     *float64 `json:"price,omitempty"`
	BuildDate *string  `json:"build_date,omitempty"`
}

type CarDeletedEvent struct {
	ID        uint64 `json:"id"`
	DeletedAt string `json:"deleted_at"`
}
