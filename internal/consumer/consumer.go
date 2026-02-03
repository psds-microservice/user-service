package consumer

// Consumer — обработчики очередей RabbitMQ.
// Добавьте обработчики по мере подключения очередей, например:
//   RegisterHandlers(conn *amqp.Connection, handlers map[string]Handler)
//   HandleUserEvents(ctx context.Context, msgs <-chan amqp.Delivery)
type Consumer struct{}

func New() *Consumer {
	return &Consumer{}
}
