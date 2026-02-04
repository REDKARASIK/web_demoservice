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
		OrderUID:          order.Order.ID.String(),
		TrackNumber:       order.Order.TrackNumber,
		Entry:             order.Order.Entry,
		Locale:            order.Order.Locale,
		InternalSignature: getValue(order.Order.InternalSignature),
		CustomerID:        order.Order.CustomerID,
		DeliveryService:   getValue(order.Order.DeliveryService),
		ShardKey:          order.Order.ShardKey,
		SmID:              getValue(order.Order.SmID),
		DateCreated:       order.Order.DateCreated,
		OofShard:          order.Order.OofShard,
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
