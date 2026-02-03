package validator

// Validator — валидация входящих данных (в т.ч. через внешние gRPC-сервисы).
// Добавьте методы по мере необходимости, например:
//   ValidateRegisterRequest(req *dto.RegisterRequest) error
//   ValidateCreateSessionRequest(req *dto.CreateSessionRequest) error
type Validator struct{}

func New() *Validator {
	return &Validator{}
}
