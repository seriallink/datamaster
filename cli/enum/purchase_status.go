package enum

type PurchaseStatusEnum string

const (
	PurchasePending   PurchaseStatusEnum = "pending"
	PurchaseConfirmed PurchaseStatusEnum = "confirmed"
	PurchaseShipped   PurchaseStatusEnum = "shipped"
	PurchaseDelivered PurchaseStatusEnum = "delivered"
	PurchaseCancelled PurchaseStatusEnum = "cancelled"
)

func (enum PurchaseStatusEnum) Values() []Enum {
	return []Enum{
		PurchasePending,
		PurchaseConfirmed,
		PurchaseShipped,
		PurchaseDelivered,
		PurchaseCancelled,
	}
}

func (enum PurchaseStatusEnum) String() string {
	return string(enum)
}

func (enum PurchaseStatusEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum PurchaseStatusEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
