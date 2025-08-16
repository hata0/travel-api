package refreshtoken

type RefreshTokenID struct {
	value string
}

func NewRefreshTokenID(id string) RefreshTokenID {
	return RefreshTokenID{value: id}
}

func (id RefreshTokenID) String() string {
	return id.value
}

func (id RefreshTokenID) Equals(other RefreshTokenID) bool {
	return id.value == other.value
}
