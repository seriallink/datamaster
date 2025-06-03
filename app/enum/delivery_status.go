package enum

type DeliveryStatusEnum string

const (
	DeliveryPending   DeliveryStatusEnum = "pending"
	DeliveryShipped   DeliveryStatusEnum = "shipped"
	DeliveryDelivered DeliveryStatusEnum = "delivered"
	DeliveryFailed    DeliveryStatusEnum = "failed"
)

func (enum DeliveryStatusEnum) Values() []Enum {
	return []Enum{
		DeliveryPending,
		DeliveryShipped,
		DeliveryDelivered,
		DeliveryFailed,
	}
}

func (enum DeliveryStatusEnum) String() string {
	return string(enum)
}

func (enum DeliveryStatusEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum DeliveryStatusEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
