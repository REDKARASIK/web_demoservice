package dto

import "web_demoservice/internal/domain"

func MapToOrderDTO(order *domain.OrderWithInformation) OrderWithInformationDTO {
	// Маппинг товаров
	itemsDTO := make([]ItemDTO, 0, len(order.Items))
	for _, item := range order.Items {
		itemsDTO = append(itemsDTO, ItemDTO{
			ChrtID:      getValue(item.ChrtID),
			TrackNumber: item.TrackNumber,
			Price:       item.Price,
			RID:         item.RID,
			Name:        item.Name,
			Sale:        getValue(item.Sale),
			Size:        getValue(item.Size),
			TotalPrice:  item.TotalPrice,
			NmID:        item.NmID,
			Brand:       item.Brand,
			Status:      item.Status,
		})
	}

	return OrderWithInformationDTO{
		OrderUID:          order.ID.String(),
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Locale:            order.Locale,
		InternalSignature: getValue(order.InternalSignature),
		CustomerID:        order.CustomerID,
		DeliveryService:   getValue(order.DeliveryService),
		ShardKey:          order.ShardKey,
		SmID:              getValue(order.SmID),
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
		Delivery: DeliveryDTO{
			Name:    order.Delivery.Name,
			Phone:   order.Delivery.Phone,
			Zip:     order.Delivery.Zip,
			City:    order.Delivery.City,
			Address: order.Delivery.Address,
			Region:  getValue(order.Delivery.Region),
			Email:   order.Delivery.Email,
		},
		Payment: PaymentDTO{
			Transaction:  order.Payment.Transaction,
			RequestID:    getValue(order.Payment.RequestID),
			Currency:     order.Payment.Currency,
			Provider:     order.Payment.Provider,
			Amount:       order.Payment.Amount,
			PaymentDt:    order.Payment.PaymentDt,
			Bank:         order.Payment.Bank.Name,
			DeliveryCost: order.Payment.DeliveryCost,
			GoodsTotal:   order.Payment.GoodsTotal,
			CustomFee:    order.Payment.CustomFee,
		},
		Items: itemsDTO,
	}
}

// Вспомогательная функция для безопасного получения значений из указателей
func getValue[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}
