package enum

type TrackingStatusEnum string

const (
	TrackingCreated        TrackingStatusEnum = "created"
	TrackingPickedUp       TrackingStatusEnum = "picked_up"
	TrackingInTransit      TrackingStatusEnum = "in_transit"
	TrackingDelivered      TrackingStatusEnum = "delivered"
	TrackingDelayed        TrackingStatusEnum = "delayed"
	TrackingFailedDelivery TrackingStatusEnum = "failed_delivery"
	TrackingReturned       TrackingStatusEnum = "returned"
	TrackingLost           TrackingStatusEnum = "lost"
	TrackingDamaged        TrackingStatusEnum = "damaged"
)

func (enum TrackingStatusEnum) Values() []Enum {
	return []Enum{
		TrackingCreated,
		TrackingPickedUp,
		TrackingInTransit,
		TrackingDelivered,
		TrackingDelayed,
		TrackingFailedDelivery,
		TrackingReturned,
		TrackingLost,
		TrackingDamaged,
	}
}

func (enum TrackingStatusEnum) String() string {
	return string(enum)
}

func (enum TrackingStatusEnum) IsEmpty() bool {
	return isEmptyEnum(enum)
}

func (enum TrackingStatusEnum) IsValid() bool {
	return isValidEnum(enum, enum)
}
