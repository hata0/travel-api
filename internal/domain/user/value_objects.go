package user

type UserID struct {
	value string
}

func NewUserID(id string) UserID {
	return UserID{value: id}
}

func (id UserID) String() string {
	return id.value
}

func (id UserID) Equals(other UserID) bool {
	return id.value == other.value
}
