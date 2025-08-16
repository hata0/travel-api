package revokedtoken

type RevokedTokenID struct {
	value string
}

func NewRevokedTokenID(id string) RevokedTokenID {
	return RevokedTokenID{value: id}
}

func (id RevokedTokenID) String() string {
	return id.value
}

func (id RevokedTokenID) Equals(other RevokedTokenID) bool {
	return id.value == other.value
}
