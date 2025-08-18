package utils

import "marketflow/internal/domain"

func CheckExchangeName(exchange string) error {
	for _, val := range domain.Exchanges {
		if exchange == val {
			return nil
		}
	}
	return domain.ErrInvalidExchangeVal
}

func CheckSymbolName(symbol string) error {
	for _, val := range domain.Symbols {
		if symbol == val {
			return nil
		}
	}
	return domain.ErrInvalidSymbolVal
}
