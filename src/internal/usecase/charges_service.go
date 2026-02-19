package usecase

type ChargesService struct {
	chargePercent float64
	vatPercent    float64
}

func NewChargesService(chargePercent float64, vatPercent float64) *ChargesService {
	return &ChargesService{
		chargePercent: chargePercent,
		vatPercent:    vatPercent,
	}
}
