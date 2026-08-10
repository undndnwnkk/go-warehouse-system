package transport

type OrderItem struct {
	OrderId  string
	Quantity int64
}

type ReserveStockRequest struct {
	OrderId  string
	Repeated OrderItem
}

type ReserveStockResponse struct {
	Success bool
	Message string
}

type WarehouseServer interface {
	ReserveStock(request ReserveStockRequest) ReserveStockResponse
}
