package trip

// TripID は旅行IDを表現する値オブジェクト
type TripID struct {
	value string
}

func NewTripID(id string) TripID {
	return TripID{value: id}
}

func (id TripID) String() string {
	return id.value
}

func (id TripID) Equals(other TripID) bool {
	return id.value == other.value
}
